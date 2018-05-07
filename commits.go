package main

import (
	"regexp"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func getStartCommit(head plumbing.Reference, from string) string {
	if from == "" {
		return head.Hash().String()
	}

	return from
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

func getDiff(currentHash string, repo git.Repository) object.Changes {
	commitObject, _ := repo.CommitObject(plumbing.NewHash(currentHash))
	currentDirState, _ := commitObject.Tree()

	prevCommitObject, _ := commitObject.Parents().Next()
	prevDirState, _ := prevCommitObject.Tree()

	changes, _ := prevDirState.Diff(currentDirState)

	return changes
}

func matchAny(credentialPatterns []regexp.Regexp, text string) bool {
	for index := range credentialPatterns {
		if credentialPatterns[index].MatchString(text) {
			return true
		}
	}

	return false
}
