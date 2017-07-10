package mego

import (
	"html/template"

	"github.com/simbory/mego/assert"
	"github.com/simbory/mego/views"
)

// ViewEngine the mego view engine struct
type ViewEngine struct {
	engine   *views.ViewEngine
	viewDir  string
	viewExt  string
	viewFunc template.FuncMap
}

// Render get the template by viewName and then generate the mego result based on the template
func (e *ViewEngine) Render(viewName string, data interface{}) Result {
	if e.engine == nil {
		func(e *ViewEngine) {
			eg, err := views.NewEngine(e.viewDir, e.viewExt)
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
		data:     data,
		engine:   e.engine,
	}
}

// Extend the view template by view func
func (e *ViewEngine) ExtendView(name string, viewFunc interface{}) {
	if len(name) == 0 || viewFunc == nil {
		return
	}
	e.viewFunc[name] = viewFunc
}

// NewViewEngine create a new view engine. the view files is located at viewDir and the view file extension is viewExt.
func NewViewEngine(viewDir, viewExt string) *ViewEngine {
	assert.NotEmpty("viewDir", viewDir)
	assert.NotEmpty("viewExt", viewExt)
	return &ViewEngine{
		engine:   nil,
		viewDir:  viewDir,
		viewExt:  viewExt,
		viewFunc: make(template.FuncMap),
	}
}
