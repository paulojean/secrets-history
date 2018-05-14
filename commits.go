package main

import (
	"regexp"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"sync"
	"strings"
)

func getStartCommit(head plumbing.Reference, from string) string {
	if from == "" {
		return head.Hash().String()
	}

	return from
}

func getDirtyCommits(repo git.Repository, commits []string, credentialPatterns []regexp.Regexp) []string {
	var dirtyCommits = make(map[string]interface{}, len(commits))
	var dirtyCommitsSimple []string
	var changesWg sync.WaitGroup

	for commitIndex := 0; commitIndex < len(commits); commitIndex++ {
		changesWg.Add(1)

		commit := commits[commitIndex]
		changes := getDiff(commit, repo)
		var wg sync.WaitGroup

		go func() {
			for changeIndex := 0; changeIndex < changes.Len(); changeIndex++ {
				patch, _ := changes[changeIndex].Patch()
				go checkPatch(wg, dirtyCommits, credentialPatterns, patch, commit)
			}

			changesWg.Done()
		}()

		wg.Wait()
	}

	changesWg.Wait()

	for commit := range dirtyCommits {
		dirtyCommitsSimple = append(dirtyCommitsSimple, commit)
	}

	return dirtyCommitsSimple
}

func checkPatch(wg sync.WaitGroup, dirtyCommits map[string]interface{}, credentialPatterns []regexp.Regexp, patch *object.Patch, currentCommit string) {
	wg.Add(1)

	additionsOnlyExpression := regexp.MustCompile(`(?m)^\+(.*)$`)

	text := additionsOnlyExpression.FindAllString(patch.String(), -1)
	additionsText := strings.Join(text, "\n")

	if matchAny(credentialPatterns, additionsText) {
		dirtyCommits[currentCommit] = currentCommit
	}

	wg.Done()
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
