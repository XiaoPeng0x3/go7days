package gee

import (
	"fmt"
	"net/http"
	"strings"
)

// 改写router, 封装寻找路由的功能
type router struct {
	// 存放的是kv map
	handlers map[string] HandlerFunc
	// 存放每种请求方式的根节点
	root map[string] *node
}

// 构造函数
func newRouter() *router {
	return &router{
		handlers: make(map[string]HandlerFunc),
		root: make(map[string] *node),
	}
}

// 解析parts数组
func parsePattern(pattern string) []string {
	parts := make([]string, 0)
	// 将pattern字符串按照 / 进行切分
	vs := strings.Split(pattern, "/")
	for _, part := range vs {
		if part != "" { // 添加到路由里面
			parts = append(parts, part)
			if part[0] == '*' { // 后面的路径需要全部匹配，所以不需要append
				break
			}
		}
	}
	return parts
}

// 添加路由
// 添加路由的时候还需要构建前缀树
// pattern是我们自定义的路由匹配规则
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	// 构建前缀树
	parts := parsePattern(pattern)
	// 先判断method对应的根节点是否存在
	_, ok := r.root[method]
	if !ok { // 不存在说明是第一次建立，先创建好根节点
		r.root[method] = &node{}
	}
	// 将parts构建为前缀树
	r.root[method].insert(pattern, parts, 0)
	// 下面的逻辑不用变 
	key := method + "-" + pattern
	r.handlers[key] = handler
}

// 查询路由
// path是用户传入的真实URL
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	// 待搜索的路径
	searchParts := parsePattern(path)
	params := make(map[string]string)

	// 根据method来搜索对应方法的路由树的根节点
	root, ok := r.root[method]
	if !ok { // 说明不存在， 用户请求有误
		return nil, nil
	}
	// 开始进行搜索
	leaf := root.search(searchParts, 0) // 如果不存在这个路径那么就会返回nil
	if leaf != nil { // 可以找到
		parts := parsePattern(leaf.pattern) // 解析定义的路由树
		// 这里要进行赋值操作
		for i, part := range parts {
			if part[0] == ':' { // 模糊匹配， 赋值
				params[part[1:]] = searchParts[i]
			}
			if part[0] == '*' && len(part) > 1 {
				// 将用户路径全部拼接起来
				params[part[1:]] = strings.Join(searchParts[i:], "/")
			}
		}
		return leaf, params
	}
	return nil, nil
}



// 使用中间件后，由于中间件函数全部存储在c.handlers列表里面
// 为了更好的进行执行，我们就使用c.Next()函数来遍历执行列表里面的函数
// 此时，应该将业务逻辑函数append添加在c.handlers里面
func (r *router) handle(c *Context) {
	// 查询params
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil{
		c.Params = params
		// 构造key
		// 不去直接使用r.handlers判断是否存在是因为这次存在动态路径，所以要使用前缀树的搜索方法去匹配
		key := c.Method + "-" + n.pattern
		fmt.Println(c.Path, n.pattern)
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		// 找不到对应路径
		// 也需要将找不到路径的函数添加进来
		c.handlers = append(c.handlers, func(ctx *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	// 顺序执行
	c.Next()
}

