package main

import (
	"net/http"
	"fmt"
	"log"
	"geecache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	geecache.NewGroup("scores", 2 << 10, geecache.GetterFunc( // 找不到数据会调用这个回调函数
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if val, ok := db[key]; ok {
				return []byte(val), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	// 创建测试结点
	addr := "localhost:9999"
	peers := geecache.NewHTTPPool(addr)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}