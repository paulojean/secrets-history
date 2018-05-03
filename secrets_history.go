package main

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"os"
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

func getAllHashesButInitial(commits object.CommitIter) []string {
	var hashes []string
	currentCommit, _ := commits.Next()
	for isNotInitialCommit(*currentCommit) {
		if isNotMergeCommit(*currentCommit) {
			hashes = append(hashes, currentCommit.Hash.String())
		}
		currentCommit, _ = commits.Next()
	}

	return hashes
}

func getAllHashesUntil(commits object.CommitIter, until string) []string {
	var hashes []string

	currentCommit, _ := commits.Next()
	for ! strings.HasPrefix(currentCommit.Hash.String(), until) {
		if isNotMergeCommit(*currentCommit) {
			hashes = append(hashes, currentCommit.Hash.String())
		}
		currentCommit, _ = commits.Next()
	}

	hashes = append(hashes, currentCommit.Hash.String())

	return hashes
}

func hashesToInspect(repository git.Repository, from, to string) []string {
	var hashes []string
	ref, _ := repository.Head()
	c, _ := repository.CommitObject(ref.Hash())

	commits := object.NewCommitPostorderIter(c, nil)

	currentCommit, _ := commits.Next()
	for ! strings.HasPrefix(currentCommit.Hash.String(), from) {
		currentCommit, _ = commits.Next()
	}

	if to == "" {
		hashes = getAllHashesButInitial(commits)
	} else {
		hashes = getAllHashesUntil(commits, to)
	}

	return hashes
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func main() {
	JWT_REGEX := regexp.MustCompile(`^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]*$`)
	var dirtyCommits []string

	repoExists, err := exists(os.Args[1])
	if ! repoExists {
		panic("path given does not exist")
	} else if err != nil {
		panic(err)
	}

	repo, _ := git.PlainOpen(os.Args[1])

	head, _ := repo.Head()

	var from, to string
	flag.StringVar(&from, "from", head.Hash().String(), "start commit")
	flag.StringVar(&to, "to", "", "final commit")
	flag.Parse()

	hashes := hashesToInspect(*repo, from, to)

	for i := 0; i < len(hashes); i++ {
		currentCommit := hashes[i]
		changes := getDiff(currentCommit, *repo)
		for j := 0; j < changes.Len(); j++ {
			patch, _ := changes[j].Patch()

			if JWT_REGEX.MatchString(patch.String()) {
				dirtyCommits = append(dirtyCommits, currentCommit)
			}
		}
	}

	fmt.Println(dirtyCommits)
}
