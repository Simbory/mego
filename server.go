package mego

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/simbory/mego/assert"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"path/filepath"
)

type routeSetting struct {
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
	hijackColl    hijackContainer
	routeSettings []*routeSetting
	viewEngine    *ViewEngine
	engineLock    sync.RWMutex
	ctxId         uint64
	serverVar     map[string]interface{}
}

// assertUnlocked assert that the server is not running
func (s *Server) assertUnlocked() {
	if s.locked {
		panic(errors.New("the server is locked"))
	}
}

func (s *Server) addRoute(path string, handler interface{}) {
	s.routeSettings = append(s.routeSettings, &routeSetting{
		routePath: path,
		processor: handler,
		area:      nil,
	})
}

func (s *Server) addAreaRoute(p string, area *Area, h interface{}) {
	assert.NotNil("area", area)
	assert.NotNil("h", h)
	assert.NotEmpty("p", p)
	s.routeSettings = append(s.routeSettings, &routeSetting{
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
			s.routing.addRoute(setting.routePath, setting.area, setting.processor)
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

func findHandler(handler interface{}, method string) (func(ctx *HttpCtx)interface{}, bool) {
	switch method {
	case "GET":
		h, ok := handler.(RouteGet)
		if ok {
			return h.Get, ok
		}
	case "POST":
		h, ok := handler.(RoutePost)
		if ok {
			return h.Post, true
		}
	case "PUT":
		h, ok := handler.(RoutePut)
		if ok {
			return h.Put, true
		}
	case "OPTIONS":
		h, ok := handler.(RouteOptions)
		if ok {
			return h.Options, true
		}
	case "DELETE":
		h, ok := handler.(RouteDelete)
		if ok {
			return h.Delete, true
		}
	case "TRACE":
		h, ok := handler.(RouteTrace)
		if ok {
			return h.Trace, true
		}
	case "PATCH":
		h, ok := handler.(RoutePatch)
		if ok {
			return h.Patch, true
		}
	case "CONNECT":
		h, ok := handler.(RouteConnect)
		if ok {
			return h.Connect, true
		}
	case "HEAD":
		h, ok := handler.(RouteHead)
		if ok {
			return h.Head, true
		}
	case "PROPFIND":
		h, ok := handler.(RoutePropFind)
		if ok {
			return h.PropFind, true
		}
	case "PROPPATCH":
		h, ok := handler.(RoutePropPatch)
		if ok {
			return h.PropPatch, true
		}
	case "MKCOL":
		h, ok := handler.(RouteMkcol)
		if ok {
			return h.Mkcol, true
		}
	case "COPY":
		h, ok := handler.(RouteCopy)
		if ok {
			return h.Copy, true
		}
	case "MOVE":
		h, ok := handler.(RouteMove)
		if ok {
			return h.Move, true
		}
	case "LOCK":
		h, ok := handler.(RouteLock)
		if ok {
			return h.Lock, true
		}
	case "UNLOCK":
		h, ok := handler.(RouteUnlock)
		if ok {
			return h.Unlock, true
		}
	}
	h, ok := handler.(RouteProcessor)
	if ok {
		return h.ProcessRequest, true
	} else {
		return nil, false
	}
}

func (s *Server) processDynamicRequest(w http.ResponseWriter, r *http.Request, urlPath string) interface{} {
	method := strings.ToUpper(r.Method)
	handler, routeData, area, err := s.routing.lookup(urlPath)
	assert.PanicErr(err)
	var processor func(ctx *HttpCtx)interface{}
	var filterFunc func(ctx *HttpCtx)
	if handler == nil {
		return nil
	}
	handlerFunc,ok := handler.(func(ctx *HttpCtx)interface{})
	if ok {
		processor = handlerFunc
	} else {
		p, ok := findHandler(handler, method)
		if ok {
			filter, ok := handler.(RouteFilter)
			if ok {
				filterFunc = filter.Filter
			}
			processor = p
		} else {
			return nil
		}
	}
	if processor != nil {
		ctxId := atomic.AddUint64(&(s.ctxId), 1)
		var ctx = &HttpCtx{
			req:       r,
			res:       w,
			routeData: routeData,
			Server:    s,
			area:      area,
			ctxId:     ctxId,
		}
		if area != nil {
			area.hijackColl.exec(urlPath, ctx)
		} else {
			s.hijackColl.exec(urlPath, ctx)
		}
		if ctx.ended {
			return &emptyResult{}
		}
		if filterFunc != nil {
			filterFunc(ctx)
			if ctx.ended {
				return &emptyResult{}
			}
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
			s.viewEngine = NewViewEngine(s.MapRootPath("views"))
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
