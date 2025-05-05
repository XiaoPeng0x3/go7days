package gee

import (
	"html/template"
	"net/http"
	"path"
	"strings"
)

// 定义通用的key
type H map[string]interface{}

// 思路是：在Engine里面添加一个存储路由的map, 在路由里面定义好方法

// 凡是函数参数是这个类型的函数都用 HandleFunc来接受
type HandlerFunc func(*Context)

// 路由分组功能
// day6 添加HTML渲染
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
		htmlTemplates *template.Template
		funcMap template.FuncMap
	}
)


// 相当于构造函数
func New() *Engine {

	engine := &Engine {router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (e *Engine) setFuncMap(funcMap *template.FuncMap) {
	e.funcMap = *funcMap
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
	c.engine = engine
	engine.router.handle(c)
}

/**

梳理一下整个文件映射的思路，首先就是去绑定映射关系，使用Static函数即可，例如，绑定
r.Static("/assets", "/usr/geektutu/blog/static")
就是把"/assets"绑定到 "/usr/geektutu/blog/static"目录下
1. 注册路由。
我们可以假定用户访问 “assets/xxx”就是在访问一些静态资源，所以，在注册路由的时候就可以用动态路由来实现匹配
所以，r.Static("/assets", "/usr/geektutu/blog/static") 在路由树上注册的路由实际上是 assets/*filepath

2. 绑定
绑定就是去将路由树上的路由添加方法即可。所以，为了返回服务器上的资源，我们只需要创建对应的方法即可。当请求来临时，假设访问url是 assets/css/main.html
而这个assets路径在服务器上的root路径其实就是"/usr/geektutu/blog/static"这个路径，"assets"是不存在的，但 “css/main.html"却是真实存在的
所以我们可以将访问url的前缀去除，得到真实的文件路径 css/main

3. 实现方法
将对应方法进行绑定后，怎么打开文件呢？在查询路径的时候，查询到/assets/*filepath这个路径，这里就会把c.Params["filepath"]给赋值为url里面的路径， 例如css/main.html，然后再去尝试打开这个文件
如果可以正常打开，说明存在这个资源；否则说明不存在这个资源

**/

// day6 模板渲染
// 将静态资源映射返回
// fs是服务器资源的根路径
func (rg *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	// 获取绝对路径
	absolutePath := path.Join(rg.prefix, relativePath)
	// 去掉前缀，并且将去掉前缀后的文件地址给http.FileServer
	// 这个http.FileServer会去创建一个
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return func(c *Context) {
		file := c.Params["filepath"]
		// open, 如果不存在就会open失败
		if _, err := fs.Open(file); err != nil { // open失败
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

func (rg *RouterGroup) Static(relativePath string, root string) {
	// 业务逻辑
	handler := rg.createStaticHandler(relativePath, http.Dir(root))
	// urlPath = /assets/*filepath
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	// 添加路由
	rg.GET(urlPattern, handler)
}