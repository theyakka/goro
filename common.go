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
	// "fmt"
	"log"
	"os"
)

// logger - shared logger instance
var logger *log.Logger

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

// DebugLevel - debug information output level
type DebugLevel int

const (
	// DebugLevelNone - debugging is off
	DebugLevelNone DebugLevel = 1 << iota
	// DebugLevelTimings - show timings only
	DebugLevelTimings
	// DebugLevelFull - show all debugging information
	DebugLevelFull
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
