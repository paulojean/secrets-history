package main

import (
	"testing"
	"gopkg.in/src-d/go-git.v4"
	"regexp"
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

func TestHashesToInspectContainsAllCommitsOnRange(t *testing.T) {
	expectedHashes := []string{
		"7e6a7f39324f8feb1c24562fb70650ead4e42604",
		"62e99d7b4d6e628f175f0650275c2749babcc599",
		"fc9ced076e0574a3a7c1ae47981b27c2c9f8aff1",
		"3a58c45d0d1aa6ab8c87df5e2114ee0012fd46c0",
		"87f8d724575b7d7caaef81b7ddb6cf9a31ddb47c",
		"b8ee84c44ca4035f578eb3c20d304c1a25cd9b6e",
		"80e33de42b562df4f3abb5a3c340409b9417471a",
		"d644566ed7b990cff8b5a480bfdcc37bfa911916"}

	repo, _ := git.PlainOpen(".")
	from := "7e6a7f39324f8feb1c24562fb70650ead4e42604"

	hashes, _ := hashesToInspect(*repo, from, "")

	hashesMatches := testEq(expectedHashes, hashes)
	if !hashesMatches {
		t.Errorf("Hashes returned do not match the expecteds")
	}
}

func TestHashesToInspectDoesNotContainsMergeCommit(t *testing.T) {
	mergeCommit := "96d52e2ec790afc40643c1e73899ac95bd9ab299"

	repo, _ := git.PlainOpen(".")
	from := "f2a5f5142316cc3f7b07ed235dbe07d378c5f2b4"

	hashes, _ := hashesToInspect(*repo, from, "")

	for _, hash := range hashes {
		if hash == mergeCommit {
			t.Errorf("Range contains merge commit")
		}
	}
}

func TestHashesToInspectContainsAllCommitsOnRangeWhenSpecifyLowerBound(t *testing.T) {
	expectedHashes := []string{
		"7e6a7f39324f8feb1c24562fb70650ead4e42604",
		"62e99d7b4d6e628f175f0650275c2749babcc599",
		"fc9ced076e0574a3a7c1ae47981b27c2c9f8aff1",
		"3a58c45d0d1aa6ab8c87df5e2114ee0012fd46c0",
		"87f8d724575b7d7caaef81b7ddb6cf9a31ddb47c",
		"b8ee84c44ca4035f578eb3c20d304c1a25cd9b6e"}

	repo, _ := git.PlainOpen(".")
	from := "7e6a7f39324f8feb1c24562fb70650ead4e42604"
	to := "b8ee84c44ca4035f578eb3c20d304c1a25cd9b6e"

	hashes, _ := hashesToInspect(*repo, from, to)

	hashesMatches := testEq(expectedHashes, hashes)
	if !hashesMatches {
		t.Errorf("Hashes returned do not match the expecteds")
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

func TestDirtyCommitsBringsCommitsWithSecretsAddedAndRemoved(t *testing.T) {

	JWT_PATTERN := regexp.MustCompile(`eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]*`)
	expectedDirtyCommits := []string{
		"140e081f12e3d462113912311c375f0d4ad1c7ec",
		"43c37ddecac9a93301d15ee2d0a56edac8fb1ad9"}
	commits := []string{
		"3cbea9e48dc5c25bef1ccef8f4c526e9c612f51f",
		"140e081f12e3d462113912311c375f0d4ad1c7ec",
		"ed8e5c0f79cea01453d7251f6764d6a4528b1a66",
		"43c37ddecac9a93301d15ee2d0a56edac8fb1ad9",
		"95902cdb04eace68e76e4084c8516a00668ef3de",
		"7e6a7f39324f8feb1c24562fb70650ead4e42604",
		"62e99d7b4d6e628f175f0650275c2749babcc599",
		"fc9ced076e0574a3a7c1ae47981b27c2c9f8aff1",
		"3a58c45d0d1aa6ab8c87df5e2114ee0012fd46c0",
		"87f8d724575b7d7caaef81b7ddb6cf9a31ddb47c",
		"b8ee84c44ca4035f578eb3c20d304c1a25cd9b6e",
		"80e33de42b562df4f3abb5a3c340409b9417471a",
		"d644566ed7b990cff8b5a480bfdcc37bfa911916"}
	repo, _ := git.PlainOpen(".")

	dirtyCommits := getDirtyCommits(*repo, commits, *JWT_PATTERN)

	commitsMatches := testEq(expectedDirtyCommits, dirtyCommits)
	if ! commitsMatches {
		t.Errorf("Wrong dirtycommits")
	}
}
