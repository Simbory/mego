package mego

import (
	"net/http"
	"strings"
	"bytes"
	"runtime/debug"
)

// Error500Handler define the internal server error handler func
type Error500Handler func(http.ResponseWriter, *http.Request, interface{})

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