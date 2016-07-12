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
	"path/filepath"
	"strings"
)

// Router is the main routing class
type Router struct {
	// NotFoundHandler - called when no route was matched
	NotFoundHandler ContextHandler

	// globalContext is the global Context object
	globalContext context.Context

	// routeMatcher - the primary route matcher instance
	routeMatcher *Matcher

	// variables - unwrapped (clean) variables that have been defined
	variables map[string]string

	// methodKeyedRoutes - all routes registered with the router
	routes *Tree

	globalHandlers map[string]ContextHandler

	// debugMode - if enabled will output debugging information
	debugMode bool
}

// NewRouter - creates a new default instance of the Router type
func NewRouter() *Router {
	router := &Router{
		globalContext:  context.Background(),
		variables:      map[string]string{},
		routes:         &Tree{},
		globalHandlers: map[string]ContextHandler{},
		debugMode:      false,
	}
	router.routeMatcher = NewMatcher(router)
	return router
}

// EnableDebugMode - enables or disables debug mode
func (r *Router) EnableDebugMode(enabled bool) {
	onOffString := "on"
	if !enabled {
		onOffString = "off"
	}
	Log("Debug mode is", onOffString)
	r.debugMode = enabled
	r.routeMatcher.LogMatchTime = enabled
}

// NewMatcher returns a new matcher for the given Router
func (r *Router) NewMatcher() *Matcher {
	return NewMatcher(r)
}

// Add creates a new Route and registers the instance within the Router
func (r *Router) Add(method string, path string) *Route {
	route := NewRoute(method, path)
	return r.Use(route)
}

// Use registers a Route instance within the Router
func (r *Router) Use(route *Route) *Route {
	r.routes.AddRouteToTree(route, r.variables)
	return route
}

// SetGlobalHandler configures a ContextHandler to handle all requests for a given method
func (r *Router) SetGlobalHandler(method string, handler ContextHandler) {
	r.globalHandlers[strings.ToUpper(method)] = handler
}

// SetGlobalHandlerFunc configures a ContextHandlerFunc to handle all requests for a given method
func (r *Router) SetGlobalHandlerFunc(method string, handlerFunc ContextHandlerFunc) {
	r.SetGlobalHandler(method, ContextHandler(handlerFunc))
}

// AddStringVariable adds a string variable value for substitution
func (r *Router) AddStringVariable(variable string, value string) {
	varname := variable
	if !strings.HasPrefix(varname, "$") {
		varname = "$" + varname
	}
	r.variables[varname] = value
}

// PrintTreeInfo prints debugging information about all registered Routes
func (r *Router) PrintTreeInfo() {
	for _, node := range r.routes.nodes {
		fmt.Println(" - ", node)
		printSubNodes(node, 0)
	}
}

func printSubNodes(node *Node, level int) {
	if node.HasChildren() {
		for _, subnode := range node.nodes {
			indent := ""
			for i := 0; i < level+1; i++ {
				indent += " "
			}
			indent += "-"
			fmt.Println("", indent, " ", subnode)
			if subnode.HasChildren() {
				printSubNodes(subnode, level+1)
			}
		}
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	useReq := req
	usePath := strings.ToLower(useReq.URL.Path) // always compare lower
	usePath = filepath.Clean(usePath)
	method := strings.ToUpper(req.Method)

	outCtx := context.WithValue(r.globalContext, "path", usePath)

	// check to see if a global handler has been registered for the method
	globalHandler := r.globalHandlers[method]
	if globalHandler != nil {
		Log("should handle", method, "globally")
		globalHandler.ServeHTTPContext(outCtx, w, req)
		return
	}

	// check to see if we have a matching route
	matchError := ""
	matchErrorCode := 0
	match := r.routeMatcher.MatchPathToRoute(method, usePath)
	if match != nil {
		Log("Match:", match)
		// TODO - error checking
		route := match.Node.routes[method]
		if route != nil {
			handler := route.Handler
			if handler != nil {
				handler.ServeHTTPContext(outCtx, w, req)
				return
			}
		} else {
			matchError = "Method Not Allowed"
			matchErrorCode = http.StatusMethodNotAllowed
		}
	} else {
		matchError = "Not Found"
		matchErrorCode = http.StatusNotFound
	}

	if matchErrorCode != 0 {
		// not found error
		if matchErrorCode == http.StatusNotFound {
			if r.NotFoundHandler != nil {
				r.NotFoundHandler.ServeHTTPContext(outCtx, w, req)
				return
			}
		}
		// method not allowed
		if matchErrorCode == http.StatusMethodNotAllowed {
			return
		}
		// return a generic http error
		errorHandler(w, req, matchError, matchErrorCode)
	}
}

func errorHandler(w http.ResponseWriter, req *http.Request, errorString string, errorCode int) {
	http.Error(w, errorString, errorCode)
}
