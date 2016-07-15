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
func (ch *Chain) Add(v ...http.Handler) *Chain {
	ch.Handlers = append(ch.Handlers, v...)
	return ch
}

// AddFunc adds one or more HandlerFuncs to the end of the chain
func (ch *Chain) AddFunc(v ...http.HandlerFunc) *Chain {
	for _, hfunc := range v {
		ch.Handlers = append(ch.Handlers, http.Handler(hfunc))
	}
	return ch
}

// Then calls the chain and then the designated Handler
func (ch *Chain) Then(handler http.Handler) *Chain {
	chain := &Chain{
		Handlers: ch.Handlers,
	}
	chain.Add(handler)
	return chain
}

// ThenFunc calls the chain and then the designated HandlerFunc
func (ch *Chain) ThenFunc(handlerFunc http.HandlerFunc) *Chain {
	return ch.Then(http.Handler(handlerFunc))
}

func (ch *Chain) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, handler := range ch.Handlers {
		handler.ServeHTTP(w, req)
	}
}
