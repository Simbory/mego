package handlers

import (
	"github.com/Simbory/mego"
	"github.com/Simbory/mego/session"
)

func testSession(ctx *mego.Context) interface{} {
	sessionStore := session.Default().Start(ctx)
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