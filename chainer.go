// chainer.go
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

// ChainedHandler - chainer handler type
type ChainedHandler func(w http.ResponseWriter, r *http.Request) (int, error)

// ChainerErrorHandler - global error handler
type ChainerErrorHandler func(w http.ResponseWriter, r *http.Request, status int, err error)

// Chainer - its the chainer
type Chainer struct {
	handlers     []ChainedHandler
	ErrorHandler ChainerErrorHandler
}

// NewChainer - returns new chainer with standard defaults
func NewChainer(handlers ...ChainedHandler) Chainer {
	chainer := Chainer{}
	chainer.handlers = append(chainer.handlers, handlers...)
	return chainer
}

// Append - appends the given handler to the list of common handlers
func (c *Chainer) Append(handler ChainedHandler) {
	c.handlers = append(c.handlers, handler)
}

// Then - execute the list of common handlers and then execute the provided
// sub-chain. Returns a HandlerFunc.
func (c *Chainer) Then(handlers ...ChainedHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		execHandlers := append(c.handlers, handlers...)
		c.executeChain(execHandlers, w, r)
	})
}

// ThenChain - execute the list of common handlers and then execute the provided
// sub-chain. Returns a ChainedHandler.
func (c *Chainer) ThenChain(handlers ...ChainedHandler) ChainedHandler {
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		execHandlers := append(c.handlers, handlers...)
		return c.executeChain(execHandlers, w, r)
	}
}

// ThenFuncs - execute the list of common handler functions and then execute the
// provided sub-chain. Returns a HandlerFunc.
func (c *Chainer) ThenFuncs(handlerFuncs ...http.HandlerFunc) http.HandlerFunc {
	wrappedHandlers := []ChainedHandler{}
	for _, handlerFunc := range handlerFuncs {
		wrappedHandlers = append(wrappedHandlers, WrapWithChainedHandlerFunc(handlerFunc))
	}
	return c.Then(wrappedHandlers...)
}

// WrapWithChainedHandlerFunc - allows a regular handler function to be wrapped if the
// chaining checks are not important
func WrapWithChainedHandlerFunc(handlerFunc http.HandlerFunc) ChainedHandler {
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		handlerFunc(w, r)
		return http.StatusOK, nil
	}
}

func (c Chainer) executeChain(handlers []ChainedHandler, w http.ResponseWriter, r *http.Request) (int, error) {
	for _, handler := range handlers {
		status, err := handler(w, r)
		if err != nil {
			if c.ErrorHandler != nil {
				c.ErrorHandler(w, r, status, err)
			}
			return status, err
		}
	}
	return http.StatusOK, nil
}
