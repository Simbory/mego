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

type routeProcessor struct {
	getFun  ReqHandler
	postFun ReqHandler
	putFun  ReqHandler
	optFun  ReqHandler
	delFun  ReqHandler
	traFun  ReqHandler
	hdlFun  ReqHandler
}

func (r *routeProcessor)Get(ctx *HttpCtx) interface{} {
	if r.getFun == nil {
		return nil
	}
	return r.getFun(ctx)
}

func (r *routeProcessor) Post(ctx *HttpCtx) interface{} {
	if r.postFun == nil {
		return nil
	}
	return r.postFun(ctx)
}

func (r *routeProcessor) Put(ctx *HttpCtx) interface{} {
	if r.putFun == nil {
		return nil
	}
	return r.putFun(ctx)
}

func (r *routeProcessor) Options(ctx *HttpCtx) interface{} {
	if r.optFun == nil {
		return nil
	}
	return r.optFun(ctx)
}

func (r *routeProcessor) Delete(ctx *HttpCtx) interface{} {
	if r.delFun == nil {
		return nil
	}
	return r.delFun(ctx)
}

func (r *routeProcessor) Trace(ctx *HttpCtx) interface{} {
	if r.traFun == nil {
		return nil
	}
	return r.traFun(ctx)
}

func (r *routeProcessor) ProcessRequest(ctx *HttpCtx) interface{} {
	if r.hdlFun == nil {
		return nil
	}
	return r.hdlFun(ctx)
}