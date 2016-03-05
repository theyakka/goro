package goro

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type RouteComponentType int

const (
	ComponentTypeSegment RouteComponentType = 1 << iota
	ComponentTypeWildcard
)

type routeComponent struct {
	Type      RouteComponentType
	Value     string
	Wildcards []Match
}

type cacheEntry struct {
	Params  map[string]interface{}
	Handler http.Handler
	Route   route
}

// route - representation of a route
type route struct {
	Method       string
	PathFormat   string
	HasWildcards bool
	Components   []routeComponent
	Handler      http.Handler
}

func NotFoundRoute() route {
	return route{
		Method:     "NOTFOUND",
		PathFormat: "",
	}
}

// MatchesPath - check to see if the route matches the given path
func (r *route) MatchesPath(path string, checkSlash bool) (bool, map[string]interface{}, int) {
	// is exact match?
	params := map[string]interface{}{}
	if r.PathFormat == path {
		return true, params, http.StatusOK
	}

	redirectOnMatch := false
	pathMatches := true
	pathComps := strings.Split(path, "/")
	routeComps := r.Components
	deSlashedComps := []string{}
	if checkSlash && path != "/" && strings.HasSuffix(path, "/") {
		deSlashedPath := path[:len(path)-1]
		deSlashedComps = strings.Split(deSlashedPath, "/")
		redirectOnMatch = true
	}

	// fmt.Printf("path comps = %v, %s\n", pathComps, path)
	if len(pathComps) != len(routeComps) && len(deSlashedComps) != len(routeComps) {
		// fmt.Printf(" mismatch: %d, %d\n", len(pathComps), len(routeComps))
		return false, params, http.StatusNotFound
	}
	checkComps := pathComps
	if len(pathComps) != len(routeComps) {
		checkComps = deSlashedComps
	}

	for idx, comp := range routeComps {
		if comp.Type == ComponentTypeSegment {
			// fmt.Printf(" check: %s, %s\n", comp.Value, pathComps[idx])
			if comp.Value != checkComps[idx] {
				return false, params, http.StatusNotFound
			}
		} else if comp.Type == ComponentTypeWildcard {
			// TODO - parse wildcard format option
			params[comp.Value] = checkComps[idx]
		}
	}

	code := http.StatusOK
	if redirectOnMatch {
		code = http.StatusMovedPermanently
		if r.Method != "GET" {
			code = http.StatusTemporaryRedirect
		}
	}

	return pathMatches, params, code
}

// Router - the router definition
type Router struct {
	routeCache        map[string]cacheEntry
	methodKeyedRoutes map[string][]route
	registeredRoutes  []route
	variables         map[string]interface{}

	// RedirectTrailingSlash - should we redirect a requested path with a trailing
	// slash to a defined route without the slash (if one exists)? Will use code 301
	// for GET and 307 otherwise
	ShouldRedirectTrailingSlash bool

	// ShouldCacheMatchedRoutes - should we cache a route after it has been matched?
	ShouldCacheMatchedRoutes bool

	// RouteFilters - the registered route filters
	RouteFilters []Filter

	// Context - used to store context during the router lifecycle
	Context ContextInterface

	// NotFoundHandler - route / resource not found handler
	NotFoundHandler http.Handler

	// MethodNotAllowedHandler - if defined, will be hit wen requesting a defined route
	// via a non-defined http method (e.g.: requesting via POST when only GET is defined).
	// if not defined, we will fallback to the NotFoundHandler
	MethodNotAllowedHandler http.Handler

	// PanicHandler - handler for when things gets real
	PanicHandler http.Handler
}

func NewRouter() *Router {
	return &Router{
		routeCache:                  make(map[string]cacheEntry),
		methodKeyedRoutes:           make(map[string][]route),
		registeredRoutes:            make([]route, 0),
		variables:                   make(map[string]interface{}),
		ShouldRedirectTrailingSlash: true,
		ShouldCacheMatchedRoutes:    true,
		RouteFilters:                []Filter{},
	}
}

// variable registration
func (r *Router) AddStringVar(variable string, value string) {
	wrappedVarName := value
	if strings.HasPrefix(variable, "$") {
		wrappedVarName = "{" + variable + "}"
	} else if !strings.HasPrefix(variable, "{$") {
		wrappedVarName = "{$" + variable + "}"
	}
	r.variables[wrappedVarName] = value
}

// route registration
// DELETE - Convenience func for a call using the http DELETE method
func (r *Router) DELETE(path string, handler http.HandlerFunc) {
	r.Route("DELETE", path, handler)
}

// GET - Convenience func for a call using the http GET method
func (r *Router) GET(path string, handler http.HandlerFunc) {
	r.Route("GET", path, handler)
}

// PATCH - Convenience func for a call using the http PATCH method
func (r *Router) PATCH(path string, handler http.HandlerFunc) {
	r.Route("PATCH", path, handler)
}

// POST - Convenience func for a call using the http POST method
func (r *Router) POST(path string, handler http.HandlerFunc) {
	r.Route("POST", path, handler)
}

// PUT - Convenience func for a call using the http PUT method
func (r *Router) PUT(path string, handler http.HandlerFunc) {
	r.Route("PUT", path, handler)
}

func (r *Router) Route(method string, path string, handler http.Handler) error {
	if !strings.HasPrefix(path, "/") {
		// missing slash at the start, we aaaaare out
		return errors.New("Path is missing leading slash ('/')")
	}

	routePath := path
	wildcards, variables, wcErr := findWildcards(path)
	if wcErr != nil {
		// TODO - error
	}

	addRoute := route{
		Method:       method,
		PathFormat:   routePath,
		HasWildcards: len(wildcards) > 0,
		Handler:      handler,
	}

	if len(variables) > 0 {
		routePath = r.substituteVariables(addRoute, variables)
		addRoute.PathFormat = routePath
	}

	routeComponents := make([]routeComponent, 0)
	routeCompStrings := strings.Split(routePath, "/")
	for _, comp := range routeCompStrings {
		compType := ComponentTypeSegment
		if strings.HasPrefix(comp, "{") {
			compType = ComponentTypeWildcard
		}
		rcomp := routeComponent{
			Type:      compType,
			Value:     comp,
			Wildcards: wildcards,
		}
		routeComponents = append(routeComponents, rcomp)
	}

	addRoute.Components = routeComponents

	methodRoutes := r.methodKeyedRoutes[method]
	if methodRoutes == nil {
		methodRoutes = make([]route, 0)
	}
	methodRoutes = append(methodRoutes, addRoute)
	r.methodKeyedRoutes[strings.ToUpper(method)] = methodRoutes
	r.registeredRoutes = append(r.registeredRoutes, addRoute)

	return nil
}

// recoverError - recover from any errors and call the panic handler
func (r *Router) recoverError(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler.ServeHTTP(w, req)
	}
}

// ServeHTTP -
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	useReq := req
	usePath := useReq.URL.Path

	if r.PanicHandler != nil {
		defer r.recoverError(w, useReq)
	}
	if r.Context != nil {
		defer r.Context.Clear(useReq)
	}

	// execute the pre-process filters before we use the request / path
	if len(r.RouteFilters) > 0 {
		for _, filter := range r.RouteFilters {
			filter.ExecuteFilter(useReq, &usePath, r.Context)
		}
	}

	var cacheKey string
	if r.ShouldCacheMatchedRoutes {
		cacheKey = fmt.Sprintf("%s:%s", req.Method, usePath)
		if cachedRoute := r.routeCache[cacheKey]; cachedRoute.Handler != nil {
			if r.Context != nil {
				r.Context.Put(useReq, ContextKeyRoutePathFormat, cachedRoute.Route.PathFormat)
				r.Context.Put(useReq, ContextKeyMatchedRoute, cachedRoute)
				for k, v := range cachedRoute.Params {
					r.Context.Put(useReq, stripTokenDelims(k), v)
				}
			}
			cachedRoute.Handler.ServeHTTP(w, useReq)
			return
		}
	}

	method := strings.ToUpper(useReq.Method)
	var routes []route

	// if the method not allowed handler IS set, we should check ALL routes
	// if the handler is NOT set, we can just limit it to the routes for this
	// particular method to make things faster
	if r.MethodNotAllowedHandler != nil {
		routes = r.registeredRoutes
	} else {
		routes = r.methodKeyedRoutes[method]
	}

	var doesMatch bool
	var params map[string]interface{}
	var matchCode int = http.StatusOK

	nonMethodMatched := false
	matchedRoute := NotFoundRoute()
	for _, route := range routes {
		doesMatch, params, matchCode = route.MatchesPath(usePath, r.ShouldRedirectTrailingSlash)
		if doesMatch {
			if matchCode == http.StatusMovedPermanently || matchCode == http.StatusTemporaryRedirect &&
				r.ShouldRedirectTrailingSlash {
				req.URL.Path = usePath[:len(usePath)-1]
				http.Redirect(w, req, req.URL.String(), matchCode)
				return
			}
			nonMethodMatched = (route.Method != useReq.Method)
			if !nonMethodMatched {
				nonMethodMatched = false
				matchedRoute = route
				break
			}
		}
	}

	if nonMethodMatched {
		r.MethodNotAllowedHandler.ServeHTTP(w, useReq)
		return
	}

	if matchedRoute.PathFormat == "" {
		if r.NotFoundHandler != nil {
			r.NotFoundHandler.ServeHTTP(w, useReq)
			return
		} else {
			http.NotFound(w, useReq)
			return
		}
	}

	if r.Context != nil {
		r.Context.Put(useReq, ContextKeyRoutePathFormat, matchedRoute.PathFormat)
		r.Context.Put(useReq, ContextKeyMatchedRoute, matchedRoute)
		for k, v := range params {
			r.Context.Put(useReq, stripTokenDelims(k), v)
		}
	}

	matchedRoute.Handler.ServeHTTP(w, useReq)

	if r.ShouldCacheMatchedRoutes {
		// cache the route
		r.routeCache[cacheKey] = cacheEntry{
			Params:  params,
			Handler: matchedRoute.Handler,
			Route:   matchedRoute,
		}
	}
}

// substituteVariables - swap any variable parts in the path format for their
// defined values
func (r *Router) substituteVariables(rte route, variables []Match) string {
	// iterate the variables in reverse order so that we match back to front
	// this will preserve the match ranges compared to the modified final path
	finalPath := rte.PathFormat
	for _, variable := range variables {
		variableVal := r.variables[variable.Value]
		if variableVal != nil {
			finalPath = strings.Replace(finalPath, variable.Value, variableVal.(string), -1)
		}
	}
	return finalPath
}

// PrintRoutes - prints debugging information about all the registered routes
func (r *Router) PrintRoutes() {
	for k, routes := range r.methodKeyedRoutes {
		fmt.Printf("Method: %s\n", k)
		for _, match := range routes {
			fmt.Printf("  - %s\n", match.PathFormat)
		}
		fmt.Println("")
	}
}

// helper methods
// isWildCardPath - check if the path contains a wildcard portion
func isWildcardPath(path string) bool {
	return strings.Index(path, "{") != -1
}

func findWildcards(path string) (wildcards []Match, variables []Match, parseErr error) {
	hasWildcard := (strings.Index(path, "{") != -1)
	if !hasWildcard {
		// no wildcards, return now
		return []Match{}, []Match{}, nil
	}

	wildcardMatches := make([]Match, 0)
	variableMatches := make([]Match, 0)
	matcher := NewMatcher(path, "{", "}")

	match := matcher.NextMatch()
	for match != NotFoundMatch() {
		if match.Type == "wildcard" {
			wildcardMatches = append(wildcardMatches, match)
		} else {
			variableMatches = append(variableMatches, match)
		}
		match = matcher.NextMatch()
	}
	return wildcardMatches, variableMatches, nil
}

func stripTokenDelims(value string) string {
	replacer := strings.NewReplacer("{", "", "}", "")
	return replacer.Replace(value)
}
