// +build !windows

package view

import (
	"strings"
)

func fixPath(src string) string {
	return strings.Replace(src, "\\", "/", -1)
}

func defaultViewDir() string {
	return fixPath(workingDir() + "/views/")
}

func dirSlash(dir string) string {
	if strings.HasSuffix(dir, "/") {
		return dir
	}
	if strings.HasSuffix(dir, "\\") {
		return strings.TrimRight(dir, "\\") + "/"
	}
	return dir + "/"
}