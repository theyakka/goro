package goro

import (
	"net/http"
)

type ChainedHandler func(w http.ResponseWriter, r *http.Request) (int, error)
type ChainerErrorHandler func(w http.ResponseWriter, r *http.Request, status int, err error)

type Chainer struct {
	handlers []ChainedHandler

	ErrorHandler ChainerErrorHandler
}

func NewChainer(handlers ...ChainedHandler) Chainer {
	chainer := Chainer{}
	chainer.handlers = append(chainer.handlers, handlers...)
	return chainer
}

func (c *Chainer) Append(handler ChainedHandler) {
	c.handlers = append(c.handlers, handler)
}

func (c *Chainer) Then(handlers ...ChainedHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		execHandlers := append(c.handlers, handlers...)
		c.executeChain(execHandlers, w, r)
	})
}

func (c *Chainer) ThenChain(handlers ...ChainedHandler) ChainedHandler {
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		execHandlers := append(c.handlers, handlers...)
		return c.executeChain(execHandlers, w, r)
	}
}

func (c *Chainer) ThenFuncs(handlerFuncs ...http.HandlerFunc) http.HandlerFunc {
	wrappedHandlers := []ChainedHandler{}
	for _, handlerFunc := range handlerFuncs {
		wrappedHandlers = append(wrappedHandlers, WrapWithChainedHandlerFunc(handlerFunc))
	}
	return c.Then(wrappedHandlers...)
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

// WrapWithChainedHandler - allows a regular handler function to be wrapped if the
// chaining checks are not important
func WrapWithChainedHandlerFunc(handlerFunc http.HandlerFunc) ChainedHandler {
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		handlerFunc(w, r)
		return http.StatusOK, nil
	}
}
