package mego

import (
	"bytes"
	"io"
	"net/http"
	"github.com/simbory/mego/assert"
	"strings"
)

// Result the request result interface
type Result interface {
	ExecResult(w http.ResponseWriter, r *http.Request)
}

// BufResult the buffered result
type BufResult struct {
	buf         *bytes.Buffer
	ContentType string
	Encoding    string
	headers     map[string]string
	StatusCode  int
}

func (b *BufResult) makeBuf() {
	if b.buf == nil {
		b.buf = bytes.NewBuffer(nil)
	}
}

// Write write byte array p to the result buffer
func (b *BufResult) Write(p []byte) (n int, err error) {
	b.makeBuf()
	return b.buf.Write(p)
}

// WriteByte write byte c to the result buffer
func (b *BufResult) WriteByte(c byte) error {
	b.makeBuf()
	return b.buf.WriteByte(c)
}

// WriteString write string s to the result buffer
func (b *BufResult) WriteString(s string) (n int, err error) {
	b.makeBuf()
	return b.buf.WriteString(s)
}

// WriteRune write rune r to the result buffer
func (b *BufResult) WriteRune(r rune) (n int, err error) {
	b.makeBuf()
	return b.buf.WriteRune(r)
}

// ReadFrom read the data from reader r and write to the result buffer
func (b *BufResult) ReadFrom(r io.Reader) (n int64, err error) {
	b.makeBuf()
	return b.buf.ReadFrom(r)
}

// AddHeader write http header to the result
func (b *BufResult) AddHeader(key, value string) {
	assert.NotEmpty("key", key)
	assert.NotEmpty("value", value)
	if strings.EqualFold(key, "content-type") {
		b.ContentType = value
		return
	}
	if b.headers == nil {
		b.headers = make(map[string]string)
	}
	b.headers[key] = value
}

// ExecResult write the data in the result buffer to the response writer
func (b *BufResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	if len(b.headers) > 0 {
		for key, value := range b.headers {
			w.Header().Add(key, value)
		}
	}
	if len(b.Encoding) == 0 {
		b.Encoding = "utf-8"
	}
	if len(b.ContentType) == 0 {
		b.ContentType = "text/plain"
	}
	w.Header().Add("content-type", strAdd(b.ContentType, "; charset=", b.Encoding))
	if b.StatusCode <= 0 {
		b.StatusCode = 200
	}
	w.WriteHeader(b.StatusCode)
	if b.buf != nil {
		io.Copy(w, b.buf)
	}
}

// NewBufResult create a new buffer result with default value buf
func NewBufResult(buf *bytes.Buffer) *BufResult {
	return &BufResult{buf: buf}
}

// emptyResult the empty buffer result
type emptyResult struct {}

// ExecResult do nothing
func (er *emptyResult) ExecResult(w http.ResponseWriter, r *http.Request) {
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