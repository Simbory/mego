package mego

import (
	"net/http"
	"github.com/simbory/mego/views"
)

type viewResult struct {
	viewName string
	data     interface{}
	engine   *views.ViewEngine
}

func (vr *viewResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	err := vr.engine.Render(w, vr.viewName, vr.data)
	if err != nil {
		panic(err)
	}
}