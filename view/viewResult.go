package view

import (
	"net/http"
)

type viewResult struct {
	viewName string
	data     interface{}
}

func (vr *viewResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	resultBytes, err := Render(vr.viewName, vr.data)
	if err != nil {
		panic(err)
	}
	w.Header().Add("Content-Type", "utf-8")
	w.Write(resultBytes)
}