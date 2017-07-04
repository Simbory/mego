package mego

import (
	"net/http"
	"regexp"
	"strings"
	"sync"
	"path"
	"github.com/simbory/mego/assert"
)

// OnStart attach an event handler to the s start event
func (s *Server) OnStart(h func()) {
	s.assertUnlocked()
	if h != nil {
		s.initEvents = append(s.initEvents, h)
	}
}

var routeNameReg = regexp.MustCompile("^[a-zA-Z][\\w]*$")
// AddRouteFunc add route validation func
func (s *Server) AddRouteFunc(name string, fun RouteFunc) {
	s.assertUnlocked()
	assert.Assert("name", func() bool {
		return routeNameReg.Match([]byte(name))
	})
	s.routing.addFunc(name, fun)
}

// Get used to register router for GET method
func (s *Server) Get(routePath string, handler ReqHandler) {
	s.assertUnlocked()
	s.addRoute("GET", routePath, handler)
}

// Post used to register router for POST method
func (s *Server) Post(routePath string, handler ReqHandler) {
	s.assertUnlocked()
	s.addRoute("POST", routePath, handler)
}

// Put used to register router for PUT method
func (s *Server) Put(routePath string, handler ReqHandler) {
	s.assertUnlocked()
	s.addRoute("PUT", routePath, handler)
}

// Options used to register router for OPTIONS method
func (s *Server) Options(routePath string, handler ReqHandler) {
	s.assertUnlocked()
	s.addRoute("OPTIONS", routePath, handler)
}

// Head used to register router for HEAD method
func (s *Server) Head(routePath string, handler ReqHandler) {
	s.assertUnlocked()
	s.addRoute("HEAD", routePath, handler)
}

// Delete used to register router for DELETE method
func (s *Server) Delete(routePath string, handler ReqHandler) {
	s.assertUnlocked()
	s.addRoute("DELETE", routePath, handler)
}

// Trace used to register router for TRACE method
func (s *Server) Trace(routePath string, handler ReqHandler) {
	s.assertUnlocked()
	s.addRoute("TRACE", routePath, handler)
}

// Connect used to register router for CONNECT method
func (s *Server) Connect(routePath string, handler ReqHandler) {
	s.assertUnlocked()
	s.addRoute("CONNECT", routePath, handler)
}

// Any used to register router for all methods
func (s *Server) Any(routePath string, handler ReqHandler) {
	s.assertUnlocked()
	s.addRoute("*", routePath, handler)
}

var areaNameReg = regexp.MustCompile("[/a-zA-Z0-9_-]+")

// GetArea get or create the mego area
func (s *Server) GetArea(pathPrefix string) *Area {
	prefix := ClearPath(pathPrefix)
	assert.Assert("pathPrefix", func() bool {
		return areaNameReg.Match([]byte(prefix))
	})
	prefix = EnsurePrefix(prefix, "/")
	return &Area{
		pathPrefix: prefix,
		server: s,
		engineLock: &sync.RWMutex{},
	}
}

// MapRootPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func (s *Server) MapRootPath(virtualPath string) string {
	p := path.Join(s.webRoot, virtualPath)
	return path.Clean(strings.Replace(p, "\\", "/", -1))
}

// MapContentPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func (s *Server) MapContentPath(virtualPath string) string {
	p := path.Join(s.contentRoot, virtualPath)
	return path.Clean(strings.Replace(p, "\\", "/", -1))
}

// Handle404 set custom error handler for status code 404
func (s *Server) Handle404(h http.HandlerFunc) {
	s.assertUnlocked()
	if h != nil {
		s.err404Handler = h
	}
}

// Handle404 set custom error handler for status code 404
func (s *Server) Handle400(h http.HandlerFunc) {
	s.assertUnlocked()
	if h != nil {
		s.err400Handler = h
	}
}

// Handle404 set custom error handler for status code 404
func (s *Server) Handle403(h http.HandlerFunc) {
	s.assertUnlocked()
	if h != nil {
		s.err403Handler = h
	}
}

// Handle500 set custom error handler for status code 500
func (s *Server) Handle500(h func(http.ResponseWriter, *http.Request, interface{})) {
	s.assertUnlocked()
	if h != nil {
		s.err500Handler = h
	}
}

// HandleFilter handle the path
// the path prefix ends with char '*', the filter will be available for all urls that
// starts with pathPrefix. Otherwise, the filter only be available for the featured url
func (s *Server) HandleFilter(pathPrefix string, h func(*HttpCtx)) error {
	s.assertUnlocked()
	assert.NotEmpty("pathPrefix", pathPrefix)
	assert.NotNil("h", h)
	if !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = "/" + pathPrefix
	}
	var matchAll bool
	if strings.HasSuffix(pathPrefix, "*") {
		matchAll = true
		pathPrefix = EnsureSuffix(strings.TrimRight(pathPrefix ,"*"), "/")
	} else {
		matchAll = false
	}
	s.filters.add(pathPrefix, matchAll, h)
	return nil
}

// ExtendView extend the view engine with func f
func (s *Server)ExtendView(name string, f interface{}) {
	s.assertUnlocked()
	s.initViewEngine()
	s.viewEngine.ExtendView(name, f)
}

// Run run the application as http
func (s *Server)Run() {
	s.onInit()
	err := http.ListenAndServe(s.addr, s)
	assert.PanicErr(err)
}

// RunTLS run the application as https
func (s *Server)RunTLS(certFile, keyFile string) {
	s.onInit()
	err := http.ListenAndServeTLS(s.addr, certFile, keyFile, s)
	assert.PanicErr(err)
}
