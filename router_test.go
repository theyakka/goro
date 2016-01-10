package goro

import (
	"fmt"
	"log"
	"testing"
)

func TestRouter(t *testing.T) {

	router := NewRouter()
	router.AddStringVar("id_format", "id:[a-Z]+")

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
		_, wc, parseErr := parsePath(path)
		log.Printf("Path = %s\n", path)
		log.Printf("  - %v\n", wc)
		if parseErr != nil {
			log.Printf("  • Error: %s\n", parseErr)
		}
	}
	fmt.Printf("\n")
}
