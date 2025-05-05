# 七天Go系列Web总结
## 前言
通过tutu这个`Go-Web`系列，确实学习到了许多东西，很多东西与`Gin`框架十分相像，在使用完这个框架后，后续可以继续使用`Gin`框架来加深学习。
## Day1
在`Day1`中，`tutu`演示了如何去启动一个http服务，对于一个http服务来说，底层实现的细节不需要我们去掌握,需要了解的内容是：
- 路由注册
- 实现对应的业务逻辑
- 启动服务
### 路由注册
路由注册是服务端提供给用户端的一个访问路径。注册路由的时候，通常需要指定一个路径，只有通过这个路径才可以享受服务端提供的各种服务，如果用户访问的不是我们注册的路径，那么就应该返回404表示找不到对应资源。
### 业务逻辑
在访问路径后，服务端需要给用户端提供一些服务，一般来说会把这些服务封装为一个函数，在这个函数内实现一些业务逻辑。例如，使用`http`包下的`HandleFunc`就可以将访问路径与业务函数绑定在一起，就像下面这样：
``` Go[]
import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/name", printHello)
}

func printHello(w http.ResponeWriter, req *http.Request) {
    fmt.Println("Url = [%q]", req.URL.Path)
}
```

值得注意的是,这个`printHello`函数的函数参数是固定的，其函数参数必须是上文中写到的那样。
这样，我们就把业务逻辑与"/hello"这个`URL`绑定在一起，当用户访问这个路径的时候，服务端就会去执行`printHello`这个函数里面的内容(当然，必须得启动服务)

### 启动服务
启动服务也十分简单，只要一行代码即可
```Go
http.ListenAndServe(":9999", nil)
```
意思是我们在`localhost:9999`端口启动了服务(`localhost`可以不写)，第二个参数传递了一个`nil`，这个一会再说。

当然，这个函数也是有返回值`err`的，如果`9999`

### ServeHTTP
在之前的实现中，手动绑定所有函数似乎有些过于繁琐。假设要绑定很多个服务，也不利于维护。幸好，`http`提供了一种类似于“请求入口”的接口，就是`ServeHTTP`这个接口，这个接口可以接受所有的`HTTP`请求。所以，在启动服务的时候，我们可以传递一个实现了这个接口的实现类，然后由这个类来管理所有的请求。

```Go
// 实现ServeHTTP接口
type Engine struct{}

// 接受所有的HTTP请求
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {

    switch req.URL.Path {
        case "/":
            // 业务逻辑
            // Do sth
        case "/hello":
            // 业务逻辑
            // printHello
        default:
            // 404 Not Found
    }
}
```

这样，在启动服务的时候就可以把`Engine`对象传入，从而起到控制作用
```Go
http.ListenAndServe(":9999", &Engine)
```

### 封装Engine
在上文中，这种根据访问路径去寻找的方式似乎还是有些麻烦。其实在这个时候，添加一张路由表已经十分自然了，我们可以将访问路径与之对应业务逻辑函数封装起来，封装到一张哈希表里面，调用方法时直接取出对应的函数即可。
因为这些函数都有一个共同的函数参数，可以将这个函数类型封装起来

```Go
type HandlerFunc func(http.ResponeWriter, *http.Request)


type Engine struct {
    router map[string]HandlerFunc
}

// 在注册路由的时候，可以根据请求的方法和URL来作为key

func (engine *Engine) addRouter(method string, pattern string, handler HandlerFunc) {
    // 拼接key
    key := method + "-" + "pattern" 
    engine.router[key] = handler
}
```

当用户端请求时，所有的请求都在`ServeHTTP`函数里面，我们需要改造一个这个函数

```Go 
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // 得到key
    // 从http请求里面获取
    key := req.Method + "-" + req.URL.Path
    // 查看router里面注册这个路由
    if handler, ok := engine.router[key]; ok { // 表示存在
        handler(w, req)
    } else {
        // Not Found
    }
}
```

同时，对外封装几个常用的服务：
```Go
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
```

## Day2
### 封装上下文context对象
封装`context`对象是因为`http`提供的功能太过于基础和细致，功能不够强大，同时为了后面实现动态路由，我们将这个`context`抽取出来。设计为下面这样

```Go
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
}
```

同时也可以将路由的功能分离出来，目前来看，改动不多。

## Day3
### 实现动态路由
在之前的路由实现中，我们只可以实现了静态路由，而动态路由实现了占位的效果。假设我的文件路径下有许多`docs`文档，例如
- p/c/docs
- p/python/docs
- p/java/docs
- p/go/docs

实际上这些路由都十分相像，如果需要使用静态路由，那么就需要重复注册许多路由。我们可以使用动态路由来起到“占位”的作用，对于这些路由，我们可以只注册一个路由即可： /p/:lang/docs
在匹配的时候，这一个路由就会去匹配上面的所有请求。

动态路由算是比较难理解的一部分了，先来想一下整个动态路由应该实现哪些内容。由于是前缀树的实现，这里就参考tutu原文里面的设计。

```Go
type node struct {
	// 待匹配的路由
	pattern string
	// 路由中的一部分
	part string
	// 子节点
	child [] *node 
	// 是否精确匹配
	isWiled bool
}

```
`pattern`这个字段是为了方便寻找才构建的，如果不是叶子结点，那么其`pattern`都为空，只有在叶子结点的时候才为其整个的路径一次来表示确实找到了。
`isWiled`表示是否是模糊匹配，如果不是动态路由的话是不会去模糊匹配的，即为`false`。

#### 插入结点
在插入结点的时候，为了方便赋值`part`部分，我们将注册的路由进行分割拆分。例如，如果一个路由是`p/java/docs`，经过拆分后就变为["p", "java", "docs"]。 

同样的，如果一个是动态路由`p/:lang/docs`时，也会切分为["p", ":lang", "docs"]。

这是插入函数的签名，下面来说一下思路
```Go
func (n *node) insert(pattern string, parts []string, height int) {}
```

首先，`pattern`是注册的路由，由于是递归实现，来考虑一下实现细节
- base case
什么时候是递归的结束呢，即将`parts`中的所有`part`全部插入完毕，即 `len(parts) == height`的时候，说明插入完毕，这时候需要把这个结点的`pattern`设置为传递的`pattern`来表示是叶子结点

- 中间的逻辑
首先要找到插入的位置`child`，找到插入的位置后，需要将当前结点的`part`设置为`parts[height]`，这样就将当前结点`part`设置好了

- 递归
让找到`child`之后，为了能够成功插入，继续向下一层递归即可，`child.insert(pattern, parts, height+1)

这里面有一个bug
假设第一次注册的路由是 /p/:lang/doc
然后又注册了一个路由 /p/go/doc
那么第一个结点的pattern就会由 "/p/:lang/doc" 变为 "/p/go/doc"
虽然，当一个请求 /p/python/doc来到时依然会去匹配到这个路由， 但是 在router.go 100行 生成key时就会有些问题
这个时候生成的key是 ”GET-"/p/go/doc" 而不是 ”GET-"/p/：lang/doc“
此时就会把新到来的处理函数“错误绑定”， 变为{”GET-"/p/go/doc": handlePython}
而不是 {”GET-"/p/：lang/doc": handlePython}
也就是说，"/p/go/doc"这个路由并没有被注册，但是却实实在在的影响力“/p/:lang/doc”的叶子结点的pattern