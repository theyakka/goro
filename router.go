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

	// registeredRoutes - all routes registered with the router
	registeredRoutes []Route

	// registeredRoutes - all routes registered with the router
	methodKeyedRoutes map[string][]Route

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
		methodKeyedRoutes:           map[string][]Route{},
		registeredRoutes:            []Route{},
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

// DELETE - Convenience func for a call using the http DELETE method
func (r *Router) DELETE(path string, handler http.HandlerFunc) {
	r.AddRoute("DELETE", path, handler)
}

// GET - Convenience func for a call using the http GET method
func (r *Router) GET(path string, handler http.HandlerFunc) {
	r.AddRoute("GET", path, handler)
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

	route := Route{
		Method:         strings.ToUpper(method),
		PathFormat:     pathToUse,
		HasWildcards:   len(wildcards) > 0,
		Handler:        handler,
		pathComponents: components,
	}

	// append the route to the ordered list of routes
	r.registeredRoutes = append(r.registeredRoutes, route)

	// append the route to the HTTP method keyed map of routes
	methodRoutes := r.methodKeyedRoutes[route.Method]
	if methodRoutes == nil {
		methodRoutes = make([]Route, 0)
	}
	methodRoutes = append(methodRoutes, route)
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

// ServeHTTP - where the magic happens
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
}

// Debugging functions

// PrintRoutes - print all registered routes
func (r *Router) PrintRoutes() {
	fmt.Println("REGISTERED ROUTES --")
	for _, route := range r.registeredRoutes {
		fmt.Printf("- [%s] %s\n", route.Method, route.PathFormat)
	}
}
