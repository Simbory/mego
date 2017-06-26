package mego

import (
	"github.com/Simbory/mego/views"
	"errors"
)

type ViewEngine struct {
	engine *views.ViewEngine
	viewDir string
	viewExt string
}

func (e *ViewEngine) Render(viewName string, data interface{}) Result {
	if e.engine == nil {
		func() {
			eg,err := views.NewEngine(MapPath(e.viewDir), e.viewExt)
			if err != nil {
				panic(err)
			}
			e.engine = eg
		}()
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

func NewViewEngine(viewDir, viewExt string) *ViewEngine {
	if len(viewDir) == 0 {
		panic(errors.New("The view directory cannot be empty"))
	}
	engine := &ViewEngine{engine: nil, viewDir: viewDir, viewExt: viewExt}
	return engine
}