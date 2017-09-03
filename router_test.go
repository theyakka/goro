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
	"net/http/httptest"
	"reflect"
	"testing"
)

type RequestMock struct {
	URL                string
	Method             string
	CheckSuccess       bool
	CheckParams        bool
	CheckParamsValue   map[string][]string
	CheckCatchAll      bool
	CheckCatchAllValue string
}

type TestFilter struct {
}

func testHandler1(chain Chain, rw http.ResponseWriter, req *http.Request) {
	log.Println("TH1")
	chain.Next(req)
}

func testHandler2(chain Chain, rw http.ResponseWriter, req *http.Request) {
	log.Println("TH2")
	chain.Next(req)
}

func testHandler3(chain Chain, rw http.ResponseWriter, req *http.Request) {
	log.Println("TH3")
	chain.Next(req)
}

func (tf TestFilter) ExecuteFilter(req **http.Request) {
	oldReq := *req
	newCtx := context.WithValue(oldReq.Context(), "TESTVAL", "this is a test")
	*req = oldReq.WithContext(newCtx)
}

func okHandler(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	Log("OK!")
	fmt.Fprintf(rw, "OK: called '%s' -> %s", ctx.Value(PathContextKey), req.Method)
}

func errHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, "error")
}

func compareMaps(map1 map[string][]string, map2 map[string][]string) bool {
	return reflect.DeepEqual(map1, map2)
}

func TestMain(t *testing.T) {

	domains := NewDomainMap()
	router := domains.NewRouter("^(?:www+[.])*(localhost.local)(?::\\d+)?")
	router.SetDebugLevel(DebugLevelTimings)

	testFilter := TestFilter{}
	router.AddFilter(testFilter)

	router.AddStaticWithPrefix("./assets", "assets")
	router.AddStaticWithPrefix("./test", "test")

	calledTestHandler := false
	var finalParams map[string][]string
	var finalCatchAll string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("Test handler!!")
		calledTestHandler = true
		ctx := req.Context()
		finalParams = ctx.Value(ParametersContextKey).(map[string][]string)
		catchAllObj := ctx.Value(CatchAllValueContextKey)
		finalCatchAll = ""
		if catchAllObj != nil {
			finalCatchAll = catchAllObj.(string)
		}
	})

	v2TestHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("V2 Test handler!!")
		calledTestHandler = true
		ctx := req.Context()
		finalParams = ctx.Value(ParametersContextKey).(map[string][]string)
		catchAllObj := ctx.Value(CatchAllValueContextKey)
		finalCatchAll = ""
		if catchAllObj != nil {
			finalCatchAll = catchAllObj.(string)
		}
	})

	chain := NewChainWithFuncs(testHandler1, testHandler3, testHandler2)
	router.BeforeChain = chain

	testHandle := testHandler

	v2group := router.Group("/v2")
	v2group.Add("GET", "/_ping").Handle(v2TestHandler)
	v2group.Add("GET", "/_status").Handle(v2TestHandler)
	v2group.Add("GET", "/").Handle(v2TestHandler)
	v2group.Add("GET", "/*").Handle(v2TestHandler)

	router.SetStringVariable("a", "alpha$b$c")
	router.SetStringVariable("b", "bar")
	router.SetStringVariable("c", "$d")
	router.SetStringVariable("d", "baz")

	router.Add("GET", "/").
		Handle(testHandle)
	router.Add("POST", "/login").
		Handle(testHandle)
	router.Add("GET", "/users/:id/*").
		Handle(testHandle)
	router.Add("GET", "/users/:id/show").
		Handle(testHandle)
	router.Add("POST", "/users/:id/show").
		Handle(testHandle)
	router.Add("GET", "/users/:id/:action").
		Handle(testHandle)
	router.Add("GET", "/users/:id/show/:what").
		Handle(testHandle)
	router.Add("GET", "/*").
		Handle(testHandle)
	router.Add("GET", "/$a$b").
		Handle(testHandle)

	reqMocks := []RequestMock{
		{Method: "GET", URL: "/", CheckSuccess: true},
		{Method: "POST", URL: "/", CheckSuccess: false},
		{Method: "GET", URL: "/users/123/show", CheckSuccess: true},
		{Method: "GET", URL: "/test/route", CheckSuccess: true},
		{Method: "POST", URL: "/test/route", CheckSuccess: false},
		{Method: "GET", URL: "/users/123/show/something", CheckSuccess: true},
		{Method: "GET", URL: "users/123/show/something", CheckSuccess: true},
		{Method: "POST", URL: "/login", CheckSuccess: true},
		{Method: "GET", URL: "/login", CheckSuccess: false},
		{Method: "GET", URL: "/something", CheckSuccess: true},
		{Method: "GET", URL: "/v2/_ping", CheckSuccess: true},
		{Method: "GET", URL: "/v2/testing", CheckSuccess: true},
	}

	router.PrintRoutes()

	for _, mock := range reqMocks {
		Log("Calling:", mock.URL, "------")
		r, _ := http.NewRequest(mock.Method, mock.URL, nil)
		w := httptest.NewRecorder()
		calledTestHandler = false
		router.ServeHTTP(w, r)
		Log("• TEST: url =", mock.URL)
		Log("  method =", mock.Method, "; called handler? =", calledTestHandler)
		if calledTestHandler != mock.CheckSuccess {
			t.Error("*** Handler check failed for", mock.URL)
		}
		Log("------\n")
	}

	paramMocks := []RequestMock{
		RequestMock{
			Method:       "GET",
			URL:          "/users/123/show",
			CheckSuccess: true,
			CheckParams:  true,
			CheckParamsValue: map[string][]string{
				"id": []string{"123"},
			},
		},
		RequestMock{
			Method:             "GET",
			URL:                "/test/route",
			CheckSuccess:       true,
			CheckCatchAll:      true,
			CheckCatchAllValue: "test/route",
		},
		RequestMock{
			Method:       "GET",
			URL:          "/users/123/show/something",
			CheckSuccess: true,
			CheckParams:  true,
			CheckParamsValue: map[string][]string{
				"id":   []string{"123"},
				"what": []string{"something"},
			},
		},
		RequestMock{
			Method:       "GET",
			URL:          "/users/123/show/something?test=55",
			CheckSuccess: true,
			CheckParams:  true,
			CheckParamsValue: map[string][]string{
				"id":   []string{"123"},
				"what": []string{"something"},
			},
		},
		RequestMock{
			Method:       "GET",
			URL:          "/users/123/show/something?what=55",
			CheckSuccess: true,
			CheckParams:  true,
			CheckParamsValue: map[string][]string{
				"id":   []string{"123"},
				"what": []string{"something"},
			},
		},
	}

	for _, mock := range paramMocks {
		r, _ := http.NewRequest(mock.Method, mock.URL, nil)
		w := httptest.NewRecorder()
		calledTestHandler = false
		finalParams = nil
		finalCatchAll = ""
		router.ServeHTTP(w, r)
		Log("Testing: ", mock.URL)
		Log("  should_pass:", calledTestHandler, " method:", mock.Method)
		if calledTestHandler != mock.CheckSuccess {
			t.Error("*** Handler check failed for", mock.URL)
		} else {
			Log("  params: ", finalParams)
			Log("  catch-all:", finalCatchAll)
			if mock.CheckCatchAll == true && mock.CheckCatchAllValue != "" {
				if finalCatchAll != mock.CheckCatchAllValue {
					t.Error("Catch-all value check failed")
				}
			}
			if mock.CheckParams == true && mock.CheckParamsValue != nil {
				if !compareMaps(finalParams, mock.CheckParamsValue) {
					t.Error("Params value check failed")
				}
			}
		}
	}

}
