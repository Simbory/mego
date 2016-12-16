package mego

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
)

// Error500Handler define the internal server error handler func
type Error500Handler func(http.ResponseWriter, *http.Request, interface{})

var (
	locked                         = false
	routing                        = newRouteTree()
	initEvents                     = []func(){}
	staticDirs                     = make(map[string]http.Handler)
	staticFiles                    = make(map[string]string)
	err404Handler http.HandlerFunc = handle404
	err500Handler Error500Handler  = handle500
	filters                        = make(filterContainer)
)

// AssertNotLock make sure the function only can be called just before the server is running
func AssertNotLock() {
	if locked {
		panic(errors.New("cannot call this function while the server is runing"))
	}
}

// handle404 the default error 404 handler
func handle404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write([]byte("Error 404: Not Found"))
}

// handle500 the default error 500 handler
func handle500(w http.ResponseWriter, r *http.Request, rec interface{}) {
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "text-plain")
	var debugStack = string(debug.Stack())
	debugStack = strings.Replace(debugStack, "<", "&lt;", -1)
	debugStack = strings.Replace(debugStack, ">", "&gt;", -1)
	buf := &bytes.Buffer{}
	if err, ok := rec.(error); ok {
		buf.WriteString(err.Error())
		buf.WriteString("\r\n\r\n")
	}
	buf.WriteString(debugStack)
	w.Write(buf.Bytes())
}

// init mego server
func initMego() {
	if len(initEvents) > 0 {
		for _, h := range initEvents {
			h()
		}
	}
}

// OnStart attach an event handler to the server start event
func OnStart(h func()) {
	AssertNotLock()
	if h != nil {
		initEvents = append(initEvents, h)
	}
}

// AddRouteFunc add route validation func
func AddRouteFunc(name string, fun RouteFunc) error {
	AssertNotLock()
	reg := regexp.MustCompile("^[a-zA-Z_][\\w]*$")
	if !reg.Match([]byte(name)) {
		return fmt.Errorf("Invalid route func name: %s", name)
	}
	return routing.addFunc(name, fun)
}

// Get used to register router for GET method
func Get(routePath string, handler ReqHandler) error {
	AssertNotLock()
	return routing.addRoute("GET", routePath, handler)
}

// Post used to register router for POST method
func Post(routePath string, handler ReqHandler) error {
	AssertNotLock()
	return routing.addRoute("POST", routePath, handler)
}

// Put used to register router for PUT method
func Put(routePath string, handler ReqHandler) error {
	AssertNotLock()
	return routing.addRoute("PUT", routePath, handler)
}

// Options used to register router for OPTIONS method
func Options(routePath string, handler ReqHandler) error {
	AssertNotLock()
	return routing.addRoute("OPTIONS", routePath, handler)
}

// Head used to register router for HEAD method
func Head(routePath string, handler ReqHandler) error {
	AssertNotLock()
	return routing.addRoute("HEAD", routePath, handler)
}

// Delete used to register router for DELETE method
func Delete(routePath string, handler ReqHandler) error {
	AssertNotLock()
	return routing.addRoute("DELETE", routePath, handler)
}

// Trace used to register router for TRACE method
func Trace(routePath string, handler ReqHandler) error {
	AssertNotLock()
	return routing.addRoute("TRACE", routePath, handler)
}

// Connect used to register router for CONNECT method
func Connect(routePath string, handler ReqHandler) error {
	AssertNotLock()
	return routing.addRoute("CONNECT", routePath, handler)
}

// Any used to register router for all methods
func Any(routePath string, handler ReqHandler) error {
	AssertNotLock()
	return routing.addRoute("*", routePath, handler)
}

// HandleStaticDir handle static directory
func HandleStaticDir(pathPrefix, dirPath string) error {
	AssertNotLock()
	if len(pathPrefix) == 0 {
		return errors.New("The parameter 'pathPrefix' cannot be empty")
	}
	if len(dirPath) == 0 {
		dirPath = "."
	}
	if !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = "/" + pathPrefix
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	staticDirs[pathPrefix] = http.FileServer(http.Dir(dirPath))
	return nil
}

// HandleStaticFile handle the url as static file
func HandleStaticFile(url, filePath string) {
	AssertNotLock()
	staticFiles[url] = filePath
}

// Handle404 set custom error handler for status code 404
func Handle404(h http.HandlerFunc) {
	AssertNotLock()
	if h != nil {
		err404Handler = h
	}
}

// Handle500 set custom error handler for status code 500
func Handle500(h func(http.ResponseWriter, *http.Request, interface{})) {
	AssertNotLock()
	if h != nil {
		err500Handler = h
	}
}

// Filter set path filter
// the path prefix ends with char '*', the filter will be available for all urls that
// starts with pathPrefix. Otherwise, the filter only be available for the featured url
func Filter(pathPrefix string, h func(*Context)) error {
	AssertNotLock()
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
	filters.add(pathPrefix, matchAll, h)
	return nil
}

// Run run the application as http
func Run(addr string) {
	initMego()
	svr := &server{}
	err := http.ListenAndServe(addr, svr)
	if err != nil {
		panic(err)
	}
}

// RunTLS run the application as https
func RunTLS(addr, certFile, keyFile string) {
	initMego()
	svr := &server{}
	err := http.ListenAndServeTLS(addr, certFile, keyFile, svr)
	if err != nil {
		panic(err)
	}
}
