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
	viewFunc template.FuncMap
}

// Render render the view 'viewName' with 'data' and get the view result
func (e *ViewEngine) Render(viewName string, data interface{}) Result {
	if e.engine == nil {
		func(e *ViewEngine) {
			eg, err := views.NewEngine(e.viewDir, ".gohtml")
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

// ExtendView extend the view helper functions with 'name' and 'viewFunc'
func (e *ViewEngine) ExtendView(name string, viewFunc interface{}) {
	if len(name) == 0 || viewFunc == nil {
		return
	}
	e.viewFunc[name] = viewFunc
}

// NewViewEngine create a new view engine in ViewDir with file extension '.gohtml'
func NewViewEngine(viewDir string) *ViewEngine {
	assert.NotEmpty("viewDir", viewDir)
	return &ViewEngine{
		engine:   nil,
		viewDir:  viewDir,
		viewFunc: make(template.FuncMap),
	}
}