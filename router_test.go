// Goro
//
// Created by Posse in NYC
// http://goposse.com
//
// Copyright (c) 2016 Posse Productions LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

type TestFilter struct {
}

func (tf TestFilter) ExecuteFilter(ctx context.Context, req *http.Request) context.Context {
	newCtx := context.WithValue(ctx, "TESTING!!!", "this is a test")
	return newCtx
}

func okHandler(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	Log(ctx)
	fmt.Fprintf(rw, "OK: called '%s'", ctx.Value("path"))
}

func errHandler(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	Log(ctx)
	fmt.Fprint(rw, "error")
}

func TestMain(t *testing.T) {

	router := NewRouter()
	router.SetDebugLevel(DebugLevelTimings)

	testFilter := TestFilter{}
	router.AddFilter(testFilter)

	// error handlers
	router.SetErrorHandlerFunc(http.StatusNotFound, errHandler)
	router.SetErrorHandlerFunc(http.StatusMethodNotAllowed, errHandler)

	router.Add("GET", "/").
		HandleFunc(okHandler).Describe("The root route")
	router.Add("GET", "/users/:id/*").
		HandleFunc(okHandler)
	router.Add("GET", "/users/:id/show").
		HandleFunc(okHandler)
	router.Add("GET", "/users/:id/:action").
		HandleFunc(okHandler)
	router.Add("GET", "/users/:id/show/:what").
		HandleFunc(okHandler)

	// router.PrintTreeInfo()

	Log("Server running on :8080")
	http.ListenAndServe(":8080", router)
}
