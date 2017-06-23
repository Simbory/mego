package view

import (
	"github.com/Simbory/mego/viewEngine"
	"github.com/Simbory/mego"
	"path"
)

var engine *viewEngine.Engine
var areaEngines = map[string]*viewEngine.Engine{}

func UseGlobalView(viewDir, ext string) {
	if len(viewDir) == 0 || len(ext) == 0 {
		return
	}
	mego.AssertUnlocked()
	mego.OnStart(func(){
		e,err := viewEngine.NewEngine(mego.MapPath(viewDir), ext)
		if err != nil {
			panic(err)
		}
		engine = e
	})
}

func UseAreaView(area *mego.Area, viewDir, ext string) {
	if area == nil || len(viewDir) == 0 || len(ext) == 0 {
		return
	}
	mego.AssertUnlocked()
	mego.OnStart(func() {
		dirPath := path.Join(area.Dir(), viewDir)
		e,err := viewEngine.NewEngine(dirPath, ext)
		if err != nil {
			panic(err)
		}
		areaEngines[area.Key()] = e
	})
}

func RenderView(viewName string, data interface{}) mego.Result {
	if len(viewName) == 0 {
		return nil
	}
	return &result{
		viewName: viewName,
		data: data,
		engine: engine,
	}
}

func RenderAreaView(area *mego.Area, viewName string, data interface{}) mego.Result {
	if area == nil || len(viewName) == 0 {
		return nil
	}
	eg := areaEngines[area.Key()]
	if eg == nil {
		return nil
	}
	return &result{
		viewName: viewName,
		data: data,
		engine: eg,
	}
}