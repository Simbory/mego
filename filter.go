package mego

import "strings"

// filterKey the mego filter key
type filterKey struct {
	pathPrefix string
	matchAll   bool
}

// canExec check if the request with urlPath can be executable with current filter key.
func (fk *filterKey) canExec(urlPath string) bool {
	if fk.matchAll {
		return strings.HasPrefix(urlPath, fk.pathPrefix)
	} else {
		return urlPath == fk.pathPrefix
	}
}

// filterContainer the mego filter container
type filterContainer map[filterKey]func(*HttpCtx)

// exec execute the filter func
func (fc filterContainer) exec(urlPath string, ctx *HttpCtx) {
	for key, f := range fc {
		if key.canExec(urlPath) {
			if f(ctx); ctx.ended {
				break
			}
		}
	}
}

// add add a new filter to the filter container
func (fc filterContainer) add(pathPrefix string, matchAll bool, f func(*HttpCtx)) {
	key := filterKey{
		pathPrefix: pathPrefix,
		matchAll:   matchAll,
	}
	fc[key] = f
}
