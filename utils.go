package mego

import (
	"strings"
	"unsafe"
	"path/filepath"
	"os"
	"log"
	"path"
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

// ClearPath clear the pathStr and return the shortest path.
func ClearPath(pathStr string) string {
	return path.Clean(strings.Replace(pathStr, "\\", "/", -1))
}

// EnsurePrefix ensure that the string s must has the prefix
func EnsurePrefix(s, prefix string) string {
	if !strings.HasPrefix(s, prefix) {
		return strAdd(prefix, s)
	}
	return s
}

// EnsureSuffix ensure that the string s must has suffix
func EnsureSuffix(s, suffix string) string {
	if !strings.HasSuffix(s, suffix) {
		return strAdd(s, suffix)
	}
	return s
}

// ExeDir get the current directory that the executable file is located in.
func ExeDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

// WorkingDir get the current working directory
func WorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ExeDir()
	}
	return strings.Replace(dir, "\\", "/", -1)
}