package goro

// import (
// 	"net/http"
// )

// type ChainedHandler func(w http.ResponseWriter, r *http.Request) (int, error)

// type Chainer struct {
// 	handlers []ChainedHandler
// }

// func NewChainer(handlers ...ChainedHandler) Chainer {
// 	chainer := Chainer{}
// 	chainer.handlers = append(chainer.handlers, handlers...)
// 	return chainer
// }

// func (c Chainer) Then(handlers ...ChainedHandler) http.Handler {
// 	return
// }

// func (c Chainer) ThenFuncs(handlers ...http.HandlerFunc) http.Handler {
// 	wrappedHandlers := []ChainedHandler{}

// 	return c.Then(wrappedHandlers)
// }

// // WrapWithChainedHandler - allows a regular handler function to be wrapped if the
// // chaining checks are not important
// func WrapWithChainedHandler(handlerfunc http.HandlerFunc) ChainedHandler {
// 	return func(w http.ResponseWriter, r *http.Request) (bool, error) {
// 		handlerfunc(w, r)
// 		return true, nil
// 	}
// }
