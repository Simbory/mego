package mego

// RoutePropFind WebDAV http method 'PROPFIND' processing interface
type RoutePropFind interface {
	PropFind(ctx *HttpCtx) interface{}
}

// RoutePropPatch WebDAV http method 'PROPPATCH' processing interface
type RoutePropPatch interface {
	PropPatch(ctx *HttpCtx) interface{}
}

// RouteMkcol WebDAV http method 'MKCOL' processing interface
type RouteMkcol interface {
	Mkcol(ctx *HttpCtx) interface{}
}

// RouteCopy WebDAV http method 'COPY' processing interface
type RouteCopy interface {
	Copy(ctx *HttpCtx) interface{}
}

// RouteMove WebDAV http method 'MOVE' processing interface
type RouteMove interface {
	Move(ctx *HttpCtx) interface{}
}

// RouteLock WebDAV http method 'LOCK' processing interface
type RouteLock interface {
	Lock(ctx *HttpCtx) interface{}
}

// RouteUnlock WebDAV http method 'UNLOCK' processing interface
type RouteUnlock interface {
	Unlock(ctx *HttpCtx) interface{}
}
