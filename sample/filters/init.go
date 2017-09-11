package filters

import "github.com/simbory/mego"

func Init(server *mego.Server) {
	server.HijackRequest("/", globalFilter)
}
