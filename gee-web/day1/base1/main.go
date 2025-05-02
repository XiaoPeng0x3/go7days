// 标准库启动web服务
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 第一个参数是路由地址， 第二个参数是需要传递的函数
	http.HandleFunc("/", printHello)
	http.HandleFunc("/hello", printHeader)
	log.Fatal(http.ListenAndServe(":9999", nil))
}
// print sth
func printHello(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("URL = %q\n", req.URL.Path)
}

func printHeader(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Printf("Header[%q] = %q\n", k, v)
	}
}