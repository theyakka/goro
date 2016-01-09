package goro

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
)

func TestRouter(t *testing.T) {

	paths := []string{
		"hello",
		"hello/{id}",
		"/",
		"/{something}",
		"/users/{id}",
		"/users/{first}.{second}",
		"/users/{id}/{$operation}",
		"/test/this/thing/{blah:[A-Z]+}",
	}

	for _, path := range paths {
		_, _, parseErr := parsePath(path)
		log.Printf("Path = %s", path)
		if parseErr != nil {
			log.Printf("  • Error: %s\n", parseErr)
		}
	}
	fmt.Printf("\n")
}

func parsePath(path string) (finalPath string, wildcards []string, parseErr error) {
	if !strings.HasPrefix(path, "/") {
		// missing slash at the start, we aaaaare out
		return "", []string{}, errors.New("Path is missing leading slash ('/')")
	}

	hasWildcard := (strings.Index(path, "{") != -1)
	if !hasWildcard {
		// no wildcards, return now
		return path, []string{}, nil
	}

	pathComps := strings.Split(path, "/")[1:]
	for _, comp := range pathComps {
		matcher := NewMatcher(comp, "{", "}")
		match := matcher.NextMatch()
		for match != NotFoundMatch() {
			log.Println("match = ", match.Value)
			match = matcher.NextMatch()
		}
	}

	return "", nil, nil
}
