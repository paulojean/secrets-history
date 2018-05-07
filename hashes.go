package main

import (
	"strings"
	"errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

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

func isNotMergeCommit(commit object.Commit) bool {
	return commit.NumParents() <= 1
}

func isInitialCommit(commit object.Commit) bool {
	return commit.NumParents() < 1
}
