package goro

// ContextHandler - the standard Goro handler
type ContextHandler interface {
	Serve(ctx *HandlerContext)
}

// ContextHandlerFunc - the standard Goro handler
type ContextHandlerFunc func(ctx *HandlerContext)

// Serve - implement the ContextHandler interface
func (chf ContextHandlerFunc) Serve(ctx *HandlerContext) {
	chf(ctx)
}
