package goro

import (
	"context"
	"testing"
)

// Route params tests
func readyContext() context.Context {
	paramsMap := map[string][]string{
		"id":        []string{"255"},
		"colors":    []string{"red", "green", "blue"},
		"names":     []string{"John Smith"},
		"positions": []string{"left", "topwise", "other right"},
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, ParametersContextKey, paramsMap)
	return ctx
}

func TestRouteParamsFromContext(t *testing.T) {
	ctx := readyContext()
	vals := RouteParamsFromContext(ctx)
	if vals == nil {
		t.Error("Failed to retrieve parameters from context")
	}
}

func TestRouteParamsWithoutID(t *testing.T) {
	ctx := readyContext()
	vals := RouteParamsWithoutID(RouteParamsFromContext(ctx))
	if vals.Get("id") != "" {
		t.Error("Got parameters from context but 'id' value is still present")
	}
}

func TestFirstStringRouteParam(t *testing.T) {
	ctx := readyContext()
	vals := RouteParamsWithoutID(RouteParamsFromContext(ctx))
	paramVals := vals["colors"]
	firstString := FirstStringRouteParam(paramVals)
	expectedString := "red"
	if firstString != expectedString {
		t.Errorf("Got bad parameter value. Got: '%s', expected: '%s'", firstString, expectedString)
	}
}

// Clean path tests
func TestCleanGoodPath(t *testing.T) {
	badPath := "/number/1/test"
	expectedPath := "/number/1/test"
	cleanPath := CleanPath(badPath)
	if cleanPath != expectedPath {
		t.Errorf("Path not clean. Got: '%s', expected: '%s'", cleanPath, expectedPath)
	}
}

func TestInternalDoubleSlash(t *testing.T) {
	badPath := "/testing//this"
	expectedPath := "/testing/this"
	cleanPath := CleanPath(badPath)
	if cleanPath != expectedPath {
		t.Errorf("Path not clean. Got: '%s', expected: '%s'", cleanPath, expectedPath)
	}
}

func TestCleanDoubleSlash(t *testing.T) {
	badPath := "//hello/world"
	expectedPath := "/hello/world"
	cleanPath := CleanPath(badPath)
	if cleanPath != expectedPath {
		t.Errorf("Path not clean. Got: '%s', expected: '%s'", cleanPath, expectedPath)
	}
}

func TestCleanWrongSlash(t *testing.T) {
	badPath := "\\good/times"
	expectedPath := "/good/times"
	cleanPath := CleanPath(badPath)
	if cleanPath != expectedPath {
		t.Errorf("Path not clean. Got: '%s', expected: '%s'", cleanPath, expectedPath)
	}
}

func TestCleanWrongSlashes(t *testing.T) {
	badPath := "\\windows\\is\\cool"
	expectedPath := "/windows/is/cool"
	cleanPath := CleanPath(badPath)
	if cleanPath != expectedPath {
		t.Errorf("Path not clean. Got: '%s', expected: '%s'", cleanPath, expectedPath)
	}
}
