package mego

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"strings"
	"errors"
	"path"
)

type routeSetting struct {
	method     string
	routePath  string
	reqHandler ReqHandler
}

type webServer struct {
	locked        bool
	routing       *routeTree
	initEvents    []func()
	staticDirs    map[string]http.Handler
	staticFiles   map[string]string
	err404Handler http.HandlerFunc
	err500Handler Error500Handler
	filters       filterContainer
	routeSettings []*routeSetting
	maxFormSize   int64
	webRoot       string
}

func newServer() *webServer {
	var s = &webServer{
		locked: false,
		routing: newRouteTree(),
		initEvents: []func(){},
		staticDirs: make(map[string]http.Handler, 10),
		staticFiles: make(map[string]string, 10),
		err404Handler: handle404,
		err500Handler: handle500,
		filters: make(filterContainer),
	}
	return s
}

func (s *webServer) assertUnlocked() {
	if s.locked {
		panic(errors.New("The server is locked."))
	}
}

func (s *webServer) addRoute(m, p string, h ReqHandler) {
	s.routeSettings = append(s.routeSettings, &routeSetting{
		method: m,
		routePath: p,
		reqHandler: h,
	})
}

func (s *webServer) onInit() {
	if len(server.staticDirs) > 0 {
		for pathPrefix := range server.staticDirs {
			server.staticDirs[pathPrefix] = http.FileServer(http.Dir(MapPath(pathPrefix)))
		}
	}
	if len(server.staticFiles) > 0 {
		for u := range server.staticFiles {
			server.staticFiles[u] = MapPath(u)
		}
	}
	if len(s.routeSettings) > 0 {
		for _, setting := range s.routeSettings {
			s.routing.addRoute(setting.method, setting.routePath, setting.reqHandler)
		}
	}
	if len(s.initEvents) > 0 {
		for _, h := range s.initEvents {
			h()
		}
	}
}

func (s *webServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		s.err500Handler(w, r, rec)
		rec1 := recover()
		if rec1 != nil {
			handle500(w, r, rec)
		}
	}()
	var result interface{}
	if len(s.staticFiles) > 0 {
		for urlPath, filePath := range s.staticFiles {
			if r.URL.Path == urlPath {
				result = &FileResult{
					FilePath: filePath,
				}
				break
			}
		}
	}
	if result == nil && len(s.staticDirs) > 0 {
		urlWithSlash := r.URL.Path
		if !strings.HasSuffix(urlWithSlash, "/") {
			urlWithSlash = urlWithSlash + "/"
		}
		for pathPrefix, h := range s.staticDirs {
			if strings.HasPrefix(urlWithSlash, pathPrefix) {
				r.URL.Path = strings.TrimLeft(r.URL.Path, pathPrefix[0:len(pathPrefix)-1])
				h.ServeHTTP(w, r)
				return
			}
		}
	}
	if result == nil {
		method := strings.ToUpper(r.Method)
		handlers, routeData, err := s.routing.lookup(r.URL.Path)
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
			err := ctx.parseForm()
			if err != nil {
				panic(err)
			}
			s.filters.exec(r.URL.Path, ctx)
			if ctx.ended {
				return
			}
			result = handler(ctx)
		}
	}
	if result != nil {
		s.flush(w, r, result)
	} else {
		s.err404Handler(w, r)
	}
}

func (s *webServer) flush(w http.ResponseWriter, req *http.Request, result interface{}) {
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

func (s *webServer) mapPath(virtualPath string) string {
	p := path.Join(s.webRoot, virtualPath)
	p = path.Clean(strings.Replace(p, "\\", "/", -1))
	return strings.TrimRight(p, "/")
}