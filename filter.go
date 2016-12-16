package mego

import (
	"sync"
)

var indexCounter uint64
var filterLocker sync.RWMutex

type filterKey struct {
	index      uint64
	pathPrefix string
	matchAll   bool
}

func (fk *filterKey) canExec(urlPath string) bool {
	if fk.matchAll {
		return pathHasPrefix(urlPath, fk.pathPrefix)
	} else {
		return pathEq(fk.pathPrefix, urlPath)
	}
}

func newFilterKey(pathPrefix string, matchAll bool) filterKey {
	filterLocker.Lock()
	defer filterLocker.Unlock()
	s := filterKey{
		index:      indexCounter,
		pathPrefix: pathPrefix,
		matchAll: matchAll,
	}
	indexCounter++
	return s
}

type filterContainer map[filterKey]func(*Context)

func (fc filterContainer) exec(urlPath string, ctx *Context) {
	for key, f := range fc {
		if key.canExec(urlPath) {
			if f(ctx); ctx.ended {
				break
			}
		}
	}
}

func (fc filterContainer) add(pathPrefix string, matchAll bool, f func(*Context)) {
	var key = newFilterKey(pathPrefix, matchAll)
	fc[key] = f
}
