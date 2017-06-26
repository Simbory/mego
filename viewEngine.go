package mego

import (
	"github.com/Simbory/mego/views"
	"errors"
	"html/template"
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
			eg,err := views.NewEngine(MapPath(e.viewDir), e.viewExt)
			if err != nil {
				panic(err)
			}
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
	if len(viewDir) == 0 {
		panic(errors.New("The view directory cannot be empty"))
	}
	engine := &ViewEngine{engine: nil, viewDir: viewDir, viewExt: viewExt, viewFunc: make(template.FuncMap)}
	return engine
}