package mego

import (
	"net/http"
	"strings"
	"bytes"
	"runtime/debug"
)

// ErrHandler define the internal server error handler func
type ErrHandler func(http.ResponseWriter, *http.Request, interface{})

// handle404 the default error 404 handler
func handle404(w http.ResponseWriter, r *http.Request) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("<h3>Error 404: Not Found</h3>")
	buf.WriteString("<p>The page you are looking for is not found: <i>" + r.URL.String() + "</i></p>")
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(404)
	w.Write(buf.Bytes())
}

// handle500 the default error 500 handler
func handle500(w http.ResponseWriter, r *http.Request, rec interface{}) {
	var debugStack = string(debug.Stack())
	debugStack = strings.Replace(debugStack, "<", "&lt;", -1)
	debugStack = strings.Replace(debugStack, ">", "&gt;", -1)
	buf := &bytes.Buffer{}
	buf.Write(str2Byte("<pre><code>"))
	if err, ok := rec.(error); ok {
		buf.WriteString(err.Error())
		buf.WriteString("\r\n\r\n")
	}
	buf.WriteString(debugStack)
	buf.WriteString("</code></pre>")
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(500)
	w.Write(buf.Bytes())
}