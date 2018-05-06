package goro

import "testing"

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
