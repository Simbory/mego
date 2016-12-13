package mego

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"strings"
)

type server struct{}

func (server *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		intErrorHandler(w, r, rec)
	}()
	var result interface{}
	if len(staticFiles) > 0 {
		for urlPath, filePath := range staticFiles {
			if r.URL.Path == urlPath {
				result = &FileResult{
					FilePath: filePath,
				}
				break
			}
		}
	}
	if result == nil && len(staticDirs) > 0 {
		for pathPrefix, h := range staticDirs {
			if strings.HasPrefix(r.URL.Path, pathPrefix) {
				h.ServeHTTP(w, r)
				return
			}
		}
	}
	if result == nil {
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
			filters.exec(r.URL.Path, ctx)
			if ctx.ended {
				return
			}
			result = handler(ctx)
		}
	}
	if result != nil {
		server.flush(w, r, result)
	} else {
		notFoundHandler(w, r)
	}
}

func (server *server) flush(w http.ResponseWriter, req *http.Request, result interface{}) {
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
