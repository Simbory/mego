package test

import (
	"github.com/Simbory/mego"
	"github.com/Simbory/mego/view"
	"fmt"
)

var area = mego.GetArea("test")
var viewEngine *view.ViewEngine

func getUpload(_ *mego.Context) interface{} {
	return viewEngine.RenderView("upload", nil)
}

func postUpload(ctx *mego.Context) interface{} {
	file := ctx.PostFile("file")
	filePath := mego.MapPath(file.FileName)
	err := file.SaveAndClose(filePath)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("file Size: %d", file.Size)
}

func InitArea() {
	area.Get("upload", getUpload)
	area.Post("upload", postUpload)
	viewEngine = view.NewViewEngine(area.Key() + "/views", ".html")
}