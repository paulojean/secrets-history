package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
	"encoding/json"
)

func getDiff(currentHash string, repo git.Repository) object.Changes {
	commitObject, _ := repo.CommitObject(plumbing.NewHash(currentHash))
	currentDirState, _ := commitObject.Tree()

	prevCommitObject, _ := commitObject.Parents().Next()
	prevDirState, _ := prevCommitObject.Tree()

	changes, _ := prevDirState.Diff(currentDirState)

	return changes
}

func isNotMergeCommit(commit object.Commit) bool {
	return commit.NumParents() <= 1
}

func isInitialCommit(commit object.Commit) bool {
	return commit.NumParents() < 1
}

func takeCommitsUtil(commits object.CommitIter, takeUntilFn func(object.Commit) bool) (*object.Commit, []string) {
	var hashes []string
	currentCommit, _ := commits.Next()
	for takeUntilFn(*currentCommit) {
		if isNotMergeCommit(*currentCommit) {
			hashes = append(hashes, currentCommit.Hash.String())
		}
		currentCommit, _ = commits.Next()
	}
	return currentCommit, hashes
}

func getAllHashesButInitial(initialCommit object.Commit, commits object.CommitIter) []string {
	start := []string{initialCommit.Hash.String()}
	_, hashes := takeCommitsUtil(commits, func(commit object.Commit) bool {
		return ! isInitialCommit(commit)
	})

	return append(start, hashes...)
}

func getAllHashesUntil(initialCommit object.Commit, commits object.CommitIter, until string) []string {
	start := []string{initialCommit.Hash.String()}
	currentCommit, hashes := takeCommitsUtil(commits, func(commit object.Commit) bool {
		return ! strings.HasPrefix(commit.Hash.String(), until)
	})

	hashes = append(hashes, currentCommit.Hash.String())

	return append(start, hashes...)
}

func hashesToInspect(repository git.Repository, from, to string) ([]string, error) {
	log, _ := repository.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	var err error

	startCommit, _ := log.Next()
	for ! strings.HasPrefix(startCommit.Hash.String(), from) {
		startCommit, err = log.Next()

		if err != nil {
			return nil, err
		}
	}

	if startCommit.Hash.String() == to {
		return nil, errors.New("`from` must not be equal to `to`")
	} else if isInitialCommit(*startCommit) {
		return nil, errors.New("`from` must not be the repository's initial commit")
	}

	if to == "" {
		return getAllHashesButInitial(*startCommit, log), nil
	}

	return getAllHashesUntil(*startCommit, log, to), nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func repositoryExists(path string) error {
	fileExists, err := exists(path)

	if err != nil {
		return err
	} else if ! fileExists {
		return errors.New(fmt.Sprintf("Path given is not a valid directory: %s", path))
	}

	return nil
}

func getStartCommit(head plumbing.Reference, from string) string {
	if from == "" {
		return head.Hash().String()
	}

	return from
}

func matchAny(credentialPatterns []regexp.Regexp, text string) bool {
	for index := range credentialPatterns {
		if credentialPatterns[index].MatchString(text) {
			return true
		}
	}

	return false
}

func getDirtyCommits(repo git.Repository, commits []string, credentialPatterns []regexp.Regexp) []string {
	var dirtyCommits []string
	for commitIndex := 0; commitIndex < len(commits); commitIndex++ {
		currentCommit := commits[commitIndex]
		changes := getDiff(currentCommit, repo)

		for changeIndex := 0; changeIndex < changes.Len(); changeIndex++ {
			patch, _ := changes[changeIndex].Patch()

			if matchAny(credentialPatterns, patch.String()) {
				dirtyCommits = append(dirtyCommits, currentCommit)
			}
		}
	}
	return dirtyCommits
}

type SecurityCredential struct {
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
}

func parsePatternFile(path string) ([]SecurityCredential, error) {
	patterns, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var credentials []SecurityCredential
	json.Unmarshal(patterns, &credentials)
	return credentials, nil
}

func securityCredentialsToRegex(securityCredentials []SecurityCredential) []regexp.Regexp {
	patterns := make([]regexp.Regexp, len(securityCredentials))
	for i, v := range securityCredentials {
		patterns[i] = *regexp.MustCompile(v.Pattern)
	}
	return patterns
}

func getCredentialPatterns(securityPatternPath string, useDefault bool) ([]regexp.Regexp, error) {
	var credentialsPatterns, defaultPatterns []SecurityCredential
	var err error
	if securityPatternPath != "" {
		credentialsPatterns, err = parsePatternFile(securityPatternPath)
	}

	if err != nil {
		return nil, err
	}

	if useDefault {
		defaultPatterns, err = parsePatternFile("./default_patterns.json")
	}

	if err != nil {
		return nil, err
	}

	patterns := securityCredentialsToRegex(append(credentialsPatterns, defaultPatterns...))

	return patterns, nil
}

func main() {
	var err error
	repoPath := flag.String("path", "", "path to a local git project")
	from := flag.String("from", "", "start commit")
	useDefaultPatterns := flag.Bool("default-patterns", true, "use default pattern credentials")
	patternsDirectory := flag.String("credential-patterns", "", "json file to use custom patterns on search")
	to := flag.String("to", "", "final commit")
	flag.Parse()

	if ! *useDefaultPatterns && *patternsDirectory == "" {
		panic("either use default patterns or provide custom ones")
	}

	securityCredentialPatterns, err := getCredentialPatterns(*patternsDirectory, *useDefaultPatterns)
	if err != nil {
		panic(err.Error())
	}

	repoAbsolutePath, _ := filepath.Abs(*repoPath)

	err = repositoryExists(repoAbsolutePath)
	if err != nil {
		panic(err.Error())
	}

	repo, _ := git.PlainOpen(repoAbsolutePath)
	head, _ := repo.Head()

	startCommit := getStartCommit(*head, *from)

	commits, err := hashesToInspect(*repo, startCommit, *to)

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println(getDirtyCommits(*repo, commits, securityCredentialPatterns))
	}
}
