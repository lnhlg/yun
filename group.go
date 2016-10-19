package yun

import (
	"path"
)

type (
	Group struct {
		path 		string
		middlewares	[]HandlerFunc
		engine		*Engine
		parentGr	IGroup
		router		router
	}
)

func (g *Group) Use(middles ...HandlerFunc) {
	g.middlewares= append(g.middlewares, middles...)
}

func (g *Group) Handle(rpath string) IRoute {
	fullPath := mergePath(g.path, rpath)
	r := g.engine.Handle(fullPath).(*route)
	r.group = g

	return r
}

func (g *Group) Group(path string, middles ...HandlerFunc) *Group {
	gr := new(Group)
	gr.path = mergePath(g.path, path)
	gr.engine = g.engine
	gr.parentGr = g
	gr.Use(middles...)

	return gr
}

func (g *Group) Middlewares() []HandlerFunc {
	return g.middlewares
}

func (g *Group) Up() IGroup {
	return g.parentGr
}

func (g *Group)Router() *router {
	return &g.router
}

func mergePath(basePath, joinpath string) string {
	var fullPath string
	if basePath[len(basePath) - 1] == '/' {
		fullPath = basePath[ : len(basePath) - 1]
	} else {
		fullPath = basePath
	}

	if joinpath[0] == '/' {
		fullPath = path.Join(fullPath, joinpath[1 : ])
	} else {
		fullPath = path.Join(fullPath, joinpath)
	}

	return fullPath
}