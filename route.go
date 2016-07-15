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
	"net/http"
	"strings"
)

const (
	// RouteInfoKeyIsRoot - does the route have wildcard parts
	RouteInfoKeyIsRoot string = "is_root"

	// RouteInfoKeyDescription - does the route have a catch all part
	RouteInfoKeyDescription string = "description"
)

// Route stores all the information about a route
type Route struct {
	Method     string
	Path       string
	PathFormat string
	Handler    http.Handler
	Meta       map[string]interface{}
	Info       map[string]interface{}
}

// NewRoute creates a new Route instance
func NewRoute(method string, path string) *Route {
	return NewRouteWithMeta(method, path, nil)
}

// NewRouteWithMeta creates a new Route instance with meta values
func NewRouteWithMeta(method string, path string, meta map[string]interface{}) *Route {
	upMethod := strings.ToUpper(method)
	routeMeta := meta
	if routeMeta == nil {
		routeMeta = map[string]interface{}{}
	}
	route := &Route{
		Method:     upMethod,
		PathFormat: path,
		Meta:       routeMeta,
	}
	info := map[string]interface{}{}
	if path == RootPath {
		info[RouteInfoKeyIsRoot] = true
	}
	route.Info = info
	return route
}

// Handle adds a ContextHandler to the Route
func (rte *Route) Handle(handler http.Handler) *Route {
	rte.Handler = handler
	return rte
}

// HandleFunc adds a wrapped ContextHandlerFunc (to ContextHandler) to the Route
func (rte *Route) HandleFunc(handlerFunc http.HandlerFunc) *Route {
	rte.Handler = http.Handler(handlerFunc)
	return rte
}

// Describe allows you to add a description of the route for other developers
func (rte *Route) Describe(description string) *Route {
	rte.Info[RouteInfoKeyDescription] = description
	return rte
}

// IsRoot returns true if the Route path is '/'
func (rte *Route) IsRoot() bool {
	return rte.Info[RouteInfoKeyIsRoot] == true
}
