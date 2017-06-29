# mego
mego route framework

### Sample
```
package main

import (
	"os"
	"time"

	"github.com/simbory/mego"
	"github.com/simbory/mego/cache"
	"github.com/simbory/mego/session"
	"github.com/simbory/mego/view"
)

func getDate(ctx *mego.Context) interface{} {
	return &struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Day   int `json:"day"`
	}{
		Year:  int(ctx.RouteParamInt("year")),
		Month: int(ctx.RouteParamInt("month")),
		Day:   int(ctx.RouteParamInt("day")),
	}
}

func renderView(ctx *mego.Context) interface{} {
	msg := cache.Cache().Get("msg")
	if msg == nil {
		msg = time.Now().Format(time.RFC1123Z)
		expire := time.Now().Add(5 * time.Second)
		cache.Cache().Add("msg", msg, nil, &expire)
	}
	viewData := map[string]interface{}{
		"msg": msg,
	}
	return view.View(ctx.RouteParamString("view"), viewData)
}

func testSession(ctx *mego.Context) interface{} {
	sessionStore := session.Start(ctx)
	var msg string
	data := sessionStore.Get("msg")
	if data != nil {
		msg, _ = data.(string)
		return &map[string]interface{}{
			"msg":         msg,
			"fromSession": true,
		}
	}
	msg = "Hello, world"
	sessionStore.Set("msg", msg)
	return &map[string]interface{}{
		"msg":         msg,
		"fromSession": false,
	}
}

func testFilter(ctx *mego.Context) interface{} {
	data := ctx.GetItem("user")
	if data != nil {
		return data.(string)
	}
	return nil
}

func workingDir() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}

func globalFilter(ctx *mego.Context) {
	ctx.SetItem("user", "Simbory")
}

func main() {
	cache.UseCache(10 * time.Second)
	session.UseSession(nil)
	view.UseView(workingDir() + "/views/")
	mego.HandleStaticDir("/static/", workingDir() + "/static/")
	mego.HandleStaticFile("/favicon.ico", workingDir() + "/favicon.ico")
	mego.Any("/views/<view:word>", renderView)
	mego.Get("/date/<year:int>-<month:int>-<day:int>", getDate)
	mego.Get("/session", testSession)
	mego.Any("/filter/*pathInfo", testFilter)
	mego.Filter("/*", globalFilter)
	mego.Run(":8080")
}
```