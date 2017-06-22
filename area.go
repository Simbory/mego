package mego

import (
	"strings"
	"fmt"
)

type area struct {
	pathPrefix string
	server *webServer
}

// Get used to register router for GET method
func (a *area)Get(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("GET", a.fixPath(routePath), handler)
}

// Post used to register router for POST method
func (a *area)Post(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("POST", a.fixPath(routePath), handler)
}

// Put used to register router for PUT method
func (a *area)Put(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("PUT", a.fixPath(routePath), handler)
}

// Options used to register router for OPTIONS method
func (a *area)Options(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("OPTIONS", a.fixPath(routePath), handler)
}

// Head used to register router for HEAD method
func (a *area)Head(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("HEAD", a.fixPath(routePath), handler)
}

// Delete used to register router for DELETE method
func (a *area)Delete(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("DELETE", a.fixPath(routePath), handler)
}

// Trace used to register router for TRACE method
func (a *area)Trace(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("TRACE", a.fixPath(routePath), handler)
}

// Connect used to register router for CONNECT method
func (a *area)Connect(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("CONNECT", a.fixPath(routePath), handler)
}

// Any used to register router for all methods
func (a *area)Any(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("*", a.fixPath(routePath), handler)
}

func (a *area)fixPath(routePath string) string {
	routePath = strings.Trim(routePath, "/")
	routePath = strings.Trim(routePath, "\\")
	return fmt.Sprintf("%s/%s", a.pathPrefix, routePath)
}