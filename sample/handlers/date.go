package handlers

import "github.com/simbory/mego"

func getDate(ctx *mego.HttpCtx) interface{} {
	return &struct {
		Year  string `json:"year"`
		Month string `json:"month"`
		Day   string `json:"day"`
	}{
		Year:  ctx.RouteVar("year"),
		Month: ctx.RouteVar("month"),
		Day:   ctx.RouteVar("day"),
	}
}
