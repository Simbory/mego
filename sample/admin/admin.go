package admin

import (
	"github.com/simbory/mego"
	"fmt"
	"github.com/simbory/mego/assert"
)

func getUpload(ctx *mego.HttpCtx) interface{} {
	return ctx.ViewResult("upload", nil)
}

func postUpload(ctx *mego.HttpCtx) interface{} {
	file := ctx.PostFile("file")
	filePath := ctx.MapPath(file.FileName)
	err := file.SaveAndClose(filePath)
	assert.PanicErr(err)
	return fmt.Sprintf("file Size: %d", file.Size)
}

var area *mego.Area

func Init(server *mego.Server) {
	area = server.GetArea("admin")
	area.Get("upload", getUpload)
	area.Post("upload", postUpload)
	area.HandleFilter("/*", func(ctx *mego.HttpCtx) {
		ctx.Response().Header().Add("mego-area-name", area.Key())
	})
}