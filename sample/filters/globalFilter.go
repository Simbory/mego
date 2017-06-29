package filters

import "github.com/simbory/mego"

func globalFilter(ctx *mego.Context) {
	ctx.SetItem("user", "Simbory")
}
