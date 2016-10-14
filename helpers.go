package goro

import (
	"context"
	"net/url"
)

// ParamsFromContext - get the current params value from a given context
func ParamsFromContext(ctx context.Context) url.Values {
	params := ctx.Value("params")
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
