// 封装Response和Request接口
package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// 封装context结构体
type Context struct {
	// http.ResponseWriter
	Writer http.ResponseWriter
	// http.Request
	Req *http.Request
	// Path
	Path string
	// Method
	Method string
	// 状态码
	StatusCode int
	// 动态路由获取的参数
	Params map[string]string
}

// 构造函数
func NewContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req: req,
		Path: req.URL.Path,
		Method: req.Method,
		Params: make(map[string]string),
	}
}

// 根据param获取对应参数
func (c *Context) Param(key string) string {
	val, _ := c.Params[key]
	return val
}


// 获取post表单数据
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// 获取get请求参数
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// 设置响应码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// 设置响应头
func (c *Context) SetHeader(key string, val string) {
	c.Writer.Header().Set(key, val)
}

// 设置纯文本格式 Content-Type
func (c *Context) String(code int, format string, vals ...interface{}) {
	c.SetHeader("Content-Type", "text/plain") // 纯文本格式
	c.Status(code)
	// 把数据写回去
	c.Writer.Write([]byte(fmt.Sprintf(format, vals...))) // Sprintf函数有返回值
}

// 响应JSON数据
func (c *Context) JSON(code int, obj interface{}) {
	// 设置响应头
	c.SetHeader("Content-Type", "application/json") // 以json格式展示obj
	// 得到encoder
	encoder := json.NewEncoder(c.Writer)

	if err := encoder.Encode(obj); err != nil { // 发生异常
		http.Error(c.Writer, err.Error(), 500)
	}
}

// 直接输出字节流信息
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// 响应HTML数据
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html") // 纯文本格式
	c.Status(code)
	c.Writer.Write([]byte(html))
}

