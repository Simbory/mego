package filters

import "github.com/Simbory/mego"

func Init(server *mego.Server) {
	server.HandleFilter("/*", globalFilter)
}
