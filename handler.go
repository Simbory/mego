package mego

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"strings"
)

type serverHandler struct{}

func (server *serverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer intErrorHandler(w, r)

	method := strings.ToUpper(r.Method)

	handlers, routeData, err := routing.lookup(r.URL.Path)
	if err != nil {
		panic(err)
	}
	var handler ReqHandler
	var ok bool
	if handlers != nil {
		handler, ok = handlers[method]
		if !ok {
			handler, ok = handlers["*"]
		}
	}
	if handler != nil && ok {
		var ctx = &Context{
			req:       r,
			res:       w,
			routeData: routeData,
		}
		result := handler(ctx)
		if result != nil {
			server.flush(w, r, result)
			return
		}
	}
	notFoundHandler(w, r)
}

func (server *serverHandler) flush(w http.ResponseWriter, req *http.Request, result interface{}) {
	switch result.(type) {
	case Result:
		result.(Result).ExecResult(w, req)
		return
	case *url.URL:
		res := result.(*url.URL)
		http.Redirect(w, req, res.String(), 302)
		return
	case string:
		content := result.(string)
		w.Header().Add("Content-Type", "text/plain")
		w.Write(str2Byte(content))
		return
	case []byte:
		w.Header().Add("Content-Type", "text/plain")
		w.Write(result.([]byte))
		return
	case byte:
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte{result.(byte)})
		return
	default:
		var cType = req.Header.Get("Content-Type")
		var contentBytes []byte
		var err error
		if cType == "text/xml" {
			contentBytes, err = xml.Marshal(result)
			if err != nil {
				panic(err)
			}
			w.Header().Add("Content-Type", "text/xml")
		} else {
			contentBytes, err = json.Marshal(result)
			if err != nil {
				panic(err)
			}
			w.Header().Add("Content-Type", "application/json")
		}
		w.Write(contentBytes)
	}
}
