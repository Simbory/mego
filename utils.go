package mego

import (
	"strings"
	"unsafe"
	"path/filepath"
	"os"
	"log"
)

const Version = "1.0"

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

func ExeDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func WorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ExeDir()
	}
	return strings.Replace(dir, "\\", "/", -1)
}