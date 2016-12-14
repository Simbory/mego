package view

import "html/template"

func include(viewName string, data interface{}) template.HTML {
	viewBytes,err :=  viewSingleton.renderView(viewName, data)
	if err != nil {
		panic(err)
	}
	return template.HTML(viewBytes)
}