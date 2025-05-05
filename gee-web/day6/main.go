package main

/*
(1) index
curl -i http://localhost:9999/index
HTTP/1.1 200 OK
Date: Sun, 01 Sep 2019 08:12:23 GMT
Content-Length: 19
Content-Type: text/html; charset=utf-8
<h1>Index Page</h1>

(2) v1
$ curl -i http://localhost:9999/v1/
HTTP/1.1 200 OK
Date: Mon, 12 Aug 2019 18:11:07 GMT
Content-Length: 18
Content-Type: text/html; charset=utf-8
<h1>Hello Gee</h1>

(3)
$ curl "http://localhost:9999/v1/hello?name=geektutu"
hello geektutu, you're at /v1/hello

(4)
$ curl "http://localhost:9999/v2/hello/geektutu"
hello geektutu, you're at /hello/geektutu

(5)
$ curl "http://localhost:9999/v2/login" -X POST -d 'username=geektutu&password=1234'
{"password":"1234","username":"geektutu"}

(6)
$ curl "http://localhost:9999/hello"
404 NOT FOUND: /hello
*/

import (
	"net/http"
	"time"
	"gee"
	"log"
)

func onlyForV2() gee.HandlerFunc {
	return func(c *gee.Context) {
		// Start timer
		t := time.Now()
		// if a server error occurred
		c.Fail(500, "Internal Server Error")
		// Calculate resolution time
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func main() {
	r := gee.New()
	r.Use(gee.Logger()) // global midlleware
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	})

	v2 := r.Group("/v2")
	v2.Use(onlyForV2()) // v2 group middleware
	{
		v2.GET("/hello/:name", func(c *gee.Context) {
			// expect /hello/geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
	}

	r.Run(":9999")
}