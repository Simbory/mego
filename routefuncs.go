package mego

type RouteFilter interface {
	Filter(ctx *HttpCtx)
}

type RouteGet interface {
	Get(ctx *HttpCtx) interface{}
}

type RoutePost interface {
	Post(ctx *HttpCtx) interface{}
}

type RoutePut interface {
	Put(ctx *HttpCtx) interface{}
}

type RouteOptions interface {
	Options(ctx *HttpCtx) interface{}
}

type RouteDelete interface {
	Delete(ctx *HttpCtx) interface{}
}

type RouteTrace interface {
	Trace(ctx *HttpCtx) interface{}
}

type RoutePatch interface {
	Patch(ctx *HttpCtx) interface{}
}

type RouteHead interface {
	Head(ctx *HttpCtx) interface{}
}

type RouteConnect interface {
	Connect(ctx *HttpCtx) interface{}
}

type RouteProcessor interface {
	ProcessRequest(ctx *HttpCtx) interface{}
}