package goro

import (
	// "fmt"
	"net/http"
	"testing"
)

func TestMatcher(t *testing.T) {

	router := NewRouter()
	router.SetVariable("id-format", "{id:^[a-zA-Z0-9_]*$}")

	routePaths := []string{
		"/users/{$id-format}/action/{action}",
		"/users/{$id-format}",
		"/test/this",
		"/test/{$bad-var}",
		"/monkey/add",
	}

	for _, routePath := range routePaths {
		router.GET(routePath, func(w http.ResponseWriter, req *http.Request) {})
	}

	router.PrintRoutes()

}
