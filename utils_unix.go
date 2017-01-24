// +build !windows

package mego

import "strings"

func pathEq(path1, path2 string) bool {
	return path1 == path2
}

func pathHasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}