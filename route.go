// route.go
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
	"net/http"
	"strings"
)

// RouteComponentType - route component types
// NOTE: variables will be stripped out / replaced so we dont track them
type RouteComponentType int

const (
	// ComponentTypeFixed - a fixed path component
	ComponentTypeFixed RouteComponentType = 1 << iota
	// ComponentTypeWildcard - a wildcard path component
	ComponentTypeWildcard
)

// routeComponent - stores information on route components
type routeComponent struct {
	Type            RouteComponentType
	Value           string
	WildcardMatches []Match
}

// Route - the primary struct to capture individual route information
type Route struct {
	Method         string
	PathFormat     string
	HasWildcards   bool
	Handler        http.Handler
	pathComponents []routeComponent
}

// NotFoundRoute - placeholder for when a route cannot be matched / found
func NotFoundRoute() *Route {
	return &Route{
		Method:     RouteNotFoundMethod,
		PathFormat: "",
	}
}

func splitRoutePathComponents(path string, wildcardMatches []Match) ([]routeComponent, error) {
	routeComponents := []routeComponent{}
	routeComponentStrings := strings.Split(path, "/")
	routeComponentStrings = routeComponentStrings[1:len(routeComponentStrings)]
	for _, component := range routeComponentStrings {
		componentType := ComponentTypeFixed
		if strings.HasPrefix(component, "{") {
			componentType = ComponentTypeWildcard
		} else if strings.HasPrefix(component, "{$") {
			return []routeComponent{}, errors.New("Encountered a variable. Variables should have been substituted already.")
		}
		addComponent := routeComponent{
			Type:            componentType,
			Value:           component,
			WildcardMatches: wildcardMatches,
		}
		routeComponents = append(routeComponents, addComponent)
	}
	return routeComponents, nil
}
