package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"gopkg.in/src-d/go-git.v4"
	"encoding/json"
)

func main() {
	from := flag.String("from", "", "start commit to search, ie: newest commit too look")
	patternsDirectory := flag.String("credential-patterns", "", "json file to use custom patterns on search")
	repoPath := flag.String("path", "", "path to a local git project")
	to := flag.String("to", "", "final commit to search, ie: oldest commit to look")
	useDefaultPatterns := flag.Bool("default-patterns", true, "use default pattern credentials")
	printDefaulPatterns := flag.Bool("print-default-patterns", false, "print the file with the default credential patterns")
	flag.Parse()

	if *printDefaulPatterns {
		printDefaultPatternFile()
	}

	if ! *useDefaultPatterns && *patternsDirectory == "" {
		printErrorAndStop(errors.New("either use default patterns or provide custom ones"))
	}

	searchDirtyCommits(*patternsDirectory, *useDefaultPatterns, *repoPath, *from, *to)
}

func printDefaultPatternFile() {
	patterns, err := parsePatternFile(DEFAULT_PATTERN_PATH)
	if err != nil {
		printErrorAndStop(err)
	}

	patternsMap := map[string]interface{}{}
	for _, pattern := range patterns {
		patternsMap[pattern.Kind] = pattern.Pattern
	}

	prettyPatterns, err := json.MarshalIndent(patternsMap, "", "  ")
	if err != nil {
		printErrorAndStop(err)
	}

	fmt.Println(string(prettyPatterns))
	os.Exit(0)
}

func searchDirtyCommits(patternsDirectory string, useDefaultPatterns bool, repoPath string, from string, to string) {
	securityCredentialPatterns, err := getCredentialPatterns(patternsDirectory, useDefaultPatterns)
	if err != nil {
		panic(err.Error())
	}

	repoAbsolutePath, _ := filepath.Abs(repoPath)
	err = repositoryExists(repoAbsolutePath)
	if err != nil {
		printErrorAndStop(err)
	}

	repo, err := git.PlainOpen(repoAbsolutePath)
	if err != nil {
		printErrorAndStop(err)
	}

	head, err := repo.Head()
	if err != nil {
		printErrorAndStop(err)
	}

	startCommit := getStartCommit(*head, from)
	commits, err := hashesToInspect(*repo, startCommit, to)
	if err != nil {
		printErrorAndStop(err)
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

func printErrorAndStop(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}
