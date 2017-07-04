package mego

import "strings"

type filterKey struct {
	pathPrefix string
	matchAll   bool
}

func (fk *filterKey) canExec(urlPath string) bool {
	if fk.matchAll {
		return strings.HasPrefix(urlPath, fk.pathPrefix)
	} else {
		return urlPath == fk.pathPrefix
	}
}

type filterContainer map[filterKey]func(*HttpCtx)

func (fc filterContainer) exec(urlPath string, ctx *HttpCtx) {
	for key, f := range fc {
		if key.canExec(urlPath) {
			if f(ctx); ctx.ended {
				break
			}
		}
	}
}

func (fc filterContainer) add(pathPrefix string, matchAll bool, f func(*HttpCtx)) {
	key := filterKey{
		pathPrefix: pathPrefix,
		matchAll: matchAll,
	}
	fc[key] = f
}
