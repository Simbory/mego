package filters

import "github.com/Simbory/mego"

func Init() {
	mego.HandleFilter("/*", globalFilter)
}
