package main

import (
	"time"
	"github.com/Simbory/mego"
	"github.com/Simbory/mego/cache"
	"github.com/Simbory/mego/session"
	"github.com/Simbory/mego/view"
	"github.com/google/uuid"
	testArea "github.com/Simbory/mego/sample/test"
)

func getDate(ctx *mego.Context) interface{} {
	return &struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Day   int `json:"day"`
	}{
		Year:  int(ctx.RouteInt("year")),
		Month: int(ctx.RouteInt("month")),
		Day:   int(ctx.RouteInt("day")),
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
	return viewEngine.RenderView(ctx.RouteString("view"), viewData)
}

func testSession(ctx *mego.Context) interface{} {
	sessionStore := session.Start(ctx)
	var msg string
	data := sessionStore.Get("msg")
	if data != nil {
		msg, _ = data.(string)
		return map[string]interface{}{
			"msg":         msg,
			"fromSession": true,
		}
	}
	msg = "Hello, world"
	sessionStore.Set("msg", msg)
	return map[string]interface{}{
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

func testUUID(_ *mego.Context) interface{} {
	data := make(map[int]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		id := uuid.New()
		t := id.Variant().String()
		data[i] = &struct {
			Str string `json:"str"`
			Type string `json:"type"`
		}{id.String(), t}
	}
	return data
}

func globalFilter(ctx *mego.Context) {
	ctx.SetItem("user", "Simbory")
}

var viewEngine *view.ViewEngine

func main() {
	mego.HandleDir("/static/")
	mego.HandleFile("/favicon.ico")
	mego.Any("/views/<view:word>", renderView)
	mego.Get("/date/<year:int>-<month:int>-<day:int>", getDate)
	mego.Get("/session", testSession)
	mego.Get("/uuid", testUUID)
	mego.Any("/filter/*pathInfo", testFilter)
	mego.HandleFilter("/*", globalFilter)

	cache.UseCache(10 * time.Second)
	session.UseSession(nil)
	viewEngine = view.NewViewEngine("views", ".html")

	testArea.InitArea()

	mego.Run(":8080")
}
