package goro

import (
	"net/http"
)

// Filter is an interface that can be registered on the Router to apply custom
// logic and pass-thru the route information
type Filter interface {
	// ExecuteFilter allows for rewriting/modification of the original request and/or
	// resulting path
	ExecuteFilter(req *http.Request, path *string, ctx *IContext)
}
