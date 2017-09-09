package mego

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/simbory/mego/assert"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"path/filepath"
)

type routeSetting struct {
	method    string
	routePath string
	processor interface{}
	area      *Area
}

type endCtxSignal struct{}

type Server struct {
	webRoot       string
	contentRoot   string
	addr          string
	locked        bool
	routing       *routeTree
	initEvents    []func()
	err404Handler http.HandlerFunc
	err500Handler ErrHandler
	err400Handler http.HandlerFunc
	err403Handler http.HandlerFunc
	filters       filterContainer
	routeSettings []*routeSetting
	viewEngine    *ViewEngine
	engineLock    *sync.RWMutex
	ctxId         uint64
}

// NewServer create a new server
//
// webRoot: the root of this web server. and the content root is '${webRoot}/www'
//
// addr: the address the server is listen on
//
// urlSuffix: the dynamic url suffix. if this value is empty, all the dynamic url
// should be like 'https://example.com/path/to/url' or 'https://example.com/path/to/url/'
func NewServer(webRoot, addr string) *Server {
	webRoot = path.Clean(strings.Replace(webRoot, "\\", "/", -1))
	var s = &Server{
		webRoot:       webRoot,
		contentRoot:   webRoot + "/www",
		addr:          addr,
		locked:        false,
		routing:       newRouteTree(),
		initEvents:    []func(){},
		err404Handler: handle404,
		err500Handler: handle500,
		err400Handler: handle400,
		err403Handler: handle403,
		filters:       make(filterContainer),
		engineLock:    &sync.RWMutex{},
	}
	return s
}

// assertUnlocked assert that the server is not running
func (s *Server) assertUnlocked() {
	if s.locked {
		panic(errors.New("the server is locked"))
	}
}

func (s *Server) addRoute(method, path string, handler interface{}) {
	s.routeSettings = append(s.routeSettings, &routeSetting{
		method:    method,
		routePath: path,
		processor: handler,
		area:      nil,
	})
}

func (s *Server) addAreaRoute(m, p string, area *Area, h interface{}) {
	s.routeSettings = append(s.routeSettings, &routeSetting{
		method:    m,
		routePath: p,
		processor: h,
		area:      area,
	})
}

func (s *Server) onInit() {
	if len(s.initEvents) > 0 {
		for _, h := range s.initEvents {
			h()
		}
	}
	if len(s.routeSettings) > 0 {
		for _, setting := range s.routeSettings {
			s.routing.addRoute(setting.method, setting.routePath, setting.area, setting.processor)
		}
	}
}

func (s *Server) processStaticRequest(w http.ResponseWriter, r *http.Request) {
	filePath := s.MapContentPath(r.URL.Path)
	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.err404Handler(w, r)
		} else {
			s.err403Handler(w, r)
		}
		return
	}
	if !stat.IsDir() {
		http.ServeFile(w, r, filePath)
	} else {
		s.err404Handler(w, r)
	}
}

func findHandler(handler interface{}, method string) (ReqHandler, bool) {
	switch method {
	case "GET":
		h, ok := handler.(RouteGetter)
		if ok {
			return h.Get, ok
		}
	case "POST":
		h, ok := handler.(RoutePoster)
		if ok {
			return h.Post, true
		}
	case "PUT":
		h, ok := handler.(RoutePutter)
		if ok {
			return h.Put, true
		}
	case "OPTIONS":
		h, ok := handler.(RouteOptioner)
		if ok {
			return h.Options, true
		}
	case "DELETE":
		h, ok := handler.(RouteDeleter)
		if ok {
			return h.Delete, true
		}
	case "TRACE":
		h, ok := handler.(RouteTracer)
		if ok {
			return h.Trace, ok
		}
	default:
		h, ok := handler.(RouteProcessor)
		if ok {
			return h.ProcessRequest, true
		}
	}
	return nil, false
}

func (s *Server) processDynamicRequest(w http.ResponseWriter, r *http.Request, urlPath string) interface{} {
	method := strings.ToUpper(r.Method)
	handlers, routeData, area, err := s.routing.lookup(urlPath)
	assert.PanicErr(err)

	var processor ReqHandler

	if handlers != nil {
		handler, ok := handlers[method]
		if ok {
			if handler == nil {
				return nil
			}
			p, ok := handler.(ReqHandler)
			if ok {
				processor = p
			} else {
				return nil
			}
		} else {
			handler, ok := handlers["*"]
			if !ok || handler == nil {
				return nil
			}
			switch handler.(type) {
			case ReqHandler:
				processor = handler.(ReqHandler)
			default:
				p, ok := findHandler(handler, method)
				if ok {
					processor = p
				} else {
					return nil
				}
			}
		}
	}

	if processor != nil {
		ctxId := atomic.AddUint64(&(s.ctxId), 1)
		var ctx = &HttpCtx{
			req:       r,
			res:       w,
			routeData: routeData,
			server:    s,
			area:      area,
			ctxId:     ctxId,
		}
		if area != nil {
			area.fi
		}
		s.filters.exec(urlPath, ctx)
		if ctx.ended {
			return &emptyResult{}
		}
		return processor(ctx)
	}
	return nil
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
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.Write(str2Byte(content))
		return
	case []byte:
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.Write(result.([]byte))
		return
	case byte:
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte{result.(byte)})
		return
	default:
		var cType = req.Header.Get("Content-Type")
		var contentBytes []byte
		var err error
		if cType == "text/xml" {
			contentBytes, err = xml.Marshal(result)
			assert.PanicErr(err)
			w.Header().Add("Content-Type", "text/xml; charset=utf-8")
		} else {
			contentBytes, err = json.Marshal(result)
			assert.PanicErr(err)
			w.Header().Add("Content-Type", "application/json; charset=utf-8")
		}
		w.Write(contentBytes)
	}
}

func (s *Server) initViewEngine() {
	if s.viewEngine == nil {
		s.engineLock.Lock()
		defer s.engineLock.Unlock()
		if s.viewEngine == nil {
			s.viewEngine = NewViewEngine(s.MapRootPath("views"), ".gohtml")
		}
	}
}

func (s *Server) isDynamic(cleanUrlPath string) bool {
	ext := filepath.Ext(cleanUrlPath)
	return len(ext) == 0
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// catch the panic error
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if _, ok := rec.(*endCtxSignal); ok {
			return
		}
		s.err500Handler(w, r, rec)
		rec1 := recover()
		if rec1 != nil {
			handle500(w, r, rec)
		}
	}()
	var urlPath = strings.TrimRight(r.URL.Path, "/")
	if len(urlPath) == 0 {
		urlPath = "/"
	}
	isDynamic := s.isDynamic(urlPath)
	if isDynamic {
		var result = s.processDynamicRequest(w, r, urlPath)
		if result != nil {
			s.flush(w, r, result)
		} else {
			s.err404Handler(w, r)
		}
	} else {
		s.processStaticRequest(w, r)
	}
}
