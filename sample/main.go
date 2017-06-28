package main

import (
	"github.com/Simbory/mego"
	"github.com/Simbory/mego/cache"
	"github.com/Simbory/mego/session"
	"github.com/Simbory/mego/sample/admin"
	"github.com/Simbory/mego/sample/handlers"
	"github.com/Simbory/mego/sample/filters"
)

func main() {
	server := mego.NewServer(mego.WorkingDir(), ":8080", 0, "")

	cache.UseDefault()
	provider := session.NewDiskProvider(server.MapWebRoot("/temp/sessions"))
	mgr := session.CreateManager(server,nil, provider)
	session.UseAsDefault(mgr)

	handlers.Init(server)
	filters.Init(server)
	admin.Init(server)

	server.Run()
}