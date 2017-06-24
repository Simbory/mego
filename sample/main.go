package main

import (
	"time"
	"github.com/Simbory/mego"
	"github.com/Simbory/mego/cache"
	"github.com/Simbory/mego/session"
	"github.com/Simbory/mego/sample/testArea"
	"github.com/Simbory/mego/sample/handlers"
	"github.com/Simbory/mego/sample/filters"
)

func main() {
	mego.HandleDir("/static/")
	mego.HandleFile("/favicon.ico")
	cache.UseCache(10 * time.Second)
	session.UseSession(nil)

	handlers.Init()
	filters.Init()
	testArea.Init()

	mego.Run(":8080")
}
