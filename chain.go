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

// ChainStatus - the status of the chain
type ChainStatus int

const (
	// ChainCompleted - the chain completed normally
	ChainCompleted ChainStatus = 1 << iota
	// ChainError - the chain was stopped because of an error
	ChainError
	// ChainHalted - the chain was halted before it could finish executing
	ChainHalted
)

// ChainResult - the chain execution result
type ChainResult struct {
	Status ChainStatus
	Error  error
}

// Chain allows for chaining of Handlers
type Chain struct {
	handlerIndex   int
	responseWriter http.ResponseWriter
	request        *http.Request

	// Handlers - the handlers in the Chain
	Handlers []ChainHandler

	resultCallbacks []ChainCompletedFunc

	// resultCompletedFunc - used internally when chain completes
	resultCompletedFunc ChainCompletedFunc
}

// ChainCompletedFunc - callback function executed when chain execution has
// completed
type ChainCompletedFunc func(result ChainResult)

// ChainHandler - Handler wrapper with access to the chain
type ChainHandler interface {
	Execute(chain Chain, w http.ResponseWriter, req *http.Request)
}

// ChainHandlerFunc - HandlerFunc wrapper with access to the chain
type ChainHandlerFunc func(chain Chain, w http.ResponseWriter, req *http.Request)

// Execute - execute the ChainHandlerFunc
func (chf ChainHandlerFunc) Execute(chain Chain, w http.ResponseWriter, req *http.Request) {
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

func (chw chainHandlerWrapper) Execute(chain Chain, w http.ResponseWriter, req *http.Request) {
	chw.handler.ServeHTTP(w, req)
}

// NewChain - creates a new Chain instance
func NewChain(handlers ...ChainHandler) Chain {
	return Chain{
		Handlers:        handlers,
		resultCallbacks: make([]ChainCompletedFunc, 0),
	}
}

// NewChainWithFuncs - creates a new Chain instance
func NewChainWithFuncs(handlers ...ChainHandlerFunc) Chain {
	allHandlers := make([]ChainHandler, 0, len(handlers))
	for _, hfunc := range handlers {
		allHandlers = append(allHandlers, ChainHandler(hfunc))
	}
	return Chain{
		Handlers:        allHandlers,
		resultCallbacks: make([]ChainCompletedFunc, 0),
	}
}

// Append - returns a new chain with the ChainHandler appended to
// the list of handlers
func (ch Chain) Append(handlers ...ChainHandler) Chain {
	allHandlers := make([]ChainHandler, 0, len(ch.Handlers)+len(handlers))
	allHandlers = append(allHandlers, ch.Handlers...)
	allHandlers = append(allHandlers, handlers...)
	return Chain{
		Handlers:        allHandlers,
		resultCallbacks: ch.resultCallbacks,
	}
}

// AppendFunc - returns a new chain with the ChainHandlerFunc appended to
// the list of handlers
func (ch Chain) AppendFunc(handlers ...ChainHandlerFunc) Chain {
	allHandlers := make([]ChainHandler, 0, len(ch.Handlers)+len(handlers))
	allHandlers = append(allHandlers, ch.Handlers...)
	for _, hfunc := range handlers {
		allHandlers = append(allHandlers, ChainHandler(hfunc))
	}
	return Chain{
		Handlers:        allHandlers,
		resultCallbacks: ch.resultCallbacks,
	}
}

func (ch Chain) AddResultCallback(callback ChainCompletedFunc) Chain {
	resultCallbacks := make([]ChainCompletedFunc, 0, len(ch.resultCallbacks)+1)
	resultCallbacks = append(resultCallbacks, callback)
	return Chain{
		Handlers:        ch.Handlers,
		resultCallbacks: resultCallbacks,
	}
}

// Then calls the chain and then the designated Handler
func (ch Chain) Then(handler http.Handler) Chain {
	wrapper := newChainHandlerWrapper(handler)
	return ch.Append(wrapper)
}

// ThenFunc calls the chain and then the designated HandlerFunc
func (ch Chain) ThenFunc(handlerFunc http.HandlerFunc) Chain {
	return ch.Then(http.Handler(handlerFunc))
}

// Next - execute the next handler in the chain
func (ch Chain) Next() {
	ch.handlerIndex++
	handlersCount := len(ch.Handlers)
	if ch.handlerIndex < handlersCount {
		ch.Handlers[ch.handlerIndex].Execute(ch, ch.responseWriter, ch.request)
	}
	if ch.handlerIndex == handlersCount-1 {
		result := ChainResult{Status: ChainCompleted, Error: nil}
		ch.resultCompletedFunc(result)
		ch.reset()
	}
}

// Halt - halt chain execution
func (ch Chain) Halt() {
	result := ChainResult{Status: ChainHalted, Error: nil}
	ch.resultCompletedFunc(result)
	ch.reset()
}

// Error - halt the chain and report an error
func (ch Chain) Error(chainError error) {
	result := ChainResult{Status: ChainError, Error: chainError}
	ch.resultCompletedFunc(result)
	ch.reset()
}

// reset - resets the chain
func (ch Chain) reset() {
	ch.handlerIndex = 0
}

// ServeHTTP - execute default functionality
func (ch Chain) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if len(ch.Handlers) > 0 {
		ch.reset()
		ch.responseWriter = w
		ch.request = req
		ch.Handlers[0].Execute(ch, w, req)
	}
}
