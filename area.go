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

// RouteGet used to register router for GET method
func (a *Area) RouteGet(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("GET", a.fixPath(routePath), a, handler)
}

// RoutePost used to register router for POST method
func (a *Area) RoutePost(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("POST", a.fixPath(routePath), a, handler)
}

// RoutePut used to register router for PUT method
func (a *Area) RoutePut(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("PUT", a.fixPath(routePath), a, handler)
}

// RouteOptions used to register router for OPTIONS method
func (a *Area) RouteOptions(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("OPTIONS", a.fixPath(routePath), a, handler)
}

// RouteDel used to register router for DELETE method
func (a *Area) RouteDel(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("DELETE", a.fixPath(routePath), a, handler)
}

// RouteTrace used to register router for TRACE method
func (a *Area) RouteTrace(routePath string, handler ReqHandler) {
	a.server.assertUnlocked()
	a.server.addAreaRoute("TRACE", a.fixPath(routePath), a, handler)
}

// Route used to register router for all methods
func (a *Area) Route(routePath string, handler interface{}) {
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
			a.viewEngine = NewViewEngine(a.server.MapRootPath(a.Key()+"/views"), ".gohtml")
		}
	}
}

func (a *Area) fixPath(routePath string) string {
	routePath = strings.Trim(routePath, "/")
	routePath = strings.Trim(routePath, "\\")
	return fmt.Sprintf("%s/%s", a.pathPrefix, routePath)
}
