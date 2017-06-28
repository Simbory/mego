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
	filePath := ctx.MapPath(file.FileName)
	err := file.SaveAndClose(filePath)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("file Size: %d", file.Size)
}

var area *mego.Area
var view *mego.ViewEngine

func Init(server *mego.Server) {
	area = server.GetArea("admin")
	area.Get("upload", getUpload)
	area.Post("upload", postUpload)
	view = mego.NewViewEngine(server.MapWebRoot(area.Key() + "/views"), ".html")
}