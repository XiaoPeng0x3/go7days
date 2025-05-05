package gee

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"net/http"
)

func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // 返回跳过后的可执行数目
	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func (c *Context) Recover() HandlerFunc{
	return func(c *Context) {
		defer func(){
			if err := recover(); err != nil {
				// 说明发生错误
				// 这里要捕获出错逻辑
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		// 中间件函数
		c.Next()
	}
}