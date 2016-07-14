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
)

// Chain allows for chaining of Handlers
type Chain struct {
	// Handlers - the handlers in the Chain
	Handlers []http.Handler
}

// NewChain creates a new Chain instance
func NewChain() Chain {
	return Chain{
		Handlers: []http.Handler{},
	}
}

// Add adds one or more Handlers to the end of the chain
func (ch Chain) Add(v ...http.Handler) {
	ch.Handlers = append(ch.Handlers, v...)
}

// AddFunc adds one or more HandlerFuncs to the end of the chain
func (ch Chain) AddFunc(v ...http.HandlerFunc) {
	for _, hfunc := range v {
		ch.Handlers = append(ch.Handlers, http.Handler(hfunc))
	}
}
