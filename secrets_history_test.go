package main

import (
	"testing"
	"gopkg.in/src-d/go-git.v4"
)

func TestWhenStartCommitIsEmptyReturnHeadHash(t *testing.T) {
	repo, _ := git.PlainOpen(".")
	head, _ := repo.Head()
	initialCommit := getStartCommit(*head, "")
	if initialCommit != head.Hash().String() {
		t.Errorf("Wrong initial commit, got: %s, want: %s.", initialCommit, head.Hash().String())
	}
}

func TestWhenStartCommitIsNotEmptyReturnIt(t *testing.T) {
	repo, _ := git.PlainOpen(".")
	head, _ := repo.Head()
	initialCommit := getStartCommit(*head, "bebecaf")
	if initialCommit != "bebecaf" {
		t.Errorf("Wrong initial commit, got: %s, want: %s.", initialCommit, "bebecaf")
	}
}

func TestHashesToInspectContainsInitialCommit(t *testing.T) {
	repo, _ := git.PlainOpen(".")
	from := "7e6a7f39324f8feb1c24562fb70650ead4e42604"

	hashes, _ := hashesToInspect(*repo, from, "")

	if hashes[0] != "7e6a7f39324f8feb1c24562fb70650ead4e42604" {
		t.Errorf("First hash does not match initial commit, got: %s, want: %s.", hashes[0], from)
	}
}

func TestHashesToInspectFailsWhenInitialCommitDoesNotExist(t *testing.T) {
	repo, _ := git.PlainOpen(".")
	from := "bebecaf-"

	_, err := hashesToInspect(*repo, from, "")

	if err == nil {
		t.Errorf("Didn't got error when commit was innexistent")
	}
}

func TestHashesToInspectFailsWhenFromIsEqualToTo(t *testing.T) {
	repo, _ := git.PlainOpen(".")
	from := "7e6a7f39324f8feb1c24562fb70650ead4e42604"
	to := "7e6a7f39324f8feb1c24562fb70650ead4e42604"

	_, err := hashesToInspect(*repo, from, to)

	if err == nil {
		t.Errorf("Didn't got error when `from` and `to` had the same value")
	}
}

func TestHashesToInspectFailsWhenFromIsEqualToInitialCommit(t *testing.T) {
	repo, _ := git.PlainOpen(".")
	from := "15647bb51242323335dc7ef1d841eb0d223335c8"

	_, err := hashesToInspect(*repo, from, "")

	if err == nil {
		t.Errorf("Didn't got error when `from` and `to` had the same value")
	}
}