package main
import (
	"fmt"
	"gee"
	"net/http"
)

func main() {
	// 创建两个请求， GET和POST
	r := gee.New()
	r.GET("/", func(w http.ResponseWriter, r *http.Request) {
		// 业务逻辑
		fmt.Printf("URL = %q\n", r.URL.Path)
		w.Write([]byte("GET"))
	})

	r.POST("/hello", func(w http.ResponseWriter, r *http.Request) {
		// 业务逻辑
		fmt.Printf("URL = %q\n", r.URL.Path)
		w.Write([]byte("POST"))
	})

	r.Run(":9999")

}