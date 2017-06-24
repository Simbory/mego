package handlers

import (
	"github.com/Simbory/mego/views"
	"github.com/Simbory/mego"
)

var view *views.ViewEngine

func Init() {
	view = views.NewViewEngine("views", ".html")
	mego.Get("/", home)
	mego.Any("/views/<view:word>", renderView)
	mego.Get("/date/<year:int>-<month:int>-<day:int>", getDate)
	mego.Get("/session", testSession)
	mego.Get("/uuid", testUUID)
	mego.Any("/filter/*pathInfo", testFilter)
}