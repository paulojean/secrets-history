package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"gopkg.in/src-d/go-git.v4"
)

func main() {
	var err error
	repoPath := flag.String("path", "", "path to a local git project")
	from := flag.String("from", "", "start commit to search, ie: newest commit too look")
	useDefaultPatterns := flag.Bool("default-patterns", true, "use default pattern credentials")
	patternsDirectory := flag.String("credential-patterns", "", "json file to use custom patterns on search")
	to := flag.String("to", "", "final commit to search, ie: oldest commit to look")
	flag.Parse()

	if ! *useDefaultPatterns && *patternsDirectory == "" {
		fmt.Println("either use default patterns or provide custom ones")
		os.Exit(1)
	}

	searchDirtyCommits(err, *patternsDirectory, *useDefaultPatterns, *repoPath, *from, *to)
}

func searchDirtyCommits(err error, patternsDirectory string, useDefaultPatterns bool, repoPath string, from string, to string) {
	securityCredentialPatterns, err := getCredentialPatterns(patternsDirectory, useDefaultPatterns)
	if err != nil {
		panic(err.Error())
	}
	repoAbsolutePath, _ := filepath.Abs(repoPath)
	err = repositoryExists(repoAbsolutePath)
	if err != nil {
		panic(err.Error())
	}
	repo, _ := git.PlainOpen(repoAbsolutePath)
	head, _ := repo.Head()
	startCommit := getStartCommit(*head, from)
	commits, err := hashesToInspect(*repo, startCommit, to)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	} else {
		fmt.Println(getDirtyCommits(*repo, commits, securityCredentialPatterns))
	}
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
