package handlers

import (
	"github.com/Simbory/mego"
	"github.com/Simbory/mego/session"
)

type userModel struct {
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
}

func testSession(ctx *mego.Context) interface{} {
	sessionStore := session.Default().Start(ctx)
	var msg string
	data := sessionStore.Get("msg")
	uData := sessionStore.Get("user")
	var user *userModel
	var ok bool
	if uData != nil {
		user, ok = uData.(*userModel)
	}
	if data != nil && ok {
		msg, _ = data.(string)
		return map[string]interface{}{
			"msg":         msg,
			"fromSession": true,
			"user":        user,
		}
	}
	msg = "Hello, world"
	user = &userModel{"Simbory", "Lu"}
	sessionStore.Set("msg", msg)
	sessionStore.Set("user", user)
	return map[string]interface{}{
		"msg":         msg,
		"user":        user,
		"fromSession": false,
	}
}