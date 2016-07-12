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
	// "fmt"
	"log"
	"net/http"
	"os"
)

// logger - shared logger instance
var logger *log.Logger

// ContextHandler - custom HTTP handler that also returns the current Context
type ContextHandler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request)
}

// ContextHandlerFunc - custom HTTP handler function type that also returns the current Context
type ContextHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// ServeHTTPContext - wrapper for required ContextHandler ServeHTTPContext
func (h ContextHandlerFunc) ServeHTTPContext(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	h(ctx, rw, req)
}

// RouteComponentType - route component types
// NOTE: variables will be stripped out / replaced so we dont track them
type RouteComponentType int

const (
	// ComponentTypeFixed - a fixed path component
	ComponentTypeFixed RouteComponentType = 1 << iota
	// ComponentTypeWildcard - a wildcard path component
	ComponentTypeWildcard
	// ComponentTypeCatchAll - catch all route component
	ComponentTypeCatchAll
)

const (
	// HTTPMethodGET - GET http method
	HTTPMethodGET string = "GET"
	// HTTPMethodPOST - POST http method
	HTTPMethodPOST string = "POST"
	// HTTPMethodPUT - PUT http method
	HTTPMethodPUT string = "PUT"
	// HTTPMethodDELETE - DELETE http method
	HTTPMethodDELETE string = "DELETE"
)

// initLogger - initializes the shared logger instance
func initLogger() {
	logger = log.New(os.Stdout, "GORO: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Log - logging wrapper for standard output to log
func Log(v ...interface{}) {
	if logger == nil {
		initLogger()
	}
	logger.Println(v...)
}
