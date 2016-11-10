package yun

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type (
	//Engine 框架引擎
	Engine struct {
		middlewares []HandlerFunc
		pool        sync.Pool
		router      router
		mode        Mode
	}

	//IGroup 路由组接口
	IGroup interface {
		Use(...HandlerFunc)
		Handle(string) IRoute
		Group(string, ...HandlerFunc) *Group
		Middlewares() []HandlerFunc
		Up() IGroup
	}
)

//New 新建框架引擎
//mode 运行模式，可选DEBUG,TEST,RELEASE
//return 返回框架引擎
func New(mode Mode) *Engine {
	eng := &Engine{}

	eng.mode = mode
	eng.pool.New = func() interface{} {
		return &Context{}
	}
	eng.router = router{minPrefix: 9999}

	eng.printDebugInfo(`[WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using code:	yun.New(yun.RELEASE) or yun.SetMode(yun.RELEASE)
`)

	return eng
}

func (eng *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := eng.pool.Get().(*Context)
	c.reset(w, req)

	//eng.handleHTTPRequest(c)
	hs := eng.findStaticRoute(req.URL.Path, req.Method)

	var params Params
	if hs == nil {
		hs, params = eng.findDynamicRoute(req.URL.Path, req.Method)
	}

	if hs != nil {
		c.setHandlers(hs)
		c.Params = params
		c.Next()
	} else if req.Method == "OPTIONS" {
		c.setHandlers(eng.middlewares)
		c.Next()
	}
	/*	if !c.Written() {
		p := req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			p = p + "?" + req.URL.RawQuery
		}
		eng.printDebugInfo("%s, %s, %s", req.Method, c.Status(), p)
	}*/
	if !c.Written() {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
	}

	eng.pool.Put(c)
}

//Run 运行http服务
//addr 服务地址
//return 返回错误
func (eng *Engine) Run(addr ...string) error {
	address := resolveAddress(addr)
	eng.printDebugInfo("Listening and serving HTTP on %s\n", address)
	err := http.ListenAndServe(address, eng)
	return err
}

//Handle 路由
//path 路由路径
//return 路由接口
func (eng *Engine) Handle(path string) IRoute {
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	if strings.Index(path, "//") >= 0 {
		panic("Wrong path format: '//' not allowed")
	}

	if path[0] != '/' {
		path = "/" + path
	}

	r := new(route)

	if strings.Index(path, ":") <= 0 && strings.Index(path, "*") <= 0 {
		r.ruType = sTATIC
	} else {
		r.ruType = dYNAMIC
	}

	r.path = path
	r.router = &eng.router
	r.group = eng

	return r
}

//Use 加入中间件
//middleware 中间件handle
func (eng *Engine) Use(middleware ...HandlerFunc) {
	eng.middlewares = append(eng.middlewares, middleware...)
}

//Group 创建路由组
//path 路由组路径
//middles 中间件
//return 路由组对象
func (eng *Engine) Group(path string, middles ...HandlerFunc) *Group {
	g := new(Group)
	g.path = path
	g.engine = eng
	g.parentGr = eng
	g.Use(middles...)

	return g
}

//SetMode 设置运行模式
//mode 运行模式，可选DEBUG,TEST,RELEASE
func (eng *Engine) SetMode(mode Mode) {
	eng.mode = mode
}

//Middlewares 获取中间件
//return 中间件handles
func (eng *Engine) Middlewares() []HandlerFunc {
	return eng.middlewares
}

//Up 获取上层路由组
//return 路由组接口
func (eng *Engine) Up() IGroup {
	return nil
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); len(port) > 0 {
			//debugPrint("Environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		//debugPrint("Environment variable PORT is undefined. Using port :8080 by default")
		return ":8000"
	case 1:
		{
			_, err := strconv.ParseUint(addr[0][1:], 10, 32)
			if err != nil {
				err = fmt.Errorf("Port number format error: %s", err)
				panic(err)
			}
			return addr[0]
		}
	default:
		panic("too much parameters")
	}
}

//findStaticRoute: 查找静态路由
func (eng *Engine) findStaticRoute(path, meth string) Handlers {
	key := staticRouteKey{
		method: meth,
		path:   path,
	}
	if hs, has := eng.router.staticRoutes[key]; has {
		return hs
	}

	return nil
}

//findDynamicRoute: 查找动态路由
func (eng *Engine) findDynamicRoute(path, method string) (Handlers, Params) {
	pathLen := uint16(len(path))
	levelNum := uint8(strings.Count(path, "/"))

	for i := eng.router.minPrefix; i <= eng.router.maxPrefix; i++ {
		if i > pathLen {
			break
		}

		key := dynamicRouteKey{prefix: path[:int(i)], levels: levelNum, method: method}
		rs, has := eng.router.dynamicRoutes[key]
		if !has {
			continue
		}

		for k := range rs {
			ppath := path[i+1:]
			node := rs[k].path
			params := make(Params, rs[k].paramNum)
			n, nextStart, match := 0, 0, true
		pathLoop:
			for {
				switch node.ntype {
				case fIXED:
					pplen := len(ppath)
					if pplen < node.length || pplen >= node.length && ppath[:node.length] != node.path {
						match = false
						break pathLoop
					}
					nextStart = node.length
				case pARAM, cATCHAll:
					end := 0
					for len := len(ppath); end < len && ppath[end] != '/'; end++ {
					}
					param := Param{Key: node.path, Value: ppath[:end]}
					params[n] = param
					n++
					nextStart = end
				}
				if node.next != nil {
					ppath = ppath[nextStart+1:]
					node = node.next
				} else {
					break
				}
			}

			if match {
				return rs[k].handlers, params
			}
		}
	}
	return nil, nil
}
