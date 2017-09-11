package mego

import (
	"fmt"
	//"github.com/simbory/mego/assert"
	"strings"
	"sync"
	"github.com/simbory/mego/assert"
)

// Area implement mego area
type Area struct {
	pathPrefix string
	server     *Server
	viewEngine *ViewEngine
	hijackColl hijackContainer
	engineLock sync.RWMutex
}

// Key get the area key/pathPrefix
func (a *Area) Key() string {
	return a.pathPrefix
}

// Dir get the physical directory of the current area
func (a *Area) Dir() string {
	return a.server.MapRootPath(a.pathPrefix)
}

// Route used to register router for all methods
func (a *Area) Route(routePath string, handler interface{}) {
	a.server.assertUnlocked()
	a.server.addAreaRoute(a.fixPath(routePath), a, handler)
}

// HijackRequest hijack the area dynamic request that starts with pathPrefix
func (a *Area) HijackRequest(pathPrefix string, h func(*HttpCtx)) {
	a.server.assertUnlocked()
	assert.NotEmpty("pathPrefix", pathPrefix)
	assert.NotNil("h", h)
	prefix := ClearPath(pathPrefix)
	prefix = a.pathPrefix + "/" + strings.Trim(prefix, "/")
	a.hijackColl.add(prefix, h)
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
			a.viewEngine = NewViewEngine(a.server.MapRootPath(a.Key() + "/views"))
		}
	}
}

func (a *Area) fixPath(routePath string) string {
	routePath = strings.Trim(routePath, "/")
	routePath = strings.Trim(routePath, "\\")
	return fmt.Sprintf("%s/%s", a.pathPrefix, routePath)
}
