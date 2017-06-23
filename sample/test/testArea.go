package test

import (
	"github.com/Simbory/mego"
	"github.com/Simbory/mego/view"
	"fmt"
)

var testArea = mego.GetArea("test")

func getUpload(_ *mego.Context) interface{} {
	return view.RenderAreaView(testArea,"upload", nil)
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
	testArea.Get("upload", getUpload)
	testArea.Post("upload", postUpload)
	view.UseAreaView(testArea, "views", ".html")
}