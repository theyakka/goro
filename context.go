package goro

type ContextInterface interface {
	Get(key string) interface{}
	Put(key string, value interface{})
	Clear()
}

type Context struct {
	values map[string]interface{}
}

func NewContext() *Context {
	c := &Context{
		values: make(map[string]interface{}),
	}
	return c
}

func (c *Context) Put(key string, value interface{}) {
	c.values[key] = value
}

func (c *Context) Get(key string) interface{} {
	return c.values[key]
}

func (c *Context) Clear() {
	c.values = make(map[string]interface{})
}
