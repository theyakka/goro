package goro

import (
	"fmt"
	"net/http"
	"testing"
)

var context *Context

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("MATCHED HANDLER FOR PATH")
	matchedRoute := context.Get("matched_route")
	if matchedRoute != nil {
		fmt.Printf("  - Route: %s\n", (matchedRoute.(route)).PathFormat)
	}
	if context != nil {
		fmt.Printf("  - ID: %v\n", context.Get("id"))
	}
	fmt.Printf("  - Path: %s\n", r.URL.Path)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("")
	fmt.Println("NOT FOUND!")
}

func TestRouter(t *testing.T) {

	fmt.Printf("\n")

	context = NewContext()
	router := NewRouter()
	router.Context = context
	router.NotFoundHandler = notFoundHandler
	router.AddStringVar("$id_format", "{id}")
	router.AddStringVar("$operation", "this_op")

	paths := []string{
		"/users/{$id_format}",
		"hello",
		"hello/{id}",
		"/",
		"/{something}",
		"/users/{name}",
		"/users/{$id_format}/{$operation}",
		"/test/this/thing/{blah:[A-Z]+}",
	}

	for _, path := range paths {
		router.GET(path, testHandler)
	}

	router.PrintRoutes()

	checkPath := "/users/aa-aaaa-aaaa/this_op"
	w := new(mockResponseWriter)
	req, _ := http.NewRequest("GET", checkPath, nil)
	router.ServeHTTP(w, req)

	fmt.Printf("\n")
}
