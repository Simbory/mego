package mego

import (
	"strings"
	"fmt"
)

type Area struct {
	pathPrefix string
	server *webServer
}

// Key get the area key/pathPrefix
func (a *Area) Key() string {
	return a.pathPrefix
}

// Dir get the area directory
func (a *Area) Dir() string {
	return server.mapPath(a.pathPrefix)
}

// Get used to register router for GET method
func (a *Area)Get(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("GET", a.fixPath(routePath), handler)
}

// Post used to register router for POST method
func (a *Area)Post(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("POST", a.fixPath(routePath), handler)
}

// Put used to register router for PUT method
func (a *Area)Put(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("PUT", a.fixPath(routePath), handler)
}

// Options used to register router for OPTIONS method
func (a *Area)Options(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("OPTIONS", a.fixPath(routePath), handler)
}

// Head used to register router for HEAD method
func (a *Area)Head(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("HEAD", a.fixPath(routePath), handler)
}

// Delete used to register router for DELETE method
func (a *Area)Delete(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("DELETE", a.fixPath(routePath), handler)
}

// Trace used to register router for TRACE method
func (a *Area)Trace(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("TRACE", a.fixPath(routePath), handler)
}

// Connect used to register router for CONNECT method
func (a *Area)Connect(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("CONNECT", a.fixPath(routePath), handler)
}

// Any used to register router for all methods
func (a *Area)Any(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addRoute("*", a.fixPath(routePath), handler)
}

func (a *Area)fixPath(routePath string) string {
	routePath = strings.Trim(routePath, "/")
	routePath = strings.Trim(routePath, "\\")
	return fmt.Sprintf("%s/%s", a.pathPrefix, routePath)
}