package goro

import (
	"net/http"
)

// Context helpers

// ParamsMap - returns the parameter map or an empty map
func ParamsMap(context *Context, req *http.Request) map[string]interface{} {
	outMap := map[string]interface{}{}
	contextVal := context.Get(req, ContextKeyRouteParams)
	if contextVal != nil {
		outMap = contextVal.(map[string]interface{})
	}
	return outMap
}

// Common helpers

// MinInt - Get the minimum integer value
func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}
