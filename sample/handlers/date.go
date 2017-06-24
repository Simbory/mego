package handlers

import "github.com/Simbory/mego"

func getDate(ctx *mego.Context) interface{} {
	return &struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Day   int `json:"day"`
	}{
		Year:  int(ctx.RouteInt("year")),
		Month: int(ctx.RouteInt("month")),
		Day:   int(ctx.RouteInt("day")),
	}
}
