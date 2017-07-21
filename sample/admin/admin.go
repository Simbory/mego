package admin

import (
	"fmt"
	"github.com/simbory/mego"
	"github.com/simbory/mego/assert"
	"github.com/simbory/mego/session"
	"github.com/simbory/mego/session/memory"
	"strings"
)

func getUpload(ctx *mego.HttpCtx) interface{} {
	return ctx.ViewResult("upload", nil)
}

func postUpload(ctx *mego.HttpCtx) interface{} {
	file := ctx.PostFile("file")
	filePath := ctx.MapContentPath(file.FileName)
	err := file.SaveAndClose(filePath)
	assert.PanicErr(err)
	return fmt.Sprintf("file Size: %d", file.Size)
}

func getLogin(ctx *mego.HttpCtx) interface{} {
	return ctx.ViewResult("login", nil)
}

func postLogin(ctx *mego.HttpCtx) interface{} {
	user := ctx.Request().FormValue("username")
	pwd := ctx.Request().FormValue("pwd")
	if user != "test" || pwd != "test" {
		return ctx.ViewResult("login", nil)
	}
	var returnUrl = ctx.Request().URL.Query().Get("returnUrl")
	if len(returnUrl) == 0 || !strings.HasPrefix(returnUrl, "/admin/") {
		returnUrl = "/admin/shell"
	}
	sessionManager.Start(ctx).Set("admin-user", "test")
	return ctx.RedirectResult(returnUrl, false)
}

var area *mego.Area
var sessionManager *session.Manager

func Init(server *mego.Server) {
	area = server.GetArea("admin")
	area.Get("/shell/upload", getUpload)
	area.Post("/shell/upload", postUpload)
	area.Get("/login", getLogin)
	area.Post("/login", postLogin)

	provider := memory.NewProvider()
	config := &session.Config{
		CookiePath: "/admin/",
		CookieName: "ADMIN_SESSION_ID",
	}
	sessionManager = session.CreateManager(config, provider)

	area.HandleFilter("/shell/*", func(ctx *mego.HttpCtx) {
		s := sessionManager.Start(ctx)
		userData := s.Get("admin-user")
		if userData == nil {
			ctx.Redirect("/admin/login?returnUrl="+ctx.Request().URL.Path, false)
		}
	})
}
