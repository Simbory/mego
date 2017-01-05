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

var svr = newServer()

// AssertNotLock make sure the function only can be called just before the server is running
func AssertNotLock() {
	if svr.locked {
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

// OnStart attach an event handler to the server start event
func OnServerStart(h func()) {
	AssertNotLock()
	if h != nil {
		svr.initEvents = append(svr.initEvents, h)
	}
}

// AddRouteFunc add route validation func
func AddRouteFunc(name string, fun RouteFunc) {
	AssertNotLock()
	reg := regexp.MustCompile("^[a-zA-Z_][\\w]*$")
	if !reg.Match([]byte(name)) {
		panic(fmt.Errorf("Invalid route func name: %s", name))
	}
	svr.routing.addFunc(name, fun)
}

// Get used to register router for GET method
func Get(routePath string, handler ReqHandler) {
	AssertNotLock()
	svr.appendRouteSetting("GET", routePath, handler)
}

// Post used to register router for POST method
func Post(routePath string, handler ReqHandler) {
	AssertNotLock()
	svr.appendRouteSetting("POST", routePath, handler)
}

// Put used to register router for PUT method
func Put(routePath string, handler ReqHandler) {
	AssertNotLock()
	svr.appendRouteSetting("PUT", routePath, handler)
}

// Options used to register router for OPTIONS method
func Options(routePath string, handler ReqHandler) {
	AssertNotLock()
	svr.appendRouteSetting("OPTIONS", routePath, handler)
}

// Head used to register router for HEAD method
func Head(routePath string, handler ReqHandler) {
	AssertNotLock()
	svr.appendRouteSetting("HEAD", routePath, handler)
}

// Delete used to register router for DELETE method
func Delete(routePath string, handler ReqHandler) {
	AssertNotLock()
	svr.appendRouteSetting("DELETE", routePath, handler)
}

// Trace used to register router for TRACE method
func Trace(routePath string, handler ReqHandler) {
	AssertNotLock()
	svr.appendRouteSetting("TRACE", routePath, handler)
}

// Connect used to register router for CONNECT method
func Connect(routePath string, handler ReqHandler) {
	AssertNotLock()
	svr.appendRouteSetting("CONNECT", routePath, handler)
}

// Any used to register router for all methods
func Any(routePath string, handler ReqHandler) {
	AssertNotLock()
	svr.appendRouteSetting("*", routePath, handler)
}

// HandleStaticDir handle static directory
func HandleStaticDir(pathPrefix, dirPath string) {
	AssertNotLock()
	if len(pathPrefix) == 0 {
		panic(errors.New("The parameter 'pathPrefix' cannot be empty"))
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
	svr.staticDirs[pathPrefix] = http.FileServer(http.Dir(dirPath))
}

// HandleStaticFile handle the url as static file
func HandleStaticFile(url, filePath string) {
	AssertNotLock()
	svr.staticFiles[url] = filePath
}

// Handle404 set custom error handler for status code 404
func Handle404(h http.HandlerFunc) {
	AssertNotLock()
	if h != nil {
		svr.err404Handler = h
	}
}

// Handle500 set custom error handler for status code 500
func Handle500(h func(http.ResponseWriter, *http.Request, interface{})) {
	AssertNotLock()
	if h != nil {
		svr.err500Handler = h
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
	svr.filters.add(pathPrefix, matchAll, h)
	return nil
}

// Run run the application as http
func Run(addr string) {
	svr.onInit()
	err := http.ListenAndServe(addr, svr)
	if err != nil {
		panic(err)
	}
}

// RunTLS run the application as https
func RunTLS(addr, certFile, keyFile string) {
	svr.onInit()
	err := http.ListenAndServeTLS(addr, certFile, keyFile, svr)
	if err != nil {
		panic(err)
	}
}
