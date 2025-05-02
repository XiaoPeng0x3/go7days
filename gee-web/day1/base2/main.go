// 实现http的接口
package main

import (
	"fmt"
	"log"
	"net/http"
)

// 实现继承关系
type Engine struct{}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 
	switch req.URL.Path{
	case "/":
		fmt.Printf("URL = %q\n", req.URL.Path)
	case "/hello":
		fmt.Printf("URL = %q\n", req.URL.Path)
	default:
		fmt.Printf("404 not found %q", req.URL.Path)
	}
}

// 只要实现这个接口，我们就不必要去手动绑定所有路径，同时还可以将所有的请求拦截在我们的自定义实现中

func main() {
	// 启动监听
	log.Fatal(http.ListenAndServe(":9090", &Engine{}))
}
