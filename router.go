package goro

import (
	"errors"
	"net/http"
	"strings"
)

type route struct {
	Method       string
	PathFormat   string
	HasWildcards bool
	Wildcards    []Match
	Handler      http.Handler
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
	ShouldRedirectTrailingSlash bool

	// NotFoundHandler - route / resource not found handler
	NotFoundHandler http.Handler

	// MethodNotAllowedHandler - if defined, will be hit wen requesting a defined route
	// via a non-defined http method (e.g.: requesting via POST when only GET is defined).
	// if not defined, we will fallback to the NotFoundHandler
	MethodNotAllowedHandler http.Handler

	// PanicHandler - handler for when things gets real
	PanicHandler http.Handler
}

func NewRouter() Router {
	return Router{
		routeCache:                  make(map[string]route),
		registedRoutes:              make(map[string][]route),
		variables:                   make(map[string]interface{}),
		ShouldRedirectTrailingSlash: true,
		RouteFilters:                []Filter{},
	}
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

	routePath := path
	wildcards, variables, wcErr := findWildcards(path)
	if wcErr != nil {
		// TODO - error
	}

	addRoute := route{
		Method:       method,
		PathFormat:   routePath,
		HasWildcards: len(wildcards) > 0,
		Wildcards:    wildcards,
		Handler:      handler,
	}

	if len(variables) > 0 {
		routePath = r.substituteVariables(addRoute, variables)
		addRoute.PathFormat = routePath
	}

	methodRoutes := r.registedRoutes[method]
	if methodRoutes == nil {
		methodRoutes = make([]route, 0)
	}
	methodRoutes = append(methodRoutes, addRoute)
	r.registedRoutes[method] = methodRoutes
}

// ServeHTTP -
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
}

// helper methods
// isWildCardPath - check if the path contains a wildcard portion
func isWildcardPath(path string) bool {
	return strings.Index(path, "{") != -1
}

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

func findWildcards(path string) (wildcards []Match, variables []Match, parseErr error) {
	if !strings.HasPrefix(path, "/") {
		// missing slash at the start, we aaaaare out
		return []Match{}, []Match{}, errors.New("Path is missing leading slash ('/')")
	}

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
