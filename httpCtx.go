package mego

import (
	"net/http"
	"strconv"
	"strings"
	"net/url"
	"mime/multipart"
	"mime"
	"github.com/google/uuid"
)

type sizer interface {
	Size() int64
}

// HttpCtx the mego context struct
type HttpCtx struct {
	req       *http.Request
	res       http.ResponseWriter
	routeData map[string]string
	ended     bool
	items     map[string]interface{}
	server    *Server
	area      *Area
}

// Request get the mego request
func (ctx *HttpCtx) Request() *http.Request {
	return ctx.req
}

// Response get the mego response
func (ctx *HttpCtx) Response() http.ResponseWriter {
	return ctx.res
}

// RouteString get the route parameter value as string by key
func (ctx *HttpCtx) RouteString(key string) string {
	if ctx.routeData == nil {
		return ""
	}
	return ctx.routeData[key]
}

// RouteInt get the route parameter value as int64 by key
func (ctx *HttpCtx) RouteInt(key string) int64 {
	var rawValue = ctx.RouteString(key)
	if len(rawValue) == 0 {
		return 0
	}
	value,err := strconv.ParseInt(rawValue, 0, 64)
	if err != nil {
		return 0
	}
	return value
}

// RouteUint get the route parameter value as uint64 by key
func (ctx *HttpCtx) RouteUint(key string) uint64 {
	var rawValue = ctx.RouteString(key)
	if len(rawValue) == 0 {
		return 0
	}
	value,err := strconv.ParseUint(rawValue, 0, 64)
	if err != nil {
		return 0
	}
	return value
}

// RouteFloat get the route parameter value as float by key
func (ctx *HttpCtx) RouteFloat(key string) float64 {
	var rawValue = ctx.RouteString(key)
	if len(rawValue) == 0 {
		return 0
	}
	value,err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		return 0
	}
	return value
}

// RouteBool get the route parameter value as boolean by key
func (ctx *HttpCtx) RouteBool(key string) bool {
	var rawValue = ctx.RouteString(key)
	if len(rawValue) == 0 || strings.ToLower(rawValue) == "false" || rawValue == "0"  {
		return false
	}
	return true
}

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

// SetItem add context data to mego context
func (ctx *HttpCtx) SetItem(key string, data interface{}) {
	if len(key) == 0 {
		return
	}
	if ctx.items == nil {
		ctx.items = make(map[string]interface{})
	}
	ctx.items[key] = data
}

// GetItem get the context data from mego context by key
func (ctx *HttpCtx) GetItem(key string) interface{} {
	if ctx.items == nil {
		return nil
	}
	return ctx.items[key]
}

// RemoveItem delete context item from mego context by key
func (ctx *HttpCtx) RemoveItem(key string) interface{} {
	if ctx.items == nil {
		return nil
	}
	data := ctx.items[key]
	delete(ctx.items, key)
	return data
}
func (ctx *HttpCtx) MapPath(path string) string {
	return ctx.server.MapRootPath(path)
}

// End end the mego context and stop the rest request function
func (ctx *HttpCtx) End() {
	ctx.ended = true
}

// parseForm parse the post form (both multipart and normal form)
func (ctx *HttpCtx) parseForm() error {
	isMultipart, reader, err := ctx.multipart()
	if err != nil {
		return err
	}
	if isMultipart {
		if ctx.req.MultipartForm != nil {
			return nil
		}
		if ctx.req.Form == nil {
			if err = ctx.req.ParseForm(); err != nil {
				return err
			}
		}
		f,err := reader.ReadForm(ctx.server.maxFormSize)
		if err != nil {
			return err
		}
		if ctx.req.PostForm == nil {
			ctx.req.PostForm = make(url.Values)
		}
		for k, v := range f.Value {
			ctx.req.Form[k] = append(ctx.req.Form[k], v...)
			// r.PostForm should also be populated. See Issue 9305.
			ctx.req.PostForm[k] = append(ctx.req.PostForm[k], v...)
		}
		ctx.req.MultipartForm = f
	} else {
		return ctx.req.ParseForm()
	}
	return nil
}

func (ctx *HttpCtx) multipart() (bool,*multipart.Reader,error) {
	v := ctx.req.Header.Get("Content-Type")
	if v == "" {
		return false, nil, nil
	}
	d, params, err := mime.ParseMediaType(v)
	if err != nil || d != "multipart/form-data" {
		return false, nil, nil
	}
	boundary, ok := params["boundary"]
	if !ok {
		return true, nil, http.ErrMissingBoundary
	}
	return true, multipart.NewReader(ctx.req.Body, boundary), nil
}