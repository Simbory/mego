package view

import (
	"github.com/Simbory/mego/viewEngine"
	"github.com/Simbory/mego"
	"errors"
)

type ViewEngine struct {
	engine *viewEngine.Engine
}

func (e *ViewEngine) RenderView(viewName string, data interface{}) mego.Result {
	if e.engine == nil {
		panic(errors.New("Cannot render view before the server is started."))
	}
	if len(viewName) == 0 {
		return nil
	}
	return &result{
		viewName: viewName,
		data: data,
		engine: e.engine,
	}
}

func NewViewEngine(viewDir, viewExt string) *ViewEngine {
	if len(viewDir) == 0 {
		panic(errors.New("The view directory cannot be empty"))
	}
	engine := &ViewEngine{}
	mego.OnStart(func() {
		e,err := viewEngine.NewEngine(mego.MapPath(viewDir), viewExt)
		if err != nil {
			panic(err)
		}
		engine.engine = e
	})
	return engine
}