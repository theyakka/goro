package goro

import (
	"net/http"
)

type route struct {
	Method      string
	PathFormat  string
	HasWildcard bool
	Handler     http.Handler
}

type Router struct {
	routeCache     map[string]route
	registedRoutes map[string][]routes
	variables      map[string]interface{}

	// RouteFilters - the registered route filters
	RouteFilters []Filter
}

func New() Router {
	router := Router{}
	router.RouteFilters = make(Filter, 0)
	router.registedRoutes = make()
	return router
}

// variable registration
func (r *Router) AddStringVar(variable string, value string) {
	r.variables[variable] = value
}

// route registration
// DELETE - Convenience func for a call using the http DELETE method
func (r *Router) DELETE(path string, handler http.Handler) {
}

// GET - Convenience func for a call using the http GET method
func (r *Router) GET(path string, handler http.Handler) {
}

// PATCH - Convenience func for a call using the http PATCH method
func (r *Router) PATCH(path string, handler http.Handler) {
}

// POST - Convenience func for a call using the http POST method
func (r *Router) POST(path string, handler http.Handler) {
}

// PUT - Convenience func for a call using the http PUT method
func (r *Router) PUT(path string, handler http.Handler) {
}

// path coalescing

func (r *Router) substituteVariables(pathFormat string) string {
	return pathFormat
}

func (r *Router) parseWildcards(pathFormat string, path string) []string {

}
