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
	"context"
	"net/url"
)

// ParamsFromContext - get the current params value from a given context
func ParamsFromContext(ctx context.Context) url.Values {
	params := ctx.Value(ParametersContextKey)
	if params != nil {
		return url.Values(params.(map[string][]string))
	}
	return nil
}

// ParamsWithoutID - returns copy of params with ID removed.
func ParamsWithoutID(params url.Values) url.Values {
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

// FirstStringParam - return the first item in the array if it exists, otherwise return
// an empty string
func FirstStringParam(params []string) string {
	if params != nil && len(params) > 0 {
		return params[0]
	}
	return ""
}

// FirstParam - return the first item in the array if it exists, otherwise return nil
func FirstParam(params []interface{}) interface{} {
	if params != nil && len(params) > 0 {
		return params[0]
	}
	return nil
}
