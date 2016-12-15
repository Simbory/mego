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
	locked                           = false
	routing                          = newRouteTree()
	initEvents                       = []func(){}
	staticDirs                       = make(map[string]http.Handler)
	staticFiles                      = make(map[string]string)
	notFoundHandler http.HandlerFunc = handle404
	intErrorHandler Error500Handler  = handle500
	filters                          = make(filterContainer)
)

// AssertNotLock make sure the server is not running and the function Run/RunTLS is not called.
func AssertNotLock() {
	if locked {
		panic(errors.New("cannot call this function while the server is runing"))
	}
}

func handle404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write([]byte("Error 404: Not Found"))
}

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

func initMego() {
	if len(initEvents) > 0 {
		for _, h := range initEvents {
			h()
		}
	}
}

// OnStart add a handler to the server start event
func OnStart(h func()) {
	AssertNotLock()
	if h != nil {
		initEvents = append(initEvents, h)
	}
}

// AddRouteFunc add route validation func
func AddRouteFunc(name string, fun RouteFunc) {
	AssertNotLock()
	reg := regexp.MustCompile("^[a-zA-Z_][\\w]*$")
	if !reg.Match([]byte(name)) {
		panic(fmt.Errorf("Invalid route func name: %s", name))
	}
	routing.addFunc(name, fun)
}

// Get used to register route for GET method
func Get(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("GET", routePath, handler)
}

// Post used to register route for POST method
func Post(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("POST", routePath, handler)
}

// Put used to register route for PUT method
func Put(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("PUT", routePath, handler)
}

// Options used to register route for OPTIONS method
func Options(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("OPTIONS", routePath, handler)
}

// Head used to register route for HEAD method
func Head(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("HEAD", routePath, handler)
}

// Delete used to register route for DELETE method
func Delete(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("DELETE", routePath, handler)
}

// Trace used to register route for TRACE method
func Trace(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("TRACE", routePath, handler)
}

// Connect used to register route for CONNECT method
func Connect(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("CONNECT", routePath, handler)
}

// Any used to register route for all methods
func Any(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("*", routePath, handler)
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
	staticDirs[pathPrefix] = http.FileServer(http.Dir(dirPath))
}

// HandleStaticFile handle the url as static file
func HandleStaticFile(url, filePath string) {
	AssertNotLock()
	staticFiles[url] = filePath
}

// Handle404 handle error 404
func Handle404(h http.HandlerFunc) {
	AssertNotLock()
	if h != nil {
		notFoundHandler = h
	}
}

// Handle500 handle error 500
func Handle500(h func(http.ResponseWriter, *http.Request, interface{})) {
	AssertNotLock()
	if h != nil {
		intErrorHandler = h
	}
}

// Filter set path filter
func Filter(pathPrefix string, h func(*Context)) {
	AssertNotLock()
	if len(pathPrefix) == 0 {
		panic(errors.New("The parameter 'pathPrefix' cannot be empty"))
	}
	if h == nil {
		panic(errors.New("The parameter 'h' cannot be nil"))
	}
	if !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = "/" + pathPrefix
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	filters.add(pathPrefix, h)
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
