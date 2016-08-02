package yun

import (
	"fmt"
)

/*const (
	Colon
	Asterisk
)*/

type (
	RouteType byte

	IRoute interface {
		Connect(...HandlerFunc) IRoute
		Delete(...HandlerFunc) IRoute
		Get(...HandlerFunc) IRoute
		Head(...HandlerFunc) IRoute
		Options(...HandlerFunc) IRoute
		Patch(...HandlerFunc) IRoute
		Post(...HandlerFunc) IRoute
		Put(...HandlerFunc) IRoute
		Trace(...HandlerFunc) IRoute
		Any(...HandlerFunc)
	}

	route struct {
		path   string
		ruType routeType
		engine *Engine
		group  *Group
	}

	staticRouteKey struct {
		method string
		path   string
	}

	dynamicRouteKey struct {
		prefix string
		method string
		levels uint8
	}

	dynamicRoute struct {
		path     *node
		paramNum uint8
		handlers Handlers
	}

	router struct {
		staticRoutes  map[staticRouteKey]Handlers
		dynamicRoutes map[dynamicRouteKey][]*dynamicRoute
	}
)

type paramType uint8

// HTTP methods
const (
	CONNECT = "CONNECT"
	DELETE  = "DELETE"
	GET     = "GET"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
	PATCH   = "PATCH"
	POST    = "POST"
	PUT     = "PUT"
	TRACE   = "TRACE"
)

type routeType int

const (
	STATIC routeType = iota
	DYNAMIC
)

type nodeType int

const (
	FIXED nodeType = iota
	PARAM
	CATCHAll
)

var (
	methods = [...]string{
		CONNECT,
		DELETE,
		GET,
		HEAD,
		OPTIONS,
		PATCH,
		POST,
		PUT,
		TRACE,
	}
)

func (r *route) Get(handlers ...HandlerFunc) IRoute {
	r.handle(GET, handlers)
	return r
}

func (r *route) Post(handlers ...HandlerFunc) IRoute {
	r.handle(POST, handlers)
	return r
}

func (r *route) Options(handlers ...HandlerFunc) IRoute {
	r.handle(OPTIONS, handlers)
	return r
}

func (r *route) Delete(handlers ...HandlerFunc) IRoute {
	r.handle(DELETE, handlers)
	return r
}

func (r *route) Head(handlers ...HandlerFunc) IRoute {
	r.handle(HEAD, handlers)
	return r
}

func (r *route) Patch(handlers ...HandlerFunc) IRoute {
	r.handle(PATCH, handlers)
	return r
}

func (r *route) Put(handlers ...HandlerFunc) IRoute {
	r.handle(PUT, handlers)
	return r
}

func (r *route) Trace(handlers ...HandlerFunc) IRoute {
	r.handle(TRACE, handlers)
	return r
}

func (r *route) Connect(handlers ...HandlerFunc) IRoute {
	r.handle(CONNECT, handlers)
	return r
}

func (r *route) Any(handlers ...HandlerFunc) {
	r.handle(GET, handlers)
	r.handle(POST, handlers)
	r.handle(PUT, handlers)
	r.handle(PATCH, handlers)
	r.handle(HEAD, handlers)
	r.handle(OPTIONS, handlers)
	r.handle(DELETE, handlers)
	r.handle(TRACE, handlers)
	r.handle(CONNECT, handlers)
}

//handle: 处理路由
func (r *route) handle(meth string, handlers Handlers) {
	handlers = r.mergeHandlers(handlers)

	if r.ruType == STATIC {
		if r.engine.staticRoutes == nil {
			r.engine.staticRoutes = make(map[staticRouteKey]Handlers)
		}

		key := staticRouteKey{
			path:   r.path,
			method: meth,
		}

		//路径冲突
		if _, has := r.engine.staticRoutes[key]; has {
			panic("This route already exists")
		}

		r.engine.staticRoutes[key] = handlers
	} else {
		r.dynamicHandle(meth, handlers)
	}
}

//dynamicHandle: 处理动态路由
func (r *route) dynamicHandle(meth string, handlers Handlers) {
	if r.engine.dynamicRoutes == nil {
		r.engine.dynamicRoutes = make(map[dynamicRouteKey][]*dynamicRoute)
	}

	//创建动态路由
	ds := new(dynamicRoute)
	ds.handlers = handlers

	var (
		nodeStart, nodeEnd int
		levels             uint8 //路径级数
		nod                *node
		prefix             string
		nodType            nodeType
	)

	for i, _ := range r.path {
		switch r.path[i] {
		case '/':
			levels++
			nodeEnd = i
			if nodType == PARAM || nodType == CATCHAll {
				nod = addNode(nod, r.path[nodeStart+2:nodeEnd], nodType)
				nodType = FIXED
				nodeStart = i
			}
		case ':', '*':
			if i-nodeEnd > 1 {
				panic(fmt.Sprintf("Path format error, '%c' must be next to '/'", r.path[i]))
			}

			ds.paramNum++

			if nodeStart <= 0 {
				prefix = r.path[:nodeEnd]
			} else if nodeStart < nodeEnd {
				nod = addNode(nod, r.path[nodeStart+1:nodeEnd], nodType)
			}

			if r.path[i] == ':' {
				nodType = PARAM
			} else {
				nodType = CATCHAll
			}

			nodeStart = nodeEnd
		}

		if i >= len(r.path)-1 {
			l := 1
			if nodType == PARAM || nodType == CATCHAll {
				l = 2
			}
			nod = addNode(nod, r.path[nodeStart+l:], nodType)
		}

		if ds.path == nil && nod != nil {
			ds.path = nod
		}
	}

	key := dynamicRouteKey{
		prefix: prefix,
		method: meth,
		levels: levels,
	}
	if rs, has := r.engine.dynamicRoutes[key]; has {
		rs = append(rs, ds)
		r.engine.dynamicRoutes[key] = rs
	} else {
		r.engine.dynamicRoutes[key] = make([]*dynamicRoute, 1)
		r.engine.dynamicRoutes[key][0] = ds
	}

	preLen := uint16(len(prefix))
	if r.engine.maxPrefix < preLen {
		r.engine.maxPrefix = preLen
	}
	if r.engine.minPrefix > preLen {
		r.engine.minPrefix = preLen
	}
}

func (r *route) mergeHandlers(handlers Handlers) Handlers {
	rootMiddleSize := len(r.engine.middleware)
	groupMiddleSize := 0
	if r.group != nil {
		groupMiddleSize = len(r.group.middleware)
	}
	size := rootMiddleSize + groupMiddleSize + len(handlers)
	hs := make(Handlers, size)

	copy(hs, r.engine.middleware)
	if r.group != nil {
		copy(hs[rootMiddleSize:], r.group.middleware)
	}
	copy(hs[rootMiddleSize+groupMiddleSize:], handlers)

	return hs
}
