package handlers

import "github.com/Simbory/mego"

func home(ctx *mego.Context) interface{} {
	return ctx.ViewResult("home", nil)
}