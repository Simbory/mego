package mego

import (
	"bytes"
	"errors"
	"net/http"
	"runtime/debug"
	"strings"
)

var (
	locked  = false
	routing = newRouteTree()
	RootDir = workingDir()
	initEvents = []func(){}
	notFoundHandler http.HandlerFunc = handle404
	intErrorHandler http.HandlerFunc = handle500
)

func AssertLock() {
	if locked {
		panic(errors.New("Cannot call this function while the server is runing."))
	}
}

func handle404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write([]byte("Error 404: Not Found"))
}

func handle500(w http.ResponseWriter, r *http.Request) {
	rec := recover()
	if rec == nil {
		return
	}
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

func OnInit(h func()) {
	AssertLock()
	if h != nil {
		initEvents = append(initEvents, h)
	}
}

// AddFunc add route validation func
func AddRouteFunc(name string, fun RouteFunc) {
	AssertLock()
	routing.addFunc(name, fun)
}

func Get(routePath string, handler ReqHandler) {
	AssertLock()
	routing.addRoute("GET", routePath, handler)
}

func Post(routePath string, handler ReqHandler) {
	AssertLock()
	routing.addRoute("POST", routePath, handler)
}

func Put(routePath string, handler ReqHandler) {
	AssertLock()
	routing.addRoute("PUT", routePath, handler)
}

func Options(routePath string, handler ReqHandler) {
	AssertLock()
	routing.addRoute("OPTIONS", routePath, handler)
}

func Head(routePath string, handler ReqHandler) {
	AssertLock()
	routing.addRoute("HEAD", routePath, handler)
}

func Delete(routePath string, handler ReqHandler) {
	AssertLock()
	routing.addRoute("DELETE", routePath, handler)
}

func Trace(routePath string, handler ReqHandler) {
	AssertLock()
	routing.addRoute("RACE", routePath, handler)
}

func Connect(routePath string, handler ReqHandler) {
	AssertLock()
	routing.addRoute("CONNECT", routePath, handler)
}

func Any(routePath string, handler ReqHandler) {
	AssertLock()
	routing.addRoute("*", routePath, handler)
}

func Handle404(h http.HandlerFunc) {
	if h != nil {
		notFoundHandler = h
	}
}

func Handle500(h http.HandlerFunc) {
	if h != nil {
		intErrorHandler = h
	}
}

func Run(addr string) {
	initMego()
	svr := &serverHandler{}
	err := http.ListenAndServe(addr, svr)
	if err != nil {
		panic(err)
	}
}

func RunTLS(addr, certFile, keyFile string) {
	initMego()
	svr := &serverHandler{}
	err := http.ListenAndServeTLS(addr, certFile, keyFile, svr)
	if err != nil {
		panic(err)
	}
}
