package main

import (
	"strings"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"os"
	"regexp"
	"flag"
	"fmt"
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

func isNotInitialCommit(commit object.Commit) bool {
	return commit.NumParents() >= 1
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

func getAllHashesButInitial(commits object.CommitIter) []string {
	currentCommit, _ := commits.Next()

	_, hashes := takeCommitsUtil(commits, func(commit object.Commit) bool {
		return isNotInitialCommit(*currentCommit)
	})

	return hashes
}

func getAllHashesUntil(commits object.CommitIter, until string) []string {
	currentCommit, hashes := takeCommitsUtil(commits, func(commit object.Commit) bool {
		return ! strings.HasPrefix(commit.Hash.String(), until)
	})

	hashes = append(hashes, currentCommit.Hash.String())

	return hashes
}

func hashesToInspect(repository git.Repository, from, to string) []string {
	log, _ := repository.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})

	currentCommit, _ := log.Next()
	for ! strings.HasPrefix(currentCommit.Hash.String(), from) {
		currentCommit, _ = log.Next()
	}

	if to == "" {
		return getAllHashesButInitial(log)
	}

	return getAllHashesUntil(log, to)
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

func ensureRepositoryExists(path string) {
	fileExists, err := exists(path)

	if err != nil {
		panic(err)
	} else if ! fileExists {
		panic(fmt.Sprintf("Path given is not a valid directory: %s", path))
	}
}

func getStartCommit(head plumbing.Reference, from string) string {
	if from == "" {
		return head.Hash().String()
	}

	return from
}

func main() {

	JWT_REGEX := regexp.MustCompile(`eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]*`)
	var dirtyCommits []string

	repoPath := flag.String("path", "", "path to a local git project")
	from := flag.String("from", "", "start commit")
	to := flag.String("to", "", "final commit")
	flag.Parse()

	ensureRepositoryExists(*repoPath)

	repo, _ := git.PlainOpen(*repoPath)
	head, _ := repo.Head()

	startCommit := getStartCommit(*head, *from)

	commits := hashesToInspect(*repo, startCommit, *to)

	for commitIndex := 0; commitIndex < len(commits); commitIndex++ {
		currentCommit := commits[commitIndex]
		changes := getDiff(currentCommit, *repo)

		for changeIndex := 0; changeIndex < changes.Len(); changeIndex++ {
			patch, _ := changes[changeIndex].Patch()

			if JWT_REGEX.MatchString(patch.String()) {
				dirtyCommits = append(dirtyCommits, currentCommit)
			}
		}
	}

	fmt.Println(dirtyCommits)
}
