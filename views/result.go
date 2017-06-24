package views

import (
	"net/http"
	"github.com/Simbory/mego/viewEngine"
)

type result struct {
	viewName string
	data     interface{}
	engine   *viewEngine.Engine
}

func (vr *result) ExecResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	err := vr.engine.Render(w, vr.viewName, vr.data)
	if err != nil {
		panic(err)
	}
}