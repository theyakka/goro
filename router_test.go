package goro

import (
	"fmt"
	"net/http"
	"testing"
)

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
	fmt.Println("HERE")
}

func TestRouter(t *testing.T) {

	router := NewRouter()
	router.AddStringVar("id_format", "{id:[a-Z]+}")
	router.AddStringVar("operation", "this_op")

	paths := []string{
		"hello",
		"hello/{id}",
		"/",
		"/{something}",
		"/users/{$id_format}",
		"/users/{first}.{second}",
		"/users/{$id_format}/{$operation}",
		"/test/this/thing/{blah:[A-Z]+}",
	}

	for _, path := range paths {
		router.GET(path, testHandler)
	}

	router.PrintRoutes()

	w := new(mockResponseWriter)
	req, _ := http.NewRequest("GET", "/users/1234", nil)
	router.ServeHTTP(w, req)

	fmt.Printf("\n")
}
