package mego

import (
	"net/http"
	"strconv"
	"strings"
)

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

// RouteParamString get the route parameter value as string by key
func (c *Context) RouteParamString(key string) string {
	if c.routeData == nil {
		return ""
	}
	return c.routeData[key]
}

// RouteParamInt get the route parameter value as int64 by key
func (c *Context) RouteParamInt(key string) int64 {
	var rawValue = c.RouteParamString(key)
	if len(rawValue) == 0 {
		return 0
	}
	value,err := strconv.ParseInt(rawValue, 0, 64)
	if err != nil {
		return 0
	}
	return value
}

// RouteParamUint get the route parameter value as uint64 by key
func (c *Context) RouteParamUint(key string) uint64 {
	var rawValue = c.RouteParamString(key)
	if len(rawValue) == 0 {
		return 0
	}
	value,err := strconv.ParseUint(rawValue, 0, 64)
	if err != nil {
		return 0
	}
	return value
}

// RouteParamFloat get the route parameter value as float by key
func (c *Context) RouteParamFloat(key string) float64 {
	var rawValue = c.RouteParamString(key)
	if len(rawValue) == 0 {
		return 0
	}
	value,err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		return 0
	}
	return value
}

// RouteParamBool get the route parameter value as boolean by key
func (c *Context) RouteParamBool(key string) bool {
	var rawValue = c.RouteParamString(key)
	if len(rawValue) == 0 || strings.ToLower(rawValue) == "false" || rawValue == "0"  {
		return false
	}
	return true
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

// DelItem delete context item from mego context by key
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
