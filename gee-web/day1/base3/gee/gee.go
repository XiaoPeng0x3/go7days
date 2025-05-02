package gee

import (
	"log"
	"net/http"
)

// 思路是：在Engine里面添加一个存储路由的map, 在路由里面定义好方法

// 凡是函数参数是这个类型的函数都用 HandleFunc来接受
type HandlerFunc func(http.ResponseWriter, *http.Request)

// Engine实现ServeHTTP接口
type Engine struct {
	// key值为string, val值为HandleFunc, 这样就可以进行绑定
	router map[string] HandlerFunc
}

// 相当于构造函数
func New() *Engine {

	return &Engine{make(map[string]HandlerFunc)}
}
// 添加路由， 包括使用的方法， 注册的访问路径， 绑定实现的函数
func (e *Engine) addRouter(method string, pattern string, handler HandlerFunc) {
	// key值为 method-pattern, 例如 "GET-/", 以GET的方式获取/
	// val值为绑定的函数
	key := method + "-" + pattern
	e.router[key] = handler
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
	// 从req里面获取请求
	key := req.Method + "-" + req.URL.Path
	if handler, ok := e.router[key]; ok { // 获取存储在map里面的handler
		handler(w, req)
	} else {
		log.Fatal("404 Not Found")
	}
}