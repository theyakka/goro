package goro

import (
	"errors"
	"net/http"
	"strings"
)

type route struct {
	Method      string
	PathFormat  string
	HasWildcard bool
	Handler     http.Handler
}

type Router struct {
	routeCache     map[string]route
	registedRoutes map[string][]route
	variables      map[string]interface{}

	// RouteFilters - the registered route filters
	RouteFilters []Filter

	// Context - used to store context during the router lifecycle
	Context ContextInterface

	// RedirectTrailingSlash - should we redirect a requested path with a trailing
	// slash to a defined route without the slash (if one exists)? Will use code 301
	// for GET and 307 otherwise
	RedirectTrailingSlash bool
}

func NewRouter() Router {
	router := Router{}
	router.RouteFilters = []Filter{}
	router.routeCache = make(map[string]route)
	router.registedRoutes = make(map[string][]route)
	router.variables = make(map[string]interface{})
	return router
}

// variable registration
func (r *Router) AddStringVar(variable string, value string) {
	r.variables[variable] = value
}

// route registration
// DELETE - Convenience func for a call using the http DELETE method
func (r *Router) DELETE(path string, handler http.Handler) {
	r.Route("DELETE", path, handler)
}

// GET - Convenience func for a call using the http GET method
func (r *Router) GET(path string, handler http.Handler) {
	r.Route("GET", path, handler)
}

// PATCH - Convenience func for a call using the http PATCH method
func (r *Router) PATCH(path string, handler http.Handler) {
	r.Route("PATCH", path, handler)
}

// POST - Convenience func for a call using the http POST method
func (r *Router) POST(path string, handler http.Handler) {
	r.Route("POST", path, handler)
}

// PUT - Convenience func for a call using the http PUT method
func (r *Router) PUT(path string, handler http.Handler) {
	r.Route("PUT", path, handler)
}

func (r *Router) Route(method string, path string, handler http.Handler) {

	// hasWildcard := isWildcardPath(path)
	// addRoute := route{
	// 	PathFormat: path, Method: method,
	// 	HasWildcard: hasWildcard, Handler: handler,
	// }

}

// ServeHTTP -
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
}

// helper methods
// isWildCardPath - check if the path contains a wildcard portion
func isWildcardPath(path string) bool {
	return strings.Index(path, "{") != -1
}

func (r *route) substituteVariables(variables map[string]interface{}) {

}

func parsePath(path string) (finalPath string, wildcards []Match, parseErr error) {
	if !strings.HasPrefix(path, "/") {
		// missing slash at the start, we aaaaare out
		return "", []Match{}, errors.New("Path is missing leading slash ('/')")
	}

	hasWildcard := (strings.Index(path, "{") != -1)
	if !hasWildcard {
		// no wildcards, return now
		return path, []Match{}, nil
	}

	outPath := path
	wildcardMatches := make([]Match, 0)
	matcher := NewMatcher(outPath, "{", "}")

	match := matcher.NextMatch()
	for match != NotFoundMatch() {
		wildcardMatches = append(wildcardMatches, match)
		match = matcher.NextMatch()
	}
	return outPath, wildcardMatches, nil
}
