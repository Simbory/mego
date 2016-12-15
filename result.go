package mego

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

// Result the request result interface
type Result interface {
	ExecResult(w http.ResponseWriter, r *http.Request)
}

// FileResult the file result
type FileResult struct {
	ContentType string
	FilePath    string
}

// ExecResult execute the result
func (fr *FileResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	if len(fr.ContentType) > 0 {
		w.Header().Add("Content-Type", fr.ContentType)
	}
	http.ServeFile(w, r, fr.FilePath)
}

// RedirectResult the redirect result
type RedirectResult struct {
	RedirectURL string
	StatusCode  int
}

// ExecResult execute the redirect result
func (rr *RedirectResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	var statusCode = 301
	if rr.StatusCode != 301 {
		statusCode = 302
	}
	http.Redirect(w, r, rr.RedirectURL, statusCode)
}

// ContentResult define the action result struct
type ContentResult struct {
	Writer      *bytes.Buffer
	StatusCode  int
	ContentType string
	Encoding    string
	Headers     map[string]string
}

// Header get the content result header
func (cr *ContentResult) Header() map[string]string {
	if cr.Headers == nil {
		cr.Headers = make(map[string]string)
	}
	return cr.Headers
}

// Write write content bytes into content result
func (cr *ContentResult) Write(data []byte) {
	if cr.Writer == nil {
		cr.Writer = &bytes.Buffer{}
	}
	cr.Writer.Write(data)
}

// Output get the output bytes from content result
func (cr *ContentResult) Output() []byte {
	if cr.Writer == nil {
		return nil
	}
	return cr.Writer.Bytes()
}

// ClearHeader clear the http header in content result
func (cr *ContentResult) ClearHeader() {
	cr.Headers = nil
}

// ClearOutput clear the output buffer
func (cr *ContentResult) ClearOutput() {
	cr.Writer = nil
}

// Clear clear the http headers and output buffer
func (cr *ContentResult) Clear() {
	cr.ClearHeader()
	cr.ClearOutput()
}

// ExecResult execute the content result
func (cr *ContentResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	if cr.Headers != nil {
		for k, v := range cr.Headers {
			if k == "Content-Type" {
				continue
			}
			w.Header().Add(k, v)
		}
	}
	if len(cr.ContentType) > 0 {
		encoding := cr.Encoding
		if len(encoding) == 0 {
			encoding = "utf-8"
		}
		contentType := fmt.Sprintf("%s;charset=%s", cr.ContentType, encoding)
		w.Header().Add("Content-Type", contentType)
	}
	if cr.StatusCode != 200 {
		w.WriteHeader(cr.StatusCode)
	}
	output := cr.Output()
	if len(output) > 0 {
		w.Write(output)
	}
}

// NewResult create a blank content result
func NewResult() *ContentResult {
	return &ContentResult{
		StatusCode:  200,
		ContentType: "text/html",
		Encoding:    "utf-8",
	}
}

// Content generate the mego content result
func Content(data interface{}, cntType string) Result {
	var resp = NewResult()
	if len(cntType) < 1 {
		resp.ContentType = "text/plain"
	} else {
		resp.ContentType = cntType
	}
	switch data.(type) {
	case string:
		resp.Write(str2Byte(data.(string)))
	case []byte:
		resp.Write(data.([]byte))
	case byte:
		resp.Write([]byte{data.(byte)})
	default:
		panic(errors.New("Unsupported data type"))
	}
	return resp
}

// PlainText generate the mego result as plain text
func PlainText(content string) Result {
	return Content(content, "text/plain")
}

// Javascript generate the mego result as javascript code
func Javascript(code string) Result {
	return Content(code, "text/javascript")
}

// CSS generate the mego result as CSS code
func CSS(code string) Result {
	return Content(code, "text/css")
}

// JSON generate the mego result as JSON string
func JSON(data interface{}) Result {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return Content(byte2Str(dataJSON), "application/json")
}

// JSONP generate the mego result as jsonp string
func JSONP(data interface{}, callback string) Result {
	reg := regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")
	if !reg.Match(str2Byte(callback)) {
		panic(fmt.Errorf("Invalid JSONP callback name %s", callback))
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return Content(strAdd(callback, "(", byte2Str(dataJSON), ");"), "text/javascript")
}

// XML generate the mego result as XML string
func XML(data interface{}) Result {
	xmlBytes, err := xml.Marshal(data)
	if err != nil {
		panic(err)
	}
	return Content(byte2Str(xmlBytes), "text/xml")
}

// File generate the mego result as file result
func File(path string, cntType string) Result {
	var resp = &FileResult{
		FilePath:    path,
		ContentType: cntType,
	}
	return resp
}

func redirect(url string, statusCode int) *RedirectResult {
	var resp = &RedirectResult{
		StatusCode:  statusCode,
		RedirectURL: url,
	}
	return resp
}

// Redirect redirect as 302 status code
func Redirect(url string) Result {
	return redirect(url, 302)
}

// RedirectPermanent redirect as 301 status
func RedirectPermanent(url string) Result {
	return redirect(url, 301)
}
