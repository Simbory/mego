package mego

type RouteGetter interface {
	Get(ctx *HttpCtx) interface{}
}

type RoutePoster interface {
	Post(ctx *HttpCtx) interface{}
}

type RoutePutter interface {
	Put(ctx *HttpCtx) interface{}
}

type RouteOptioner interface {
	Options(ctx *HttpCtx) interface{}
}

type RouteDeleter interface {
	Delete(ctx *HttpCtx) interface{}
}

type RouteTracer interface {
	Trace(ctx *HttpCtx) interface{}
}

type RouteProcessor interface {
	ProcessRequest(ctx *HttpCtx) interface{}
}