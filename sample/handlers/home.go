package handlers

import "github.com/simbory/mego"

func home(ctx *mego.HttpCtx) interface{} {
	return ctx.ViewResult("home", nil)
}