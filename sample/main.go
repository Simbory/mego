package main

import (
	"github.com/simbory/mego"
	"github.com/simbory/mego/cache"
	"github.com/simbory/mego/session"
	"github.com/simbory/mego/sample/admin"
	"github.com/simbory/mego/sample/handlers"
	"github.com/simbory/mego/sample/filters"
	"github.com/simbory/mego/session/disk"
)

func main() {
	server := mego.NewServer(mego.WorkingDir(), ":8080", 0, "")

	cache.UseDefault()
	provider := disk.NewProvider(server.MapRootPath("/temp/sessions"))
	mgr := session.CreateManager(nil, provider)
	session.UseAsDefault(mgr)

	handlers.Init(server)
	filters.Init(server)
	admin.Init(server)

	server.Run()
}