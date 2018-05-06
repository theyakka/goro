// Goro
//
// Created by Yakka
// http://theyakka.com
//
// Copyright (c) 2018 Yakka LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// RouteParamsFromContext - get the current params value from a given context
func RouteParamsFromContext(ctx context.Context) url.Values {
	params := ctx.Value(ParametersContextKey)
	if params != nil {
		return url.Values(params.(map[string][]string))
	}
	return nil
}

// RouteParamsWithoutID - returns copy of params with ID removed.
func RouteParamsWithoutID(params url.Values) url.Values {
	if params != nil {
		paramsCopy := make(url.Values)
		for k, v := range params {
			paramsCopy[k] = v
		}
		delete(paramsCopy, "id")
		return paramsCopy
	}
	return nil
}

// FirstStringRouteParam - return the first item in the array if it exists, otherwise return
// an empty string
func FirstStringRouteParam(params []string) string {
	if params != nil && len(params) > 0 {
		return params[0]
	}
	return ""
}

// ErrorInfoForRequest - returns the error info for the request (if any)
func ErrorInfoForRequest(req *http.Request) ErrorMap {
	errInfo := req.Context().Value(ErrorValueContextKey)
	if errInfo != nil {
		return errInfo.(ErrorMap)
	}
	return nil
}

// CleanPath - returns a path value with the following modifications:
//	1. replaces any '\' with '/'
//	2. replaces any '//' with '/'
//	3. adds a leading '/' (if missing)
func CleanPath(path string) string {
	cleanPath := path
	// replace any non-unix path separators
	cleanPath = strings.Replace(cleanPath, "\\", "/", -1)
	// replace double separators
	cleanPath = strings.Replace(cleanPath, "//", "/", -1)
	// add leading slash if missing
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}
	return cleanPath
}
