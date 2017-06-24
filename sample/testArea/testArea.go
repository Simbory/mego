package testArea

import (
	"github.com/Simbory/mego"
	"github.com/Simbory/mego/views"
	"fmt"
)

var area = mego.GetArea("testArea")
var viewEngine *views.ViewEngine

func getUpload(_ *mego.Context) interface{} {
	return viewEngine.Render("upload", nil)
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

func Init() {
	area.Get("upload", getUpload)
	area.Post("upload", postUpload)
	viewEngine = views.NewViewEngine(area.Key() + "/views", ".html")
}