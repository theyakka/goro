package goro

import (
	"net/http"
	"sync"
)

type ContextInterface interface {
	Get(req *http.Request, key string) interface{}
	Put(req *http.Request, key string, value interface{})
	Clear(req *http.Request)
	ClearAll()
}

type Context struct {
	mutex  sync.RWMutex
	values map[*http.Request]map[string]interface{}
}

func NewContext() *Context {
	c := &Context{
		values: make(map[*http.Request]map[string]interface{}),
	}
	return c
}

func (c *Context) Put(req *http.Request, key string, value interface{}) {
	c.mutex.Lock()
	if c.values[req] == nil {
		c.values[req] = make(map[string]interface{})
	}
	c.values[req][key] = value
	c.mutex.Unlock()
}

func (c *Context) Get(req *http.Request, key string) interface{} {
	c.mutex.RLock()
	if reqVals := c.values[req]; reqVals != nil {
		val := reqVals[key]
		c.mutex.RUnlock()
		return val
	}
	c.mutex.RUnlock()
	return nil
}

func (c *Context) ClearKey(req *http.Request, key string) {
	c.mutex.Lock()
	delete(c.values[req], key)
	c.mutex.Unlock()
}

func (c *Context) Clear(req *http.Request) {
	c.mutex.Lock()
	delete(c.values, req)
	c.mutex.Unlock()
}

func (c *Context) ClearAll() {
	c.mutex.Lock()
	c.values = make(map[*http.Request]map[string]interface{})
	c.mutex.Unlock()
}
