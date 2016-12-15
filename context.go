package mego

import "net/http"

// Context the mego context struct
type Context struct {
	req       *http.Request
	res       http.ResponseWriter
	routeData map[string]string
	ended     bool
	items     map[string]interface{}
}

// Request get the mego request
func (c *Context) Request() *http.Request {
	return c.req
}

// Response get the mego response
func (c *Context) Response() http.ResponseWriter {
	return c.res
}

// RouteParam find the route parameter by key
func (c *Context) RouteParam(key string) string {
	if c.routeData == nil {
		return ""
	}
	return c.routeData[key]
}

// SetItem add context data to mego context
func (c *Context) SetItem(key string, data interface{}) {
	if len(key) == 0 {
		return
	}
	if c.items == nil {
		c.items = make(map[string]interface{})
	}
	c.items[key] = data
}

// GetItem get the context data from mego context by key
func (c *Context) GetItem(key string) interface{} {
	if c.items == nil {
		return nil
	}
	return c.items[key]
}

// DelIem delete context item from mego context by key
func (c *Context) DelItem(key string) {
	if c.items == nil {
		return
	}
	delete(c.items, key)
}

// End end the mego context and stop the rest request function
func (c *Context) End() {
	c.ended = true
}
