package admin

import (
	"github.com/Simbory/mego"
	"fmt"
)

func getUpload(_ *mego.Context) interface{} {
	return view.Render("upload", nil)
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

var area = mego.GetArea("admin")
var view *mego.ViewEngine

func Init() {
	area.Get("upload", getUpload)
	area.Post("upload", postUpload)
	view = mego.NewViewEngine(area.Key() + "/views", ".html")
}