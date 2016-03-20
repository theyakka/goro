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
	"os"
	"testing"
)

var testNums []uint32
var testNumsMap map[uint32]bool

func setup() {
	testNums = []uint32{}
	testNumsMap = map[uint32]bool{}
	var i uint32
	for i = 0; i < 1000000; i++ {
		testNums = append(testNums, i)
		testNumsMap[i] = true
	}
	fmt.Println("setup complete")
}

func teardown() {
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func BenchmarkMapSearch(b *testing.B) {
	b.ResetTimer()
	b.StartTimer()
	var x bool
	x = testNumsMap[900604]
	if x == true {
		b.StopTimer()
	}
}

func BenchmarkSliceSearchLinear(b *testing.B) {
	b.ResetTimer()
	b.StartTimer()
	for num := range testNums {
		if num == 900604 {
			break
		}
	}
	b.StopTimer()
}

func globalHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("DO OPTIONS!!!")
	})
}

func TestRouter(t *testing.T) {

	fmt.Println() // blank line

	router := NewRouter()
	router.SetVariable("id-format", "{id}")
	router.AddGlobalHandler("OPTIONS", globalHandler())

	router.GET("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("I AM Root")
	})
	router.GET("/users", func(w http.ResponseWriter, req *http.Request) {})
	router.POST("/users/{$id-format}/action/{action}", func(w http.ResponseWriter, req *http.Request) {})
	router.GET("/users/{turnip}/hello", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("OH HEEEEELLLLLLOOOOOOOOO")
	})
	router.GET("/users/{$id-format}", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("LOOK FOR USER with id")
	})
	router.GET("/test/this", func(w http.ResponseWriter, req *http.Request) {})
	router.DELETE("/test/{$bad-var}", func(w http.ResponseWriter, req *http.Request) {})
	router.PUT("/monkey/update", func(w http.ResponseWriter, req *http.Request) {})

	router.PrintRoutes()

	router.PrintTrees()

	log.Fatal(http.ListenAndServe(":9900", router))

}
