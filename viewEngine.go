package mego

import (
	"github.com/simbory/mego/views"
	"html/template"
	"github.com/simbory/mego/assert"
)

type ViewEngine struct {
	engine   *views.ViewEngine
	viewDir  string
	viewExt  string
	viewFunc template.FuncMap
}

func (e *ViewEngine) Render(viewName string, data interface{}) Result {
	if e.engine == nil {
		func(e *ViewEngine) {
			eg,err := views.NewEngine(e.viewDir, e.viewExt)
			assert.PanicErr(err)
			for name, f := range e.viewFunc {
				eg.AddFunc(name, f)
			}
			e.engine = eg
		}(e)
	}
	if len(viewName) == 0 {
		return nil
	}
	return &viewResult{
		viewName: viewName,
		data: data,
		engine: e.engine,
	}
}

func (e *ViewEngine) ExtendView(name string, viewFunc interface{}) {
	if len(name) == 0 || viewFunc == nil {
		return
	}
	e.viewFunc[name] = viewFunc
}

func NewViewEngine(viewDir, viewExt string) *ViewEngine {
	assert.NotEmpty("viewDir", viewDir)
	assert.NotEmpty("viewExt", viewExt)
	return &ViewEngine{
		engine: nil,
		viewDir: viewDir,
		viewExt: viewExt,
		viewFunc: make(template.FuncMap),
	}
}