package filters

import "github.com/Simbory/mego"

func globalFilter(ctx *mego.Context) {
	ctx.SetItem("user", "Simbory")
}
