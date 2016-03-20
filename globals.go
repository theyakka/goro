// globals.go
// Goro
//
// Created by Posse in NYC
// http://goposse.com
//
// Copyright (c) 2016 Posse Productions LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

// Context constants

// ContextKeyRoutePathFormat - context value key for the original route path format value
const ContextKeyRoutePathFormat string = "goro.ctkRoutePathFmt"

// ContextKeyMatchedRoute - context value key for the matched route value
const ContextKeyMatchedRoute string = "goro.ctkMatchedRoute"

// ContextKeyRouteParams - context value key for the route parameter values
const ContextKeyRouteParams string = "goro.ctkRouteParams"

// ContextKeyCallWasCached - context value key flag determined if the last call came from the cache
const ContextKeyCallWasCached string = "goro.ctkCallWasCached"

// Route constants

// RouteNotFoundMethod - the method attached to the not found route
const RouteNotFoundMethod string = "NOTFOUND"

// Matcher constants

// MatchTypeWildcard - match type representing a wildcard match
const MatchTypeWildcard string = "wildcard"

// MatchTypeVariable - match type representing a variable match
const MatchTypeVariable string = "variable"
