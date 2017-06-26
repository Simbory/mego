package handlers

import "github.com/Simbory/mego"

func home(_ *mego.Context) interface{} {
	return mego.View("home", nil)
}