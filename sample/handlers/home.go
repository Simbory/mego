package handlers

import "github.com/Simbory/mego"

func home(_ *mego.Context) interface{} {
	return view.Render("home", nil)
}