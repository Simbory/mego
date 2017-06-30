package mego

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"strings"
	"errors"
	"sync"
	"path"
	"os"
	"github.com/simbory/mego/assert"
)

type routeSetting struct {
	method     string
	routePath  string
	reqHandler ReqHandler
}

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
	maxFormSize   int64
	viewEngine    *ViewEngine
	engineLock    *sync.RWMutex
	urlSuffix     string
}

// NewServer create a new server
//
// webRoot: the root of this web server. and the content root is '${webRoot}/www'
//
// addr: the address the server is listen on
//
// maxFormSize: the max form size. the default is 32M
//
// urlSuffix: the dynamic url suffix. if this value is empty, all the dynamic url
// should be like 'https://example.com/path/to/url' or 'https://example.com/path/to/url/'
func NewServer(webRoot, addr string, maxFormSize int64, urlSuffix string) *Server {
	if maxFormSize <= 0 {
		maxFormSize = 32 << 20
	}
	webRoot = path.Clean(strings.Replace(webRoot, "\\", "/", -1))
	var s = &Server{
		webRoot: webRoot,
		contentRoot: webRoot + "/www",
		addr: addr,
		locked: false,
		routing: newRouteTree(),
		initEvents: []func(){},
		err404Handler: handle404,
		err500Handler: handle500,
		err400Handler: handle400,
		err403Handler: handle403,
		filters: make(filterContainer),
		engineLock: &sync.RWMutex{},
		maxFormSize: maxFormSize,
		urlSuffix: urlSuffix,
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

func (s *Server) processStaticRequest(w http.ResponseWriter, r *http.Request) {
	filePath := s.MapContentPath(r.URL.Path)
	stat,err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.err404Handler(w, r)
		} else {
			s.err403Handler(w, r)
		}
		return
	}
	if stat.IsDir() {
		filePath += "/index.html"
	}
	http.ServeFile(w, r, filePath)
}

func (s *Server) processDynamicRequest(w http.ResponseWriter, r *http.Request, urlPath string) interface{} {
	method := strings.ToUpper(r.Method)

	handlers, routeData, err := s.routing.lookup(urlPath)
	assert.PanicErr(err)
	var handler ReqHandler
	var ok bool
	if handlers != nil {
		handler, ok = handlers[method]
		if !ok {
			handler, ok = handlers["*"]
		}
	}
	if handler != nil && ok {
		var ctx = &HttpCtx{
			req:       r,
			res:       w,
			routeData: routeData,
			server:    s,
		}
		// auto parse post form data
		if ctx.req.Method == "POST" || ctx.req.Method == "PUT" || ctx.req.Method == "PATCH" {
			err := ctx.parseForm()
			assert.PanicErr(err)
		}
		s.filters.exec(urlPath, ctx)
		if ctx.ended {
			return &emptyResult{}
		}
		return handler(ctx)
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
			assert.PanicErr(err)
			w.Header().Add("Content-Type", "text/xml")
		} else {
			contentBytes, err = json.Marshal(result)
			assert.PanicErr(err)
			w.Header().Add("Content-Type", "application/json")
		}
		w.Write(contentBytes)
	}
}

func (s *Server) initViewEngine() {
	if s.viewEngine == nil {
		s.engineLock.Lock()
		defer s.engineLock.Unlock()
		if s.viewEngine == nil {
			s.viewEngine = NewViewEngine(s.MapRootPath("views"), ".html")
		}
	}
}

func (s *Server) isDynamic(urlPath, cleanUrlPath string) bool {
	if len(urlPath) == 0 || urlPath == "/" {
		return true
	}
	if len(s.urlSuffix) > 0 {
		return strings.HasSuffix(urlPath, s.urlSuffix)
	}
	index1 := strings.LastIndex(urlPath, ".")
	if index1 < 0 {
		return true
	}
	return index1 < strings.LastIndex(cleanUrlPath, "/")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// catch the panic error
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

	var urlPath string
	var urlCleanPath string
	if r.URL.Path == "/" {
		urlPath = "/"
		urlCleanPath = "/"
	} else {
		urlCleanPath = path.Clean(r.URL.Path)
		if strings.HasSuffix(r.URL.Path, "/") {
			urlPath = urlCleanPath + "/"
		} else {
			urlPath = urlCleanPath
		}
	}
	isDynamic := s.isDynamic(urlPath, urlCleanPath)
	if isDynamic {
		if urlCleanPath != "/" && len(s.urlSuffix) > 0 {
			urlCleanPath = strings.TrimRight(urlCleanPath, s.urlSuffix)
		}
		var result= s.processDynamicRequest(w, r, urlCleanPath)
		if result != nil {
			s.flush(w, r, result)
		} else {
			s.err404Handler(w, r)
		}
	} else {
		s.processStaticRequest(w, r)
	}
}