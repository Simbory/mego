// +build !windows

package view

import (
	"strings"
)

func fixPath(src string) string {
	return strings.Replace(src, "\\", "/", -1)
}

func viewDir() string {
	return fixPath(strAdd(RootDir, "/views/"))
}