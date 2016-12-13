package mego

import "net/http"

// Context the request context struct
type Context struct {
	req       *http.Request
	res       http.ResponseWriter
	routeData map[string]string
	ended     bool
	items     map[string]interface{}
}

func (c *Context) Request() *http.Request {
	return c.req
}

func (c *Context) Response() http.ResponseWriter {
	return c.res
}

func (c *Context) RouteParam(key string) string {
	if c.routeData == nil {
		return ""
	}
	return c.routeData[key]
}

func (c *Context) SetItem(key string, data interface{}) {
	if len(key) == 0 {
		return
	}
	if c.items == nil {
		c.items = make(map[string]interface{})
	}
	c.items[key] = data
}

func (c *Context) GetItem(key string) interface{} {
	if c.items == nil {
		return nil
	}
	return c.items[key]
}

func (c *Context) DelItem(key string) {
	if c.items == nil {
		return
	}
	delete(c.items, key)
}

func (c *Context) End() {
	c.ended = true
}