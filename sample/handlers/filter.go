package handlers

import "github.com/Simbory/mego"

func testFilter(ctx *mego.Context) interface{} {
	data := ctx.GetItem("user")
	if data != nil {
		return data.(string)
	}
	return nil
}