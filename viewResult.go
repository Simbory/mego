package mego

import (
	"net/http"
	"github.com/simbory/mego/views"
	"github.com/simbory/mego/assert"
)

type viewResult struct {
	viewName string
	data     interface{}
	engine   *views.ViewEngine
}

func (vr *viewResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	err := vr.engine.Render(w, vr.viewName, vr.data)
	assert.PanicErr(err)
}