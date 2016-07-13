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

	// ErrorHandler - generic error handler
	ErrorHandler ContextHandler

	// ShouldCacheMatchedRoutes - if true then any matched routes should be cached
	// according to the path they were matched to
	ShouldCacheMatchedRoutes bool

	// globalContext is the global Context object
	globalContext context.Context

	// errorHandlers - map status codes to specific handlers
	errorHandlers map[int]ContextHandler

	// globalHandlers - handlers that will match all requests for an HTTP method regardless
	// of route matching
	globalHandlers map[string]ContextHandler

	// filters - registered pre-process filters
	filters []Filter

	// routeMatcher - the primary route matcher instance
	routeMatcher *Matcher

	// methodKeyedRoutes - all routes registered with the router
	routes *Tree

	// variables - unwrapped (clean) variables that have been defined
	variables map[string]string

	// cache - matched routes to path mappings
	cache *RouteCache

	// debugLevel - if enabled will output debugging information
	debugLevel DebugLevel
}

// NewRouter - creates a new default instance of the Router type
func NewRouter() *Router {
	router := &Router{
		ErrorHandler:             nil,
		ShouldCacheMatchedRoutes: true,
		globalContext:            context.Background(),
		errorHandlers:            map[int]ContextHandler{},
		globalHandlers:           map[string]ContextHandler{},
		filters:                  nil,
		routes:                   &Tree{},
		variables:                map[string]string{},
		cache:                    NewRouteCache(),
		debugLevel:               DebugLevelNone,
	}
	router.routeMatcher = NewMatcher(router)
	return router
}

// SetDebugLevel - enables or disables debug mode
func (r *Router) SetDebugLevel(debugLevel DebugLevel) {
	debugTimingsOn := debugLevel == DebugLevelTimings
	debugFullOn := debugLevel == DebugLevelFull
	debugOn := debugTimingsOn || debugFullOn
	onOffString := "on"
	if !debugOn {
		onOffString = "off"
	}
	Log("Debug mode is", onOffString)
	r.debugLevel = debugLevel
	r.routeMatcher.LogMatchTime = debugOn
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

// SetErrorHandler configures a ContextHandler to handle all errors for the supplied status code
func (r *Router) SetErrorHandler(statusCode int, handler ContextHandler) {
	r.errorHandlers[statusCode] = handler
}

// SetErrorHandlerFunc configures a ContextHandlerFunc to handle all errors for the supplied status code
func (r *Router) SetErrorHandlerFunc(statusCode int, handler ContextHandlerFunc) {
	r.SetErrorHandler(statusCode, ContextHandler(handler))
}

// AddFilter adds a filter to the list of pre-process filters
func (r *Router) AddFilter(filter Filter) {
	r.filters = append(r.filters, filter)
}

// AddStringVariable adds a string variable value for substitution
func (r *Router) AddStringVariable(variable string, value string) {
	varname := variable
	if !strings.HasPrefix(varname, "$") {
		varname = "$" + varname
	}
	r.variables[varname] = value
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	useReq := req
	usePath := strings.ToLower(useReq.URL.Path) // always compare lower
	usePath = filepath.Clean(usePath)
	method := strings.ToUpper(req.Method)

	outCtx := context.WithValue(r.globalContext, "path", usePath)

	if r.filters != nil {
		for _, filter := range r.filters {
			updatedCtx := filter.ExecuteFilter(outCtx, req)
			if updatedCtx != nil {
				outCtx = updatedCtx
			}
		}
	}

	// check to see if a global handler has been registered for the method
	globalHandler := r.globalHandlers[method]
	if globalHandler != nil {
		globalHandler.ServeHTTPContext(outCtx, w, req)
		return
	}

	// check to see if we have a matching route
	matchError := ""
	matchErrorCode := 0
	match := r.routeMatcher.MatchPathToRoute(method, usePath)
	if match != nil && len(match.Node.routes) > 0 {
		route := match.Node.RouteForMethod(method)
		if route != nil {
			handler := route.Handler
			if handler != nil {
				outCtx = context.WithValue(outCtx, "params", match.WildcardValues)
				if match.CatchAllValue != "" {
					outCtx = context.WithValue(outCtx, "catchAll", match.CatchAllValue)
				}
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
		err := map[string]interface{}{
			"code":    matchErrorCode,
			"message": matchError,
		}
		outCtx = context.WithValue(outCtx, "error", err)

		// try to call specific error handler
		errHandler := r.errorHandlers[matchErrorCode]
		if errHandler != nil {
			errHandler.ServeHTTPContext(outCtx, w, req)
			return
		}
		// if generic error handler defined, call that
		if r.ErrorHandler != nil {
			r.ErrorHandler.ServeHTTPContext(outCtx, w, req)
			return
		}
		// return a generic http error
		errorHandler(w, req, matchError, matchErrorCode)
	}
}

func errorHandler(w http.ResponseWriter, req *http.Request, errorString string, errorCode int) {
	http.Error(w, errorString, errorCode)
}

// PrintTreeInfo prints debugging information about all registered Routes
func (r *Router) PrintTreeInfo() {
	for _, node := range r.routes.nodes {
		fmt.Println(" - ", node)
		printSubNodes(node, 0)
	}
}

// PrintRoutes prints route registration information
func (r *Router) PrintRoutes() {
	fmt.Println("")
	nodes := r.routes.nodes
	for _, node := range nodes {
		for _, route := range node.routes {
			printRouteDebugInfo(route)
		}
		printSubRoutes(node)
	}
	fmt.Println("")
}

func printSubRoutes(node *Node) {
	if node.HasChildren() {
		for _, node := range node.nodes {
			for _, route := range node.routes {
				printRouteDebugInfo(route)
			}
			printSubRoutes(node)
		}
	}
}

func printRouteDebugInfo(route *Route) {
	desc := route.Info[RouteInfoKeyDescription]
	if desc == nil {
		desc = ""
	}
	fmt.Printf("%9s   %-50s %s\n", route.Method, route.PathFormat, desc)
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
