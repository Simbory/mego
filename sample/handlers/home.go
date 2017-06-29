package handlers

import "github.com/simbory/mego"

func home(ctx *mego.Context) interface{} {
	return ctx.ViewResult("home", nil)
}