package mego

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/simbory/mego/assert"
	"net/http"
	"regexp"
	"net/url"
)

type sizer interface {
	Size() int64
}

// HttpCtx the mego http context struct
type HttpCtx struct {
	req         *http.Request
	res         http.ResponseWriter
	routeData   map[string]string
	ended       bool
	ctxItems    map[string]interface{}
	area        *Area
	ctxId       uint64
	queryValues *url.Values

	Server *Server
}

// CtxId get the http context id
func (ctx *HttpCtx) CtxId() uint64 {
	return ctx.ctxId
}

// Request get the mego request
func (ctx *HttpCtx) Request() *http.Request {
	return ctx.req
}

// Response get the mego response
func (ctx *HttpCtx) Response() http.ResponseWriter {
	return ctx.res
}

// QueryStr get the value from the url query string
func (ctx *HttpCtx) QueryStr(key string) string {
	if ctx.queryValues == nil {
		query := ctx.req.URL.Query()
		ctx.queryValues = &query
	}
	return ctx.queryValues.Get(key)
}

// FormValue get the form value from request. It's the same as ctx.Request().FormValue(key)
func (ctx *HttpCtx) FormValue(key string) string {
	return ctx.req.FormValue(key)
}

// RouteString get the route parameter value as string by key
func (ctx *HttpCtx) RouteVar(key string) string {
	if ctx.routeData == nil {
		return ""
	}
	return ctx.routeData[key]
}

// PostFile get the post file info
func (ctx *HttpCtx) PostFile(formName string) *UploadFile {
	f, h, err := ctx.Request().FormFile(formName)
	if err != nil {
		return &UploadFile{Error: err}
	}
	if f == nil {
		return nil
	}
	var size int64
	if s, ok := f.(sizer); ok {
		size = s.Size()
	} else {
		size = 0
	}
	return &UploadFile{FileName: h.Filename, Size: size, File: f, Header: h}
}

// SetCtxItem add context data to mego context
func (ctx *HttpCtx) SetCtxItem(key string, data interface{}) {
	if len(key) == 0 {
		return
	}
	if data == nil {
		ctx.RemoveCtxItem(key)
		return
	}
	if ctx.ctxItems == nil {
		ctx.ctxItems = make(map[string]interface{})
	}
	ctx.ctxItems[key] = data
}

// GetCtxItem get the context data from mego context by key
func (ctx *HttpCtx) GetCtxItem(key string) interface{} {
	if ctx.ctxItems == nil {
		return nil
	}
	return ctx.ctxItems[key]
}

// RemoveCtxItem delete context item from mego context by key
func (ctx *HttpCtx) RemoveCtxItem(key string) {
	if ctx.ctxItems == nil {
		return
	}
	ctx.ctxItems[key] = nil
	delete(ctx.ctxItems, key)
}

// MapRootPath Returns the physical file path that corresponds to the specified virtual path.
func (ctx *HttpCtx) MapRootPath(path string) string {
	return ctx.Server.MapRootPath(path)
}

// MapContentPath Returns the physical file path that corresponds to the specified virtual path.
func (ctx *HttpCtx) MapContentPath(urlPath string) string {
	return ctx.Server.MapContentPath(urlPath)
}

// TextResult generate the mego result as plain text
func (ctx *HttpCtx) TextResult(content, contentType string) Result {
	result := NewBufResult(nil)
	result.WriteString(content)
	result.ContentType = contentType
	return result
}

// JsonResult generate the mego result as JSON string
func (ctx *HttpCtx) JsonResult(data interface{}) Result {
	dataJSON, err := json.Marshal(data)
	assert.PanicErr(err)
	return ctx.TextResult(byte2Str(dataJSON), "application/json")
}

// JsonpResult generate the mego result as jsonp string
func (ctx *HttpCtx) JsonpResult(data interface{}, callback string) Result {
	reg := regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")
	if !reg.Match(str2Byte(callback)) {
		panic(fmt.Errorf("invalid JSONP callback name '%s'", callback))
	}
	dataJSON, err := json.Marshal(data)
	assert.PanicErr(err)
	return ctx.TextResult(strAdd(callback, "(", byte2Str(dataJSON), ");"), "text/javascript")
}

// XmlResult generate the mego result as XML string
func (ctx *HttpCtx) XmlResult(data interface{}) Result {
	xmlBytes, err := xml.Marshal(data)
	assert.PanicErr(err)
	return ctx.TextResult(byte2Str(xmlBytes), "text/xml")
}

// FileResult generate the mego result as file result
func (ctx *HttpCtx) FileResult(path string, contentType string) Result {
	var resp = &FileResult{
		FilePath:    path,
		ContentType: contentType,
	}
	return resp
}

// ViewResult find the view by view name, execute the view template and get the result
func (ctx *HttpCtx) ViewResult(viewName string, data interface{}) Result {
	if ctx.area != nil {
		ctx.area.initViewEngine()
		return ctx.area.viewEngine.Render(viewName, data)
	} else {
		ctx.Server.initViewEngine()
		return ctx.Server.viewEngine.Render(viewName, data)
	}
}

func (ctx *HttpCtx) Redirect(urlStr string, permanent bool) {
	if permanent {
		http.Redirect(ctx.res, ctx.req, urlStr, 301)
	} else {
		http.Redirect(ctx.res, ctx.req, urlStr, 302)
	}
	ctx.End()
}

// Redirect get the redirect result. if the value of 'permanent' is true ,
// the status code is 301, else the status code is 302
func (ctx *HttpCtx) RedirectResult(urlStr string, permanent bool) Result {
	if permanent {
		return &RedirectResult{
			StatusCode:  301,
			RedirectURL: urlStr,
		}
	} else {
		return &RedirectResult{
			StatusCode:  302,
			RedirectURL: urlStr,
		}
	}
}

// End end the mego context and stop the rest request function
func (ctx *HttpCtx) End() {
	ctx.ended = true
	panic(&endCtxSignal{})
}