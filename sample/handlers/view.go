package handlers

import (
	"github.com/Simbory/mego"
	"github.com/Simbory/mego/cache"
	"time"
)

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
	return view.Render(ctx.RouteString("view"), viewData)
}
