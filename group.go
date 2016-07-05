package yun

import "path"

type (
	Group struct {
		path 		string
		middleware	[]HandlerFunc
		engine 		*Engine
	}
)

func (g *Group) Use(middles ...HandlerFunc) {
	g.middleware = append(g.middleware, middles...)
}

func (g *Group) Handle(rpath string) IRoute {
	var basePath string
	if g.path[len(g.path) - 1] == '/' {
		basePath = g.path[ : len(g.path) - 1]
	} else {
		basePath = g.path
	}

	var fullPath string
	if rpath[0] == '/' {
		fullPath = path.Join(basePath, rpath[1 : ])
	} else {
		fullPath = path.Join(basePath, rpath)
	}

	r := g.engine.Handle(fullPath).(*route)
	r.group = g

	return r
}


