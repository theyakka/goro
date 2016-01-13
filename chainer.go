package goro

import (
	"net/http"
)

type ChainedHandler func(w http.ResponseWriter, r *http.Request) (int, error)

type Chainer struct {
	handlers []ChainedHandler
}

func NewChainer(handlers ...ChainedHandler) Chainer {
	chainer := Chainer{}
	chainer.handlers = append(chainer.handlers, handlers...)
	return chainer
}

func (c Chainer) Then(handlers ...ChainedHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.handlers = append(c.handlers, handlers...)
		c.executeChain(w, r)
	})
}

func (c Chainer) ThenFuncs(handlerFuncs ...http.HandlerFunc) http.Handler {
	wrappedHandlers := []ChainedHandler{}
	for _, handlerFunc := range handlerFuncs {
		wrappedHandlers = append(wrappedHandlers, WrapWithChainedHandlerFunc(handlerFunc))
	}
	return c.Then(wrappedHandlers...)
}

func (c Chainer) executeChain(w http.ResponseWriter, r *http.Request) {
	for _, handler := range c.handlers {
		_, err := handler(w, r)
		if err != nil {
			break
		}
	}
}

// WrapWithChainedHandler - allows a regular handler function to be wrapped if the
// chaining checks are not important
func WrapWithChainedHandlerFunc(handlerFunc http.HandlerFunc) ChainedHandler {
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		handlerFunc(w, r)
		return http.StatusOK, nil
	}
}
