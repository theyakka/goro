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
	"log"
	"net/http"
	"testing"
)

func rootHandler(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	log.Println("root")
}

func testThisThing(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	log.Println("HERE!!!")
	log.Println("ctx:", ctx.Value("path"))
}

func notFoundHandler(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	Log(ctx)
	fmt.Fprint(rw, "error")
}

func TestMain(t *testing.T) {

	router := NewRouter()
	router.EnableDebugMode(true)

	router.AddStringVariable("someop", "turnip")
	router.AddStringVariable("idval", ":id")

	router.SetErrorHandlerFunc(http.StatusNotFound, notFoundHandler)

	router.Add("GET", "/").HandleFunc(rootHandler).Describe("The root route")
	// router.Add("GET", "/users/$idval/:action")
	router.Add("GET", "/users/:id/*")
	router.Add("GET", "/users/:id/show")
	router.Add("GET", "/users/:id/:action")
	router.Add("GET", "/users/:id/show/:what")

	// router.PrintTreeInfo()

	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", router)
}
