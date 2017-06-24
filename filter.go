package mego

type filterKey struct {
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
	key := filterKey{
		pathPrefix: pathPrefix,
		matchAll: matchAll,
	}
	fc[key] = f
}
