package mego

import "strings"

func pathEq(path1, path2 string) bool {
	return strings.ToLower(path1) == strings.ToLower(path2)
}

func pathHasPrefix(s, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
}