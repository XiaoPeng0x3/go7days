package gee

import (
	"net/http"
)

// 定义通用的key
type H map[string]interface{}

// 思路是：在Engine里面添加一个存储路由的map, 在路由里面定义好方法

// 凡是函数参数是这个类型的函数都用 HandleFunc来接受
type HandlerFunc func(*Context)

// Engine实现ServeHTTP接口
type Engine struct {
	// 已经把router相关封装起来
	r *router
}

// 相当于构造函数
func New() *Engine {

	return &Engine{NewRouter()}
}
// 添加路由， 包括使用的方法， 注册的访问路径， 绑定实现的函数
func (e *Engine) addRouter(method string, pattern string, handler HandlerFunc) {
	e.r.addRoute(method, pattern, handler)
}

// 实现GET
func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRouter("GET", pattern, handler)
}

// 实现POST
func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRouter("POST", pattern, handler)
}

// 实现Run
func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

// 实现ServeHTTP接口
// 实现这个接口后将会拦截所有的请求， 所以可以将请求逻辑全部放在这里来写
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	e.r.handle(c)
}