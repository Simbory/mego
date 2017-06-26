package mego

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

var server *webServer

func init() {
	server = newServer()
	server.maxFormSize = 32 << 20
	server.webRoot = WorkingDir()
}

// AssertUnlocked make sure the function only can be called just before the server is running
func AssertUnlocked() {
	server.assertUnlocked()
}

// OnStart attach an event handler to the server start event
func OnStart(h func()) {
	AssertUnlocked()
	if h != nil {
		server.initEvents = append(server.initEvents, h)
	}
}

// AddRouteFunc add route validation func
func AddRouteFunc(name string, fun RouteFunc) {
	AssertUnlocked()
	reg := regexp.MustCompile("^[a-zA-Z][\\w]*$")
	if !reg.Match([]byte(name)) {
		panic(fmt.Errorf("Invalid route func name: %s", name))
	}
	server.routing.addFunc(name, fun)
}

// Get used to register router for GET method
func Get(routePath string, handler ReqHandler) {
	AssertUnlocked()
	server.addRoute("GET", routePath, handler)
}

// Post used to register router for POST method
func Post(routePath string, handler ReqHandler) {
	AssertUnlocked()
	server.addRoute("POST", routePath, handler)
}

// Put used to register router for PUT method
func Put(routePath string, handler ReqHandler) {
	AssertUnlocked()
	server.addRoute("PUT", routePath, handler)
}

// Options used to register router for OPTIONS method
func Options(routePath string, handler ReqHandler) {
	AssertUnlocked()
	server.addRoute("OPTIONS", routePath, handler)
}

// Head used to register router for HEAD method
func Head(routePath string, handler ReqHandler) {
	AssertUnlocked()
	server.addRoute("HEAD", routePath, handler)
}

// Delete used to register router for DELETE method
func Delete(routePath string, handler ReqHandler) {
	AssertUnlocked()
	server.addRoute("DELETE", routePath, handler)
}

// Trace used to register router for TRACE method
func Trace(routePath string, handler ReqHandler) {
	AssertUnlocked()
	server.addRoute("TRACE", routePath, handler)
}

// Connect used to register router for CONNECT method
func Connect(routePath string, handler ReqHandler) {
	AssertUnlocked()
	server.addRoute("CONNECT", routePath, handler)
}

// Any used to register router for all methods
func Any(routePath string, handler ReqHandler) {
	AssertUnlocked()
	server.addRoute("*", routePath, handler)
}

// GetArea get or create the mego area
func GetArea(pathPrefix string) *Area {
	prefix := strings.Trim(pathPrefix, "/")
	prefix = strings.Trim(prefix, "\\")
	reg := regexp.MustCompile("[/a-zA-Z0-9_-]+")
	if !reg.Match([]byte(prefix)) {
		panic(errors.New("Invalid pathPrefix:" + pathPrefix))
	}
	return &Area{
		pathPrefix: "/" + prefix,
		server: server,
		engineLock: &sync.RWMutex{},
	}
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func MapPath(virtualPath string) string {
	return server.mapPath(virtualPath)
}

// HandleDir handle static directory
func HandleDir(pathPrefix string) {
	AssertUnlocked()
	if len(pathPrefix) == 0 {
		panic(errors.New("The parameter 'pathPrefix' cannot be empty"))
	}
	if !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = "/" + pathPrefix
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	server.staticDirs[pathPrefix] = nil
}

// HandleFile handle the url as static file
func HandleFile(url string) {
	AssertUnlocked()
	server.staticFiles[url] = ""
}

// Handle404 set custom error handler for status code 404
func Handle404(h http.HandlerFunc) {
	AssertUnlocked()
	if h != nil {
		server.err404Handler = h
	}
}

// Handle500 set custom error handler for status code 500
func Handle500(h func(http.ResponseWriter, *http.Request, interface{})) {
	AssertUnlocked()
	if h != nil {
		server.err500Handler = h
	}
}

// HandleFilter handle the path
// the path prefix ends with char '*', the filter will be available for all urls that
// starts with pathPrefix. Otherwise, the filter only be available for the featured url
func HandleFilter(pathPrefix string, h func(*Context)) error {
	AssertUnlocked()
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
	server.filters.add(pathPrefix, matchAll, h)
	return nil
}

func SetRoot(wwwRoot string) {
	AssertUnlocked()
	server.webRoot = wwwRoot
}

func ExtendView(name string, f interface{}) {
	AssertUnlocked()
	server.initViewEngine()
	server.viewEngine.ExtendView(name, f)
}

// Run run the application as http
func Run(addr string) {
	server.onInit()
	err := http.ListenAndServe(addr, server)
	if err != nil {
		panic(err)
	}
}

// RunTLS run the application as https
func RunTLS(addr, certFile, keyFile string) {
	server.onInit()
	err := http.ListenAndServeTLS(addr, certFile, keyFile, server)
	if err != nil {
		panic(err)
	}
}
