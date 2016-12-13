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

var (
	locked                                                                = false
	routing                                                               = newRouteTree()
	initEvents                                                            = []func(){}
	staticDirs                                                            = make(map[string]http.Handler)
	staticFiles                                                           = make(map[string]string)
	notFoundHandler http.HandlerFunc                                      = handle404
	intErrorHandler func(http.ResponseWriter, *http.Request, interface{}) = handle500
	filters                                                               = make(filterContainer)
)

func AssertNotLock() {
	if locked {
		panic(errors.New("Cannot call this function while the server is runing."))
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

func OnStart(h func()) {
	AssertNotLock()
	if h != nil {
		initEvents = append(initEvents, h)
	}
}

// AddFunc add route validation func
func AddRouteFunc(name string, fun RouteFunc) {
	AssertNotLock()
	reg := regexp.MustCompile("^[a-zA-Z_][\\w]*$")
	if !reg.Match([]byte(name)) {
		panic(fmt.Errorf("Invalid route func name: %s", name))
	}
	routing.addFunc(name, fun)
}

func Get(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("GET", routePath, handler)
}

func Post(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("POST", routePath, handler)
}

func Put(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("PUT", routePath, handler)
}

func Options(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("OPTIONS", routePath, handler)
}

func Head(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("HEAD", routePath, handler)
}

func Delete(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("DELETE", routePath, handler)
}

func Trace(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("RACE", routePath, handler)
}

func Connect(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("CONNECT", routePath, handler)
}

func Any(routePath string, handler ReqHandler) {
	AssertNotLock()
	routing.addRoute("*", routePath, handler)
}

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

func HandleStaticFile(url, filePath string) {
	AssertNotLock()
	staticFiles[url] = filePath
}

func Handle404(h http.HandlerFunc) {
	AssertNotLock()
	if h != nil {
		notFoundHandler = h
	}
}

func Handle500(h func(http.ResponseWriter, *http.Request, interface{})) {
	AssertNotLock()
	if h != nil {
		intErrorHandler = h
	}
}

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

func Run(addr string) {
	initMego()
	svr := &server{}
	err := http.ListenAndServe(addr, svr)
	if err != nil {
		panic(err)
	}
}

func RunTLS(addr, certFile, keyFile string) {
	initMego()
	svr := &server{}
	err := http.ListenAndServeTLS(addr, certFile, keyFile, svr)
	if err != nil {
		panic(err)
	}
}
