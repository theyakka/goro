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

// Matcher constants

// MatchTypeWildcard - match type representing a wildcard match
const MatchTypeWildcard string = "wildcard"

// MatchTypeVariable - match type representing a variable match
const MatchTypeVariable string = "variable"
