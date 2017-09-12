# mego
mego route framework

### Sample
```
package main

import "github.com/simbory/mego"
import (
	"fmt"
	"github.com/simbory/mego/session/memory"
	"github.com/simbory/mego/session"
	"github.com/simbory/mego/cache"
	"io/ioutil"
	"time"
)

func pageFilter(ctx *mego.HttpCtx) {
	if ctx.Request().URL.Path == "/" {
		ctx.SetCtxItem("pageName", "home")
	}
}

func handleHome(ctx *mego.HttpCtx) interface{} {
	return ctx.TextResult(
		fmt.Sprintf("Hello, mego! this is the %s page", ctx.GetCtxItem("pageName")),
		"text/plain",
	)
}

func handleSession(ctx *mego.HttpCtx) interface{} {
	s := session.Default().Start(ctx)
	data := s.Get("data")
	if data == nil {
		data = "test session data"
		s.Set("data", data)
		return map[string]interface{} {
			"from_session": false,
			"data": data,
		}
	} else {
		return map[string]interface{} {
			"from_session": true,
			"data": data,
		}
	}
}

func handleCache(ctx *mego.HttpCtx) interface{} {
	data := cache.Default().Get("cache_data")
	if data == nil {
		data = "test cache data"
		dataFile := ctx.MapRootPath("/cache-dependency-file.txt")
		ioutil.WriteFile(dataFile, []byte(data.(string)), 0777)
		cache.Default().Set("cache_data", data, []string{dataFile}, 1 * time.Hour)
		return map[string]interface{} {
			"from_cache": false,
			"data": data,
		}
	} else {
		return map[string]interface{} {
			"from_cache": true,
			"data": data,
		}
	}
}

func main() {
	server := mego.NewServer(mego.WorkingDir(), ":8080")
	server.Route("/", handleHome)
	server.Route("/test-session", handleSession)
	server.Route("/test-cache", handleCache)
	server.HijackRequest("/", pageFilter)

	sessionProvider := memory.NewProvider()
	sessionManager := session.CreateManager(nil, sessionProvider)
	session.UseAsDefault(sessionManager)

	cache.UseDefault()

	server.Run()
}

```