package handlers

import "github.com/simbory/mego"

func testFilter(ctx *mego.HttpCtx) interface{} {
	data := ctx.GetCtxItem("user")
	if data != nil {
		return data.(string)
	}
	return nil
}
