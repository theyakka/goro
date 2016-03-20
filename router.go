// router.go
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
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Router - the primary router type
type Router struct {

	// routeCache - storage for cached routes
	routeCache *RouteCache

	// routeMatcher - the primary route matcher instance
	routeMatcher *routeMatcher

	// variables - unwrapped (clean) variables that have been defined
	variables map[string]string

	// methodKeyedRoutes - all routes registered with the router
	methodKeyedRoutes map[string]*Tree

	allowedPathMethods map[string][]string

	// wrappedVariables - wrapped ({$varname}) versions of the variables
	wrappedVariables map[string]string

	// RedirectTrailingSlash - should we redirect a requested path with a trailing
	// slash to a defined route without the slash (if one exists)? Will use code 301
	// for GET and 307 otherwise
	ShouldRedirectTrailingSlash bool

	// ShouldCacheMatchedRoutes - should we cache a route after it has been matched?
	ShouldCacheMatchedRoutes bool

	// RouteFilters - the registered route filters
	RouteFilters []Filter

	// Context - used to store context during the router lifecycle
	Context IContext

	// NotFoundHandler - route / resource not found handler
	NotFoundHandler http.Handler

	// MethodNotAllowedHandler - if defined, will be hit wen requesting a defined route
	// via a non-defined http method (e.g.: requesting via POST when only GET is defined).
	// if not defined, we will fallback to the NotFoundHandler
	MethodNotAllowedHandler http.Handler

	// PanicHandler - handler for when things gets real
	PanicHandler http.Handler

	// GlobalHandlers - any global HTTP method-based handlers
	globalHandlers map[string]http.Handler
}

// NewRouter - creates a new default instance of the Router type
func NewRouter() *Router {
	return &Router{
		routeCache:                  NewRouteCache(),
		ShouldRedirectTrailingSlash: true,
		ShouldCacheMatchedRoutes:    true,
		RouteFilters:                []Filter{},
		variables:                   map[string]string{},
		wrappedVariables:            map[string]string{},
		methodKeyedRoutes:           map[string]*Tree{},
		globalHandlers:              map[string]http.Handler{},
		allowedPathMethods:          map[string][]string{},
	}
}

// Variable functions

// SetVariable - registers a variable value for route variable matching
func (r *Router) SetVariable(variableKey string, stringValue string) {
	// check to see if the variable has already been wrapped, if not, wrap
	// it so we don't need to dynamically do it later
	wrappedVarName := stringValue
	if strings.HasPrefix(variableKey, "$") {
		wrappedVarName = "{" + variableKey + "}"
	} else if !strings.HasPrefix(variableKey, "{$") {
		wrappedVarName = "{$" + variableKey + "}"
	}
	r.wrappedVariables[wrappedVarName] = stringValue
}

// ResetVariables - reset all variables (to empty)
func (r *Router) ResetVariables() {
	r.variables = map[string]string{}
	r.wrappedVariables = map[string]string{}
}

// Routing functions

// AddGlobalHandler - registers a global handler for all requests made using
// 									  the provided HTTP method
func (r *Router) AddGlobalHandler(method string, handler http.Handler) {
	r.globalHandlers[strings.ToUpper(method)] = handler
}

// DELETE - Convenience func for a call using the http DELETE method
func (r *Router) DELETE(path string, handler http.HandlerFunc) {
	r.AddRoute("DELETE", path, handler)
}

// GET - Convenience func for a call using the http GET method
func (r *Router) GET(path string, handler http.HandlerFunc) {
	r.AddRoute("GET", path, handler)
}

// HEAD - Convenience func for a call using the http HEAD method
func (r *Router) HEAD(path string, handler http.HandlerFunc) {
	r.AddRoute("HEAD", path, handler)
}

// OPTIONS - Convenience func for a call using the http OPTIONS method
func (r *Router) OPTIONS(path string, handler http.HandlerFunc) {
	r.AddRoute("OPTIONS", path, handler)
}

// PATCH - Convenience func for a call using the http PATCH method
func (r *Router) PATCH(path string, handler http.HandlerFunc) {
	r.AddRoute("PATCH", path, handler)
}

// POST - Convenience func for a call using the http POST method
func (r *Router) POST(path string, handler http.HandlerFunc) {
	r.AddRoute("POST", path, handler)
}

// PUT - Convenience func for a call using the http PUT method
func (r *Router) PUT(path string, handler http.HandlerFunc) {
	r.AddRoute("PUT", path, handler)
}

// AddRoute - generic interface for registering a route
func (r *Router) AddRoute(method string, path string, handler http.Handler) error {
	if !strings.HasPrefix(path, "/") {
		// missing slash at the start, we aaaaare out
		return errors.New("Path value is missing leading slash ('/')")
	}
	pathToUse := path
	// find all wildcards and variable matches
	wildcards, variables := findSpecialComponents(pathToUse)
	if len(variables) > 0 {
		// substitute out any variable values in the original path
		var variableError error
		pathToUse, variableError = r.substituteVariables(pathToUse, variables)
		if variableError != nil {
			return variableError
		}
	}
	// split the path into its components to assist matching later
	components, splitErr := splitRoutePathComponents(pathToUse, wildcards)
	if splitErr != nil {
		return splitErr
	}

	route := &Route{
		Method:         strings.ToUpper(method),
		PathFormat:     pathToUse,
		HasWildcards:   len(wildcards) > 0,
		Handler:        handler,
		pathComponents: components,
	}

	// append the method to the allowed methods for this path format
	allowedMethods := r.allowedMethodsForPath(pathToUse)
	allowedMethods = append(allowedMethods, method)
	r.allowedPathMethods[pathToUse] = allowedMethods

	// append the route to the HTTP method keyed map of routes
	methodRoutes := r.methodKeyedRoutes[route.Method]
	if methodRoutes == nil {
		methodRoutes = &Tree{
			nodes: []*Node{},
		}
	}
	methodRoutes.AddRoute(pathToUse, route)
	r.methodKeyedRoutes[route.Method] = methodRoutes
	return nil
}

// Route path helpers

// findSpecialComponents - find any variable or wildcard matches
func findSpecialComponents(pathString string) (wildcards []Match, variables []Match) {
	matchedWildcards := []Match{}
	matchedVariables := []Match{}
	matcher := NewMatcher(pathString, "{", "}")
	match := matcher.NextMatch()
	for match != NotFoundMatch() {
		if match.Type == MatchTypeWildcard {
			matchedWildcards = append(matchedWildcards, match)
		} else if match.Type == MatchTypeVariable {
			matchedVariables = append(matchedVariables, match)
		}
		match = matcher.NextMatch()
	}
	return matchedWildcards, matchedVariables
}

// substituteVariables - swap any valid variables in a path string with their variable values
func (r *Router) substituteVariables(pathString string, variableMatches []Match) (string, error) {
	finalPath := pathString
	for _, match := range variableMatches {
		matchedVariable := r.wrappedVariables[match.Value]
		if matchedVariable != "" {
			finalPath = strings.Replace(finalPath, match.Value, matchedVariable, -1)
		} else {
			return "", fmt.Errorf("Found no registered value for the variable '%s'", match.Value)
		}
	}
	return finalPath, nil
}

// stripTokenDelimiters - remove any token delimiter instances from the string
func stripTokenDelimiters(value string) string {
	replacer := strings.NewReplacer("{", "", "}", "")
	return replacer.Replace(value)
}

// Route matching

func (r *Router) allowedMethodsForPath(path string) []string {
	allowedMethods := r.allowedPathMethods[path]
	if allowedMethods == nil {
		allowedMethods = []string{}
	}
	return allowedMethods
}

// findMatchingRoute - find the matching route (if registered) that
func (r *Router) findMatchingRoute(path string, method string, checkCache bool) (route *Route, params map[string]interface{}, wasCached bool, matchErrCode int) {
	if checkCache {
		cacheEntry := r.routeCache.Get(path)
		if cacheEntry.hasValue {
			// got a cached route
			return cacheEntry.Route, cacheEntry.Params, true, 0
		}
	}
	routesTree := r.methodKeyedRoutes[method]
	methodHasRoutes := (len(routesTree.nodes) > 0)
	if methodHasRoutes {
		// search for a matching route
		tree := r.methodKeyedRoutes[strings.ToUpper(method)]
		route, params := tree.RouteForPath(path)
		didMatchRoute := (route != nil)
		if didMatchRoute {
			return route, params, false, 0
		}
	}

	// didn't match route
	emptyParams := map[string]interface{}{}
	allowedMethods := r.allowedMethodsForPath(path)
	if len(allowedMethods) > 0 {
		// method not allowed because we couldn't match a route
		return nil, emptyParams, false, http.StatusMethodNotAllowed
	}
	// no allowed methods for the path so not found
	return nil, emptyParams, false, http.StatusNotFound
}

// Handler / content serving functions

// ServeHTTP - where the magic happens
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	useReq := req
	usePath := useReq.URL.Path
	if r.PanicHandler != nil {
		// defer errors to the panic handler
		defer r.recoverError(w, useReq)
	}
	if r.Context != nil {
		// defer clearing the context until finished serving
		defer r.Context.Clear(useReq)
	}
	// execute the pre-process filters before we use the request / path
	if len(r.RouteFilters) > 0 {
		for _, filter := range r.RouteFilters {
			filter.ExecuteFilter(useReq, &usePath, &r.Context)
		}
	}
	// check to see if there is a registered global handler for the request's
	// HTTP method
	if len(r.globalHandlers) > 0 {
		globalHandler := r.globalHandlers[strings.ToUpper(req.Method)]
		if globalHandler != nil {
			globalHandler.ServeHTTP(w, req)
			return
		}
	}

	route, params, wasCached, matchResultCode := r.findMatchingRoute(usePath, req.Method, r.ShouldCacheMatchedRoutes)
	if matchResultCode == 0 {
		if r.ShouldCacheMatchedRoutes {
			cacheEntry := CacheEntry{
				hasValue: true,
				Route:    route,
				Params:   params,
			}
			r.routeCache.Put(usePath, cacheEntry)
		}
		r.setRequestContextVariables(req, route, params, wasCached)
		route.Handler.ServeHTTP(w, req)
		return
	}

	// there was an error matching the route
	if matchResultCode == http.StatusMethodNotAllowed {
		if r.MethodNotAllowedHandler != nil {
			r.MethodNotAllowedHandler.ServeHTTP(w, req)
		} else {
			http.Error(w, "Method not allowed", matchResultCode)
		}
	} else if matchResultCode == http.StatusNotFound {
		if r.NotFoundHandler != nil {
			r.NotFoundHandler.ServeHTTP(w, req)
		} else {
			http.NotFound(w, req)
		}
	} else {
		http.Error(w, "Error ocurred", matchResultCode)
	}
}

// recoverError - recover from any errors and call the panic handler
func (r *Router) recoverError(w http.ResponseWriter, req *http.Request) {
	if panic := recover(); panic != nil {
		r.PanicHandler.ServeHTTP(w, req)
		return
	}
}

// Context functions

func (r *Router) setRequestContextVariables(req *http.Request, route *Route, params map[string]interface{}, wasCached bool) {
	if r.Context != nil {
		r.Context.Put(req, ContextKeyRoutePathFormat, route.PathFormat)
		r.Context.Put(req, ContextKeyMatchedRoute, route)
		r.Context.Put(req, ContextKeyRouteParams, params)
		r.Context.Put(req, ContextKeyCallWasCached, wasCached)
	}
}

// Debugging functions

// printNodeRoutes - recursively print out all routes
func printNodeRoutes(nodes []*Node) {
	for _, node := range nodes {
		if node.route != nil {
			fmt.Printf("%7s: %s\n", node.route.Method, node.route.PathFormat)
		}
		if len(node.nodes) > 0 {
			printNodeRoutes(node.nodes)
		}
	}
}

// PrintRoutes - print all registered routes
func (r *Router) PrintRoutes() {
	fmt.Println("REGISTERED ROUTES --")
	for _, routeTree := range r.methodKeyedRoutes {
		printNodeRoutes(routeTree.nodes)
	}
	if len(r.globalHandlers) > 0 {
		fmt.Println("\nGlobal handler(s) registered for:")
		for method := range r.globalHandlers {
			fmt.Printf("  %s\n", method)
		}
	}
	fmt.Println()
}

func printNodesRecursively(nodes []*Node, level int) {
	var levelString string
	var i = 0
	for i = 0; i < level; i++ {
		levelString += "-"
	}
	if len(levelString) > 0 {
		levelString += " "
	}
	for _, node := range nodes {
		fmt.Printf("  %s%s\n", levelString, node.part)
		printNodesRecursively(node.nodes, level+1)
	}
}

// PrintTrees - print all the trees
func (r *Router) PrintTrees() {
	fmt.Println("REGISTERED TREES --")
	for method, routeTree := range r.methodKeyedRoutes {
		fmt.Println(method)
		printNodesRecursively(routeTree.nodes, 0)
	}
}
