package mego

import (
	"fmt"
	"github.com/simbory/mego/assert"
	"strings"
	"sync"
)

// Area implement mego area
type Area struct {
	pathPrefix string
	server     *Server
	viewEngine *ViewEngine
	engineLock *sync.RWMutex
}

// Key get the area key/pathPrefix
func (a *Area) Key() string {
	return a.pathPrefix
}

// Dir get the physical directory of the current area
func (a *Area) Dir() string {
	return a.server.MapRootPath(a.pathPrefix)
}

// Get used to register router for GET method
func (a *Area) Get(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("GET", a.fixPath(routePath), a, handler)
}

// Post used to register router for POST method
func (a *Area) Post(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("POST", a.fixPath(routePath), a, handler)
}

// Put used to register router for PUT method
func (a *Area) Put(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("PUT", a.fixPath(routePath), a, handler)
}

// Options used to register router for OPTIONS method
func (a *Area) Options(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("OPTIONS", a.fixPath(routePath), a, handler)
}

// Head used to register router for HEAD method
func (a *Area) Head(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("HEAD", a.fixPath(routePath), a, handler)
}

// Delete used to register router for DELETE method
func (a *Area) Delete(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("DELETE", a.fixPath(routePath), a, handler)
}

// Trace used to register router for TRACE method
func (a *Area) Trace(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("TRACE", a.fixPath(routePath), a, handler)
}

// Connect used to register router for CONNECT method
func (a *Area) Connect(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("CONNECT", a.fixPath(routePath), a, handler)
}

// Any used to register router for all methods
func (a *Area) Any(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("*", a.fixPath(routePath), a, handler)
}

// HandleFilter handle the filter func in current area
func (a *Area) HandleFilter(pathPrefix string, h func(*HttpCtx)) {
	a.server.assertUnlocked()
	assert.NotEmpty("pathPrefix", pathPrefix)
	assert.NotNil("h", h)
	prefix := ClearPath(pathPrefix)
	prefix = strings.Trim(prefix, "/")
	a.server.HandleFilter(strAdd(a.Key(), "/", prefix), func(ctx *HttpCtx) {
		if ctx.area != nil && ctx.area.Key() == a.Key() {
			h(ctx)
		}
	})
}

// ExtendView add view function to the view engine of the current area
func (a *Area) ExtendView(name string, f interface{}) {
	a.server.assertUnlocked()
	a.initViewEngine()
	a.viewEngine.ExtendView(name, f)
}

func (a *Area) initViewEngine() {
	if a.viewEngine == nil {
		a.engineLock.Lock()
		defer a.engineLock.Unlock()
		if a.viewEngine == nil {
			a.viewEngine = NewViewEngine(a.server.MapRootPath(a.Key()+"/views"), ".html")
		}
	}
}

func (a *Area) fixPath(routePath string) string {
	routePath = strings.Trim(routePath, "/")
	routePath = strings.Trim(routePath, "\\")
	return fmt.Sprintf("%s/%s", a.pathPrefix, routePath)
}
