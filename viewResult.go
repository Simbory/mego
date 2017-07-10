package mego

import (
	"net/http"

	"github.com/simbory/mego/assert"
	"github.com/simbory/mego/views"
)

// viewResult the view result struct
type viewResult struct {
	viewName string
	data     interface{}
	engine   *views.ViewEngine
}

// ExecResult execute the template and then write the result to the response writer
func (vr *viewResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	err := vr.engine.Render(w, vr.viewName, vr.data)
	assert.PanicErr(err)
}
