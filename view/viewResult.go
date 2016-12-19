package view

import (
	"net/http"
)

type viewResult struct {
	viewName string
	data     interface{}
}

func (vr *viewResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	err := Render(w, vr.viewName, vr.data)
	if err != nil {
		panic(err)
	}
}