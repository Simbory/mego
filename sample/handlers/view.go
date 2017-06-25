package handlers

import (
	"github.com/Simbory/mego"
	"github.com/Simbory/mego/cache"
	"time"
)

func renderView(ctx *mego.Context) interface{} {
	msg := cache.Default().Get("msg")
	if msg == nil {
		msg = time.Now().Format(time.RFC1123Z)
		expired := time.Second * 5
		cache.Default().Set("msg", msg, nil, expired)
	}
	viewData := map[string]interface{}{
		"msg": msg,
	}
	return view.Render(ctx.RouteString("view"), viewData)
}