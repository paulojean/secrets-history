package main

import (
	"regexp"
	"io/ioutil"
	"encoding/json"
)

type SecurityCredential struct {
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
}

const DEFAULT_PATTERN_PATH = "resources/default_patterns.json"

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
		defaultPatterns, err = parseDefaultPatternFile(DEFAULT_PATTERN_PATH)
	}

	if err != nil {
		return nil, err
	}

	patterns := securityCredentialsToRegex(append(credentialsPatterns, defaultPatterns...))

	return patterns, nil
}

func parsePatternFile(path string) ([]SecurityCredential, error) {
	return parseFiles(ioutil.ReadFile(path))
}

func parseDefaultPatternFile(path string) ([]SecurityCredential, error) {
	return parseFiles(Asset(path))
}

func parseFiles(data []byte, err error) ([]SecurityCredential, error) {
	if err != nil {
		return nil, err
	}

	var credentials []SecurityCredential
	json.Unmarshal(data, &credentials)
	return credentials, nil
}

func securityCredentialsToRegex(securityCredentials []SecurityCredential) []regexp.Regexp {
	patterns := make([]regexp.Regexp, len(securityCredentials))
	for i, v := range securityCredentials {
		patterns[i] = *regexp.MustCompile(v.Pattern)
	}
	return patterns
}
