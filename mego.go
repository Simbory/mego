package mego

import (
	"github.com/simbory/mego/assert"
	"net/http"
	"path"
	"regexp"
	"strings"
)

// OnStart attach an event handler to the s start event
func (s *Server) OnStart(h func()) {
	s.assertUnlocked()
	if h != nil {
		s.initEvents = append(s.initEvents, h)
	}
}

var routeNameReg = regexp.MustCompile("^[a-zA-Z][\\w]*$")

// AddRouteFunc add route validation func
func (s *Server) AddRouteFunc(name string, fun RouteFunc) {
	s.assertUnlocked()
	assert.Assert("name", func() bool {
		return routeNameReg.Match([]byte(name))
	})
	s.routing.addFunc(name, fun)
}

// Route used to register router for all methods
func (s *Server) Route(routePath string, handler interface{}) {
	s.assertUnlocked()
	s.addRoute(routePath, handler)
}

var areaNameReg = regexp.MustCompile("[/a-zA-Z0-9_-]+")

// GetArea get or create the mego area
func (s *Server) GetArea(pathPrefix string) *Area {
	prefix := ClearPath(pathPrefix)
	assert.Assert("pathPrefix", func() bool {
		return areaNameReg.Match([]byte(prefix))
	})
	prefix = EnsurePrefix(prefix, "/")
	prefix = strings.TrimRight(prefix, "/")
	return &Area{
		pathPrefix: prefix,
		server:     s,
		hijackColl: make(hijackContainer),
	}
}

// MapRootPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func (s *Server) MapRootPath(virtualPath string) string {
	p := path.Join(s.webRoot, virtualPath)
	return path.Clean(strings.Replace(p, "\\", "/", -1))
}

// MapContentPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func (s *Server) MapContentPath(virtualPath string) string {
	p := path.Join(s.contentRoot, virtualPath)
	return path.Clean(strings.Replace(p, "\\", "/", -1))
}

// Handle404 set custom error handler for status code 404
func (s *Server) Handle404(h http.HandlerFunc) {
	s.assertUnlocked()
	if h != nil {
		s.err404Handler = h
	}
}

// Handle404 set custom error handler for status code 404
func (s *Server) Handle400(h http.HandlerFunc) {
	s.assertUnlocked()
	if h != nil {
		s.err400Handler = h
	}
}

// Handle404 set custom error handler for status code 404
func (s *Server) Handle403(h http.HandlerFunc) {
	s.assertUnlocked()
	if h != nil {
		s.err403Handler = h
	}
}

// Handle500 set custom error handler for status code 500
func (s *Server) Handle500(h func(http.ResponseWriter, *http.Request, interface{})) {
	s.assertUnlocked()
	if h != nil {
		s.err500Handler = h
	}
}

// Hijack hijack the dynamic request that starts with pathPrefix
func (s *Server) HijackRequest(pathPrefix string, h func(*HttpCtx)) {
	s.assertUnlocked()
	assert.NotEmpty("pathPrefix", pathPrefix)
	assert.NotNil("h", h)
	pathPrefix = ClearPath(pathPrefix)
	pathPrefix = EnsurePrefix(pathPrefix, "/")
	s.hijackColl.add(pathPrefix, h)
}

// ExtendView extend the view engine with func f
func (s *Server) ExtendView(name string, f interface{}) {
	s.assertUnlocked()
	s.initViewEngine()
	s.viewEngine.ExtendView(name, f)
}

// Run run the application as http
func (s *Server) Run() {
	s.onInit()
	err := http.ListenAndServe(s.addr, s)
	assert.PanicErr(err)
}

// RunTLS run the application as https
func (s *Server) RunTLS(certFile, keyFile string) {
	s.onInit()
	err := http.ListenAndServeTLS(s.addr, certFile, keyFile, s)
	assert.PanicErr(err)
}

// NewServer create a new server
//
// webRoot: the root of this web server. and the content root is '${webRoot}/www'
//
// addr: the address the server is listen on
func NewServer(webRoot, addr string) *Server {
	if len(webRoot) == 0 {
		webRoot = ExeDir()
	}
	webRoot = path.Clean(ClearPath(webRoot))
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
		hijackColl:    make(hijackContainer),
	}
	return s
}