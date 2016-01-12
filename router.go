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

// route - representation of a route
type route struct {
	Method       string
	PathFormat   string
	HasWildcards bool
	Components   []routeComponent
	Handler      http.HandlerFunc
}

func NotFoundRoute() route {
	return route{
		Method:     "NOTFOUND",
		PathFormat: "",
	}
}

// MatchesPath - check to see if the route matches the given path
func (r *route) MatchesPath(path string, checkSlash bool) (bool, map[string]interface{}) {
	// is exact match?
	params := map[string]interface{}{}
	if r.PathFormat == path {
		return true, params
	}

	pathMatches := true
	pathComps := strings.Split(path, "/")
	routeComps := r.Components
	deSlashedComps := []string{}
	if checkSlash && strings.HasSuffix(path, "/") {
		deSlashedPath := path[:len(path)-1]
		deSlashedComps = strings.Split(deSlashedPath, "/")
	}

	// fmt.Printf("path comps = %v, %s\n", pathComps, path)
	if len(pathComps) != len(routeComps) && len(deSlashedComps) != len(routeComps) {
		// fmt.Printf(" mismatch: %d, %d\n", len(pathComps), len(routeComps))
		return false, params
	}
	checkComps := pathComps
	if len(pathComps) != len(routeComps) {
		checkComps = deSlashedComps
	}

	for idx, comp := range routeComps {
		if comp.Type == ComponentTypeSegment {
			// fmt.Printf(" check: %s, %s\n", comp.Value, pathComps[idx])
			if comp.Value != checkComps[idx] {
				return false, params
			}
		} else if comp.Type == ComponentTypeWildcard {
			// TODO - parse wildcard format option
			params[comp.Value] = checkComps[idx]
		}
	}
	return pathMatches, params
}

// Router - the router definition
type Router struct {
	routeCache        map[string]route
	methodKeyedRoutes map[string][]route
	registeredRoutes  []route
	variables         map[string]interface{}

	// RouteFilters - the registered route filters
	RouteFilters []Filter

	// Context - used to store context during the router lifecycle
	Context ContextInterface

	// RedirectTrailingSlash - should we redirect a requested path with a trailing
	// slash to a defined route without the slash (if one exists)? Will use code 301
	// for GET and 307 otherwise
	ShouldRedirectTrailingSlash bool

	// NotFoundHandler - route / resource not found handler
	NotFoundHandler http.HandlerFunc

	// MethodNotAllowedHandler - if defined, will be hit wen requesting a defined route
	// via a non-defined http method (e.g.: requesting via POST when only GET is defined).
	// if not defined, we will fallback to the NotFoundHandler
	MethodNotAllowedHandler http.HandlerFunc

	// PanicHandler - handler for when things gets real
	PanicHandler http.HandlerFunc
}

func NewRouter() Router {
	return Router{
		routeCache:                  make(map[string]route),
		methodKeyedRoutes:           make(map[string][]route),
		registeredRoutes:            make([]route, 0),
		variables:                   make(map[string]interface{}),
		ShouldRedirectTrailingSlash: true,
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

func (r *Router) Route(method string, path string, handler http.HandlerFunc) error {
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
		r.PanicHandler(w, req)
	}
}

// ServeHTTP -
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	useReq := req
	usePath := useReq.URL.Path

	// execute the pre-process filters before we use the request / path
	if len(r.RouteFilters) > 0 {
		for _, filter := range r.RouteFilters {
			filter.ExecuteFilter(useReq, &usePath)
		}
	}

	if r.PanicHandler != nil {
		defer r.recoverError(w, useReq)
	}
	if r.Context != nil {
		defer r.Context.Clear(useReq)
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
	matchedRoute := NotFoundRoute()
	for _, route := range routes {
		doesMatch, params = route.MatchesPath(usePath, r.ShouldRedirectTrailingSlash)
		if doesMatch {
			if route.Method != useReq.Method {
				r.MethodNotAllowedHandler(w, useReq)
				return
			}
			matchedRoute = route
			break
		}
	}

	if r.NotFoundHandler != nil && matchedRoute.PathFormat == "" {
		r.NotFoundHandler(w, useReq)
		return
	}

	if r.Context != nil {
		r.Context.Put(useReq, "matched_route", matchedRoute)
		for k, v := range params {
			r.Context.Put(useReq, stripTokenDelims(k), v)
		}
	}

	matchedRoute.Handler(w, useReq)

	// fmt.Printf("path = %s", usePath)
	// fmt.Printf("\n")
	// fmt.Printf("Ya gotta serve somebody\n")
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
