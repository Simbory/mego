package mego

import (
	"strings"
	"sync"
)

var indexCounter uint64
var filterkeyLocker sync.RWMutex

type filterKey struct {
	index      uint64
	pathPrefix string
}

func newFilterKey(pathPrefix string) filterKey {
	filterkeyLocker.Lock()
	defer filterkeyLocker.Unlock()
	s := filterKey{
		index:      indexCounter,
		pathPrefix: pathPrefix,
	}
	indexCounter++
	return s
}

type filterContainer map[filterKey]func(*Context)

func (fc filterContainer) exec(urlPath string, ctx *Context) {
	for key, f := range fc {
		if strings.HasPrefix(urlPath, key.pathPrefix) {
			if f(ctx); ctx.ended {
				break
			}
		}
	}
}

func (fc filterContainer) add(pathPrefix string, f func(*Context)) {
	var key = newFilterKey(pathPrefix)
	fc[key] = f
}
