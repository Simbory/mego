package handlers

import (
	"github.com/Simbory/mego"
)

func Init() {
	mego.Get("/", home)
	mego.Any("/views/*pathInfo", renderView)
	mego.Get("/date/<year:int>-<month:int>-<day:int>", getDate)
	mego.Get("/session", testSession)
	mego.Get("/uuid", testUUID)
	mego.Any("/filter/*pathInfo", testFilter)
}