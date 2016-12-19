package view

import (
	"html/template"
	"bytes"
)

func include(viewName string, data interface{}) template.HTML {
	buf := &bytes.Buffer{}
	err :=  singleton.renderView(buf, viewName, data)
	if err != nil {
		panic(err)
	}
	return template.HTML(buf.Bytes())
}