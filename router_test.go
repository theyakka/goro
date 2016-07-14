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

func (tf TestFilter) ExecuteFilter(req **http.Request) {
	oldReq := *req
	newCtx := context.WithValue(oldReq.Context(), "TESTVAL", "this is a test")
	*req = oldReq.WithContext(newCtx)
}

func okHandler(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	fmt.Fprintf(rw, "OK: called '%s' -> %s", ctx.Value("path"), req.Method)
}

func errHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, "error")
}

func TestMain(t *testing.T) {

	domains := NewDomainMap()
	router := domains.NewRouter("^(?:www+[.])*(localhost.local)(?::\\d+)?")
	router.SetDebugLevel(DebugLevelTimings)

	testFilter := TestFilter{}
	router.AddFilter(testFilter)

	router.AddStatic("./assets")

	// error handlers
	// router.SetErrorHandlerFunc(http.StatusNotFound, errHandler)
	// router.SetErrorHandlerFunc(http.StatusMethodNotAllowed, errHandler)

	router.Add("GET", "/").
		HandleFunc(okHandler).Describe("The root route")
	router.Add("GET", "/users/:id/*").
		HandleFunc(okHandler)
	router.Add("GET", "/users/:id/show").
		HandleFunc(okHandler)
	router.Add("POST", "/users/:id/show").
		HandleFunc(okHandler).Describe("POST form of the route")
	router.Add("GET", "/users/:id/:action").
		HandleFunc(okHandler)
	router.Add("GET", "/users/:id/show/:what").
		HandleFunc(okHandler)
	router.Add("GET", "/*").
		HandleFunc(okHandler)

	router.PrintRoutes()

	Log("Server running on :8080")
	fmt.Println("")
	http.ListenAndServe(":8080", domains)
}
