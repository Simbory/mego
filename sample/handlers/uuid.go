package handlers

import (
	"github.com/google/uuid"
	"github.com/simbory/mego"
)

func testUUID(ctx *mego.HttpCtx) interface{} {
	data := make(map[int]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		id := uuid.New()
		t := id.Variant().String()
		data[i] = &struct {
			Str  string `json:"str"`
			Type string `json:"type"`
		}{id.String(), t}
	}
	return ctx.JsonResult(data)
}
