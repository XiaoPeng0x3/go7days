package gee

import (
	"net/http"
	"strings"
)

// 定义通用的key
type H map[string]interface{}

// 思路是：在Engine里面添加一个存储路由的map, 在路由里面定义好方法

// 凡是函数参数是这个类型的函数都用 HandleFunc来接受
type HandlerFunc func(*Context)

// 路由分组功能
type (
	RouterGroup struct {
		prefix string // 前缀
		midddlewares []HandlerFunc // 中间件
		parent *RouterGroup // 分组
		engine *Engine // 共用一个engine
	}

	Engine struct {
		*RouterGroup // 组合实现继承
		router *router // 对应的路由
		groups []*RouterGroup // 记录这个Engine的所有子路由组， 创建路由时将新的路由添加进去

	}
)


// 相当于构造函数
func New() *Engine {

	engine := &Engine {router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// 创建新的分组
func (rg *RouterGroup) Group(prefix string) *RouterGroup {
	// 得到这个路由组的engine
	engine := rg.engine
	// 构建前缀和新的路由组
	prefixGroup := rg.prefix + prefix
	// 创建新的路由组
	newGroup := &RouterGroup{
		prefix: prefixGroup,
		parent: rg,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// 中间件函数
func (rg *RouterGroup) Use (handlers ...HandlerFunc) {
	rg.midddlewares = append(rg.midddlewares, handlers...) // 把这些中间件函数添加进对应的路由组
}


// 由于是分组路由
// 所以传递过来的其实是一个子路径， 在添加路由的时候要实现拼接
func (rg *RouterGroup) addRouter(method string, comp string, handler HandlerFunc) {
	pattern := rg.prefix + comp
	// 添加在路由树里面
	rg.engine.router.addRoute(method, pattern, handler)
}

// 实现GET
func (rg *RouterGroup) GET(pattern string, handler HandlerFunc) {
	rg.addRouter("GET", pattern, handler)
}

// 实现POST
func (rg *RouterGroup) POST(pattern string, handler HandlerFunc) {
	rg.addRouter("POST", pattern, handler)
}

// 实现Run
// 这个RUN方法独属于Engine
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// 实现ServeHTTP接口
// 实现这个接口后将会拦截所有的请求， 所以可以将请求逻辑全部放在这里来写
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 还需要实现中间件
	// 当用户的请求来到时， 要先做中间件处理
	midddlewares := make([]HandlerFunc, 0)
	// 先找到对应的路径
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) { // 用户请求的前缀和当前路由组前缀相等， 那么这个路由组的中间件就是应用在用户请求的中间件
			// 将这个路由组的所有中间件保存
			// 例如，假设我注册了一个路由 "/zxp", 并给这个路由指定了中间件方法
			// 那么，当用户的请求来到时， 例如 "/zxp/name", "/zxp/hello"
			// 这些用户的请求和我的路由组具有相同的前缀，那么用户的这些请求就应该全部我注册的 "/zxp"的中间件方法
			midddlewares = append(midddlewares, group.midddlewares...)
		}
	}


	c := NewContext(w, req)
	// 将这个路由组需要执行的中间件函数存在上下文中
	c.handlers = midddlewares 
	engine.router.handle(c)
}