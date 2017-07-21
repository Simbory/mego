package mego

import (
	"github.com/simbory/mego/assert"
	"github.com/simbory/mego/views"
	"net/http"
)

// viewResult the view result struct
type viewResult struct {
	viewName string
	data     interface{}
	engine   *views.ViewEngine
}

// ExecResult execute the view and write the view result to the response writer
func (vr *viewResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	err := vr.engine.Render(w, vr.viewName, vr.data)
	assert.PanicErr(err)
}
