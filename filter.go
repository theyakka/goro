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
	"net/http"
)

// Filter is an interface that can be registered on the Router to apply custom
// logic to modify the Request or calling Context
type Filter interface {
	ExecuteFilter(ctx context.Context, req *http.Request) context.Context
}
