package filters

import "github.com/simbory/mego"

func globalFilter(ctx *mego.HttpCtx) {
	ctx.SetCtxItem("user", "Simbory")
}
