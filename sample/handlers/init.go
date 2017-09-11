package handlers

import (
	"github.com/simbory/mego"
	"github.com/simbory/mego/session"
)

func Init(server *mego.Server) {
	server.Route("/", home)
	server.Route("/views/*pathInfo", renderView)
	server.Route("/date/<year:int>-<month:int>-<day:int>", getDate)
	server.Route("/session", testSession)
	server.Route("/uuid", testUUID)
	server.Route("/filter/*pathInfo", testFilter)
	session.RegisterType(&userModel{})
}