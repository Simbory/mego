package view

import (
	"strings"
	"github.com/Simbory/mego"
)

func fixPath(src string) string {
	return strings.Replace(src, "/", "\\", -1)
}

func viewDir() string {
	return fixPath(mego.RootDir + "\\views\\")
}