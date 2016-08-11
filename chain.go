// Goro
//
// Created by Posse in NYC
// http://goposse.com
//
// Copyright (c) 2016 Posse Productions LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

import "net/http"

// Chain allows for chaining of Handlers
type Chain struct {
	// Handlers - the handlers in the Chain
	Handlers       []ChainHandler
	handlerIndex   int
	responseWriter http.ResponseWriter
	request        *http.Request
}

// ChainHandler - Handler wrapper with access to the chain
type ChainHandler interface {
	Execute(chain *Chain, w http.ResponseWriter, req *http.Request)
}

// ChainHandlerFunc - HandlerFunc wrapper with access to the chain
type ChainHandlerFunc func(chain *Chain, w http.ResponseWriter, req *http.Request)

// Execute - execute the ChainHandlerFunc
func (chf ChainHandlerFunc) Execute(chain *Chain, w http.ResponseWriter, req *http.Request) {
	chf(chain, w, req)
}

// chainHandlerWrapper - wraps a regular http.Handler for chaining purposes
type chainHandlerWrapper struct {
	handler http.Handler
}

// newChainHandlerWrapper - creates a new chain handler wrapper
func newChainHandlerWrapper(handler http.Handler) chainHandlerWrapper {
	return chainHandlerWrapper{
		handler: handler,
	}
}

func (chw chainHandlerWrapper) Execute(chain *Chain, w http.ResponseWriter, req *http.Request) {
	chw.handler.ServeHTTP(w, req)
}

// NewChain creates a new Chain instance
func NewChain(handlers ...ChainHandler) *Chain {
	return &Chain{
		Handlers: handlers,
	}
}

// Add adds one or more Handlers to the end of the chain
func (ch *Chain) Add(v ...ChainHandler) *Chain {
	ch.Handlers = append(ch.Handlers, v...)
	return ch
}

// AddFunc adds one or more HandlerFuncs to the end of the chain
func (ch *Chain) AddFunc(v ...ChainHandlerFunc) *Chain {
	for _, hfunc := range v {
		ch.Handlers = append(ch.Handlers, ChainHandler(hfunc))
	}
	return ch
}

// Then calls the chain and then the designated Handler
func (ch *Chain) Then(handler http.Handler) *Chain {
	chain := &Chain{
		Handlers:     ch.Handlers,
		handlerIndex: 0,
	}
	wrapper := newChainHandlerWrapper(handler)
	chain.Add(wrapper)
	return chain
}

// ThenFunc calls the chain and then the designated HandlerFunc
func (ch *Chain) ThenFunc(handlerFunc http.HandlerFunc) *Chain {
	return ch.Then(http.Handler(handlerFunc))
}

// Next - execute the next handler in the chain
func (ch *Chain) Next() {
	ch.handlerIndex++
	if ch.handlerIndex < len(ch.Handlers) {
		ch.Handlers[ch.handlerIndex].Execute(ch, ch.responseWriter, ch.request)
	}
}

// Halt - halt chain execution
func (ch *Chain) Halt() {
	ch.reset()
}

// reset - resets the chain
func (ch *Chain) reset() {
	ch.handlerIndex = 0
}

// ServeHTTP - execute default functionality
func (ch *Chain) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if len(ch.Handlers) > 0 {
		ch.reset()
		ch.responseWriter = w
		ch.request = req
		ch.Handlers[0].Execute(ch, w, req)
	}
}
