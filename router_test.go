package goro

import (
	"fmt"
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
		"/users/{$id_format}",
		"/users/{first}.{second}",
		"/users/{id}/{$operation}",
		"/test/this/thing/{blah:[A-Z]+}",
	}

	for _, path := range paths {
		router.GET(path, nil)
	}
	fmt.Printf("registered routes:\n%v", router.registedRoutes)
	fmt.Printf("\n")
}
