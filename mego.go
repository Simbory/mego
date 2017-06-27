package mego

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"path"
)

// AssertUnlocked make sure the function only can be called just before the s is running
func (s *Server)AssertUnlocked() {
	s.assertUnlocked()
}

// OnStart attach an event handler to the s start event
func (s *Server)OnStart(h func()) {
	s.AssertUnlocked()
	if h != nil {
		s.initEvents = append(s.initEvents, h)
	}
}

// AddRouteFunc add route validation func
func (s *Server)AddRouteFunc(name string, fun RouteFunc) {
	s.AssertUnlocked()
	reg := regexp.MustCompile("^[a-zA-Z][\\w]*$")
	if !reg.Match([]byte(name)) {
		panic(fmt.Errorf("Invalid route func name: %s", name))
	}
	s.routing.addFunc(name, fun)
}

// Get used to register router for GET method
func (s *Server)Get(routePath string, handler ReqHandler) {
	s.AssertUnlocked()
	s.addRoute("GET", routePath, handler)
}

// Post used to register router for POST method
func (s *Server)Post(routePath string, handler ReqHandler) {
	s.AssertUnlocked()
	s.addRoute("POST", routePath, handler)
}

// Put used to register router for PUT method
func (s *Server)Put(routePath string, handler ReqHandler) {
	s.AssertUnlocked()
	s.addRoute("PUT", routePath, handler)
}

// Options used to register router for OPTIONS method
func (s *Server)Options(routePath string, handler ReqHandler) {
	s.AssertUnlocked()
	s.addRoute("OPTIONS", routePath, handler)
}

// Head used to register router for HEAD method
func (s *Server)Head(routePath string, handler ReqHandler) {
	s.AssertUnlocked()
	s.addRoute("HEAD", routePath, handler)
}

// Delete used to register router for DELETE method
func (s *Server)Delete(routePath string, handler ReqHandler) {
	s.AssertUnlocked()
	s.addRoute("DELETE", routePath, handler)
}

// Trace used to register router for TRACE method
func (s *Server)Trace(routePath string, handler ReqHandler) {
	s.AssertUnlocked()
	s.addRoute("TRACE", routePath, handler)
}

// Connect used to register router for CONNECT method
func (s *Server)Connect(routePath string, handler ReqHandler) {
	s.AssertUnlocked()
	s.addRoute("CONNECT", routePath, handler)
}

// Any used to register router for all methods
func (s *Server)Any(routePath string, handler ReqHandler) {
	s.AssertUnlocked()
	s.addRoute("*", routePath, handler)
}

// GetArea get or create the mego area
func (s *Server)GetArea(pathPrefix string) *Area {
	prefix := strings.Trim(pathPrefix, "/")
	prefix = strings.Trim(prefix, "\\")
	reg := regexp.MustCompile("[/a-zA-Z0-9_-]+")
	if !reg.Match([]byte(prefix)) {
		panic(errors.New("Invalid pathPrefix:" + pathPrefix))
	}
	return &Area{
		pathPrefix: "/" + prefix,
		server: s,
		engineLock: &sync.RWMutex{},
	}
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func (s *Server)MapPath(virtualPath string) string {
	p := path.Join(s.webRoot, virtualPath)
	p = path.Clean(strings.Replace(p, "\\", "/", -1))
	return strings.TrimRight(p, "/")
}

// HandleDir handle static directory
func (s *Server)HandleDir(pathPrefix string) {
	s.AssertUnlocked()
	if len(pathPrefix) == 0 {
		panic(errors.New("The parameter 'pathPrefix' cannot be empty"))
	}
	if !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = "/" + pathPrefix
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	s.staticDirs[pathPrefix] = nil
}

// HandleFile handle the url as static file
func (s *Server)HandleFile(url string) {
	s.AssertUnlocked()
	s.staticFiles[url] = ""
}

// Handle404 set custom error handler for status code 404
func (s *Server)Handle404(h http.HandlerFunc) {
	s.AssertUnlocked()
	if h != nil {
		s.err404Handler = h
	}
}

// Handle500 set custom error handler for status code 500
func (s *Server)Handle500(h func(http.ResponseWriter, *http.Request, interface{})) {
	s.AssertUnlocked()
	if h != nil {
		s.err500Handler = h
	}
}

// HandleFilter handle the path
// the path prefix ends with char '*', the filter will be available for all urls that
// starts with pathPrefix. Otherwise, the filter only be available for the featured url
func (s *Server)HandleFilter(pathPrefix string, h func(*Context)) error {
	s.AssertUnlocked()
	if len(pathPrefix) == 0 {
		return errors.New("The parameter 'pathPrefix' cannot be empty")
	}
	if h == nil {
		return errors.New("The parameter 'h' cannot be nil")
	}
	if !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = "/" + pathPrefix
	}
	var matchAll bool
	if strings.HasSuffix(pathPrefix, "*") {
		matchAll = true
		pathPrefix = strings.TrimRight(pathPrefix ,"*")
		if !strings.HasSuffix(pathPrefix, "/") {
			pathPrefix = pathPrefix + "/"
		}
	} else {
		matchAll = false
	}
	s.filters.add(pathPrefix, matchAll, h)
	return nil
}

func (s *Server)ExtendView(name string, f interface{}) {
	s.AssertUnlocked()
	s.initViewEngine()
	s.viewEngine.ExtendView(name, f)
}

// Run run the application as http
func (s *Server)Run(addr string) {
	s.onInit()
	err := http.ListenAndServe(addr, s)
	if err != nil {
		panic(err)
	}
}

// RunTLS run the application as https
func (s *Server)RunTLS(addr, certFile, keyFile string) {
	s.onInit()
	err := http.ListenAndServeTLS(addr, certFile, keyFile, s)
	if err != nil {
		panic(err)
	}
}
