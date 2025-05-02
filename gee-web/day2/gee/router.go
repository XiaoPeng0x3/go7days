package gee

import (
	"net/http"
)

type router struct {
	// 存放的是kv map
	handlers map[string]HandlerFunc
}

// 构造函数
func NewRouter() *router {
	return &router{make(map[string]HandlerFunc)}
}

// 添加路由
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	r.handlers[key] = handler
}

// 
func (r *router) handle(c *Context) {
	// 构造key
	key := c.Method + "-" + c.Path
	if handler, ok := r.handlers[key]; ok{
		handler(c)
	} else {
		// 找不到对应路径
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}

