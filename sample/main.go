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
	mego.HandleDir("/static/")
	mego.HandleFile("/favicon.ico")

	cache.UseDefault()
	session.UseDefault()

	handlers.Init()
	filters.Init()
	admin.Init()

	mego.Run(":8080")
}
