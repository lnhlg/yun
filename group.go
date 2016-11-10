package yun

import (
	"path"
)

type (
	//Group 路由数组
	Group struct {
		path        string
		middlewares []HandlerFunc
		engine      *Engine
		parentGr    IGroup
		router      router
	}
)

//Use 注册中间件
//middles 中间件handle
func (g *Group) Use(middles ...HandlerFunc) {
	g.middlewares = append(g.middlewares, middles...)
}

//Handle 组路由路径
//rpath 路径字符串
//return 路由接口
func (g *Group) Handle(rpath string) IRoute {
	fullPath := mergePath(g.path, rpath)
	r := g.engine.Handle(fullPath).(*route)
	r.group = g

	return r
}

//Group 创建子路由组
//path 组路径
//middles 组中件间
//return 返回路由组
func (g *Group) Group(path string, middles ...HandlerFunc) *Group {
	gr := new(Group)
	gr.path = mergePath(g.path, path)
	gr.engine = g.engine
	gr.parentGr = g
	gr.Use(middles...)

	return gr
}

//Middlewares 获取组路由中间件
//return 中间件执行体
func (g *Group) Middlewares() []HandlerFunc {
	return g.middlewares
}

//Up 获取上级路由组
//return 路由组接口
func (g *Group) Up() IGroup {
	return g.parentGr
}

func mergePath(basePath, joinpath string) string {
	var fullPath string
	if basePath[len(basePath)-1] == '/' {
		fullPath = basePath[:len(basePath)-1]
	} else {
		fullPath = basePath
	}

	if joinpath[0] == '/' {
		fullPath = path.Join(fullPath, joinpath[1:])
	} else {
		fullPath = path.Join(fullPath, joinpath)
	}

	return fullPath
}
