// router_test.go
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
	"fmt"
	"log"
	"net/http"
	"testing"
)

func globalHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	})
}

func TestMatcher(t *testing.T) {

	fmt.Println() // blank line

	router := NewRouter()
	router.SetVariable("id-format", "{id:^[a-zA-Z0-9_]*$}")
	router.AddGlobalHandler("OPTIONS", globalHandler())

	router.POST("/users/{$id-format}/action/{action}", func(w http.ResponseWriter, req *http.Request) {})
	router.GET("/users/{$id-format}", func(w http.ResponseWriter, req *http.Request) {})
	router.GET("/test/this", func(w http.ResponseWriter, req *http.Request) {})
	router.DELETE("/test/{$bad-var}", func(w http.ResponseWriter, req *http.Request) {})
	router.PUT("/monkey/update", func(w http.ResponseWriter, req *http.Request) {})

	router.PrintRoutes()

	log.Fatal(http.ListenAndServe(":9900", router))

}
