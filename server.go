package mego

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"strings"
	"errors"
	"sync"
)

type routeSetting struct {
	method     string
	routePath  string
	reqHandler ReqHandler
}

type Server struct {
	locked        bool
	routing       *routeTree
	initEvents    []func()
	staticDirs    map[string]http.Handler
	staticFiles   map[string]string
	err404Handler http.HandlerFunc
	err500Handler ErrHandler
	filters       filterContainer
	routeSettings []*routeSetting
	maxFormSize   int64
	webRoot       string
	viewEngine    *ViewEngine
	engineLock    *sync.RWMutex
}

func NewServer(webRoot string) *Server {
	var s = &Server{
		locked: false,
		routing: newRouteTree(),
		initEvents: []func(){},
		staticDirs: make(map[string]http.Handler, 10),
		staticFiles: make(map[string]string, 10),
		err404Handler: handle404,
		err500Handler: handle500,
		filters: make(filterContainer),
		engineLock: &sync.RWMutex{},
		maxFormSize: 32 << 20,
		webRoot: webRoot,
	}
	return s
}

func (s *Server) assertUnlocked() {
	if s.locked {
		panic(errors.New("The s is locked."))
	}
}

func (s *Server) addRoute(m, p string, h ReqHandler) {
	s.routeSettings = append(s.routeSettings, &routeSetting{
		method: m,
		routePath: p,
		reqHandler: h,
	})
}

func (s *Server) onInit() {
	if len(s.staticDirs) > 0 {
		for pathPrefix := range s.staticDirs {
			s.staticDirs[pathPrefix] = http.FileServer(http.Dir(s.MapPath(pathPrefix)))
		}
	}
	if len(s.staticFiles) > 0 {
		for u := range s.staticFiles {
			s.staticFiles[u] = s.MapPath(u)
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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
				server:    s,
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

func (s *Server) flush(w http.ResponseWriter, req *http.Request, result interface{}) {
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

func (s *Server)initViewEngine() {
	if s.viewEngine == nil {
		s.engineLock.Lock()
		defer s.engineLock.Unlock()
		if s.viewEngine == nil {
			s.viewEngine = NewViewEngine(s.MapPath("views"), ".html")
		}
	}
}