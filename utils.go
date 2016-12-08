package mego

import (
	"os"
	"strings"
	"unsafe"
)

func byte2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func str2Byte(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func strAdd(arr ...string) string {
	if len(arr) == 0 {
		return ""
	}
	return strings.Join(arr, "")
}

func isFile(path string) bool {
	state, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !state.IsDir()
}

func isDir(path string) bool {
	state, err := os.Stat(path)
	if err != nil {
		return false
	}
	return state.IsDir()
}

func workingDir() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}
