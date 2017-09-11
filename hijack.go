package mego

import "strings"

// hijackKey the key to match the hijack rule
type hijackKey string

// match check if the request with urlPath can be executable with current filter key.
func (fk hijackKey) match(urlPath string) bool {
	return string(fk) == urlPath || strings.HasPrefix(urlPath, string(fk)+"/")
}

// hijackContainer the mego hijack container
type hijackContainer map[hijackKey]func(*HttpCtx)

// exec hijack the request
func (fc hijackContainer) exec(urlPath string, ctx *HttpCtx) {
	for key, f := range fc {
		if !key.match(urlPath) {
			continue
		}
		if f(ctx); ctx.ended {
			break
		}
	}
}

// add add a new hijack rule
func (fc hijackContainer) add(pathPrefix string, f func(*HttpCtx)) {
	pathPrefix = EnsurePrefix(pathPrefix, "/")
	pathPrefix = strings.TrimRight(pathPrefix, "/")
	fc[hijackKey(pathPrefix)] = f
}