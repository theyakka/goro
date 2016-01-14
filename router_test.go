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

func testHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	fmt.Println("MATCHED HANDLER FOR PATH")
	matchedRoute := context.Get(r, "matched_route")
	if matchedRoute != nil {
		fmt.Printf("  - Route: %s\n", (matchedRoute.(route)).PathFormat)
	}
	if context != nil {
		fmt.Printf("  - ID: %v\n", context.Get(r, "id"))
		fmt.Printf("  - some_val: %s\n", context.GetString(r, "some_val"))
	}
	fmt.Printf("  - Path: %s\n", r.URL.Path)

	return http.StatusOK, nil
}

func doFirstHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	fmt.Println("DO THIS FIRST")
	context.Put(r, "some_val", "donkey")
	return http.StatusOK, nil
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("")
	fmt.Println("NOT FOUND!")
}

func notAllowedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("")
	fmt.Println("NOT ALLOWED!")
}

func panicHandler(w http.ResponseWriter, r *http.Request) {

	context.Get(r, "panic")

	fmt.Println("")
	fmt.Println("PANIC!")
}

func TestRouter(t *testing.T) {

	fmt.Printf("\n")

	chainer := NewChainer(doFirstHandler)

	context = NewContext()
	router := NewRouter()
	router.Context = context
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(notAllowedHandler)
	router.PanicHandler = http.HandlerFunc(panicHandler)
	router.ShouldRedirectTrailingSlash = true
	router.AddStringVar("$id_format", "{id}")
	router.AddStringVar("$operation", "this_op")

	paths := []string{
		"hello",
		"hello/{id}",
		"/",
		"/{something}",
		"/users/{$id_format}/{$operation}",
		"/test/this/thing/{blah:[A-Z]+}",
		"/users/{name}",
		"/users/{$id_format}",
	}

	for _, path := range paths {
		router.GET(path, chainer.Then(testHandler))
	}

	router.PrintRoutes()

	checkPath := "/users/1234"
	w := new(mockResponseWriter)
	req, _ := http.NewRequest("GET", checkPath, nil)
	router.ServeHTTP(w, req)

	req2, _ := http.NewRequest("GET", checkPath, nil)
	router.ServeHTTP(w, req2)

	req3, _ := http.NewRequest("GET", checkPath, nil)
	router.ServeHTTP(w, req3)

	fmt.Printf("\n")
}
