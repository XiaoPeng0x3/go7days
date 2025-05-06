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
### 实现分管
在完成前缀树部分后就可以将路由`router`的功能交给前缀树托管，在router里需要维护一个前缀树的`root`结点。其维护思路是这样的：
- 根据不同方法来进行维护
在`router`中根据`req`的不同方法来作为`key`，也就是说`GET`和`POST`会创建两棵前缀树，并且在`router`里面存储这两棵前缀树的根节点。
### 实现搜索
对于树类的数据结构，都可以使用递归来进行搜索，搜索的思路也可以分为：
- base case
当前`node.part = parts[height]`并且`node.pattern != nil`，说明找到了叶子结点，完成匹配。此时还要额外考虑通配符*，如果当前路由树上的`part`是`*filepath`，那么也可以返回，应为`*filepath`应该匹配所有路径
- 中间逻辑
对于中间逻辑，如果没有找到，那么应该获取当前`node`的所有`child`(记录在列表里面)，然后对所有`child`进行递归搜索
## Day4
### 路由分组
路由分组的功能就是将拥有公共前缀的`url`去分为一组，对于这一组的所有路由我们就可以使用中间件去对这些路由组进行管理，比如记录执行时间、访问控制等。
因此这里使用新的`RouteGroup`来进行记录，对于这个路由组来说，可以记录下面这些字段来进行封装
- prefix
记录这个路由组的前缀，也就是公共前缀部分
- 中间件
记录这个路由组的中间件，凡是属于这个路由组的子路径都要去执行中间件
- parent
记录上一级路由
- engine
所有路由组共用一个`engine`
这里的分组不是重点，其实是为后面的中间件实现做铺垫的，只需要了解什么是路由分组即可。
## Day5
### 中间件的实现
中间件的定义似乎很模糊，但大致上就是在请求到来前和结束后去实现的一系列内容。这里实现的中间件十分巧妙，其思路是：
- 将所有请求放在一个列表里面
因为所有请求的入口都会被汇集到`ServeHTTP`这个接口里面，这些都是来自用户的实际请求。此时，可以根据用户请求的`URL`路径来获取注册在该路由组的所有中间件方法，先整理好所有的中间件，先将所有的中间件存放在`c.handlers`里面。对于用户的那些请求，我们应该去前缀树上去根据路径去搜索，如果可以搜索到，那么在把业务逻辑存放在`c.handlers`里面，然后再调用`c.Next()`去执行。

这也就意味着所有的用户请求的实现部分都被`context`托管，并且都会在`c.Next()`里面实现调用

```Go
// 构造next函数
func (c *Context) Next() {
	c.index++
	// 遍历所有的中间件函数
	s := len(c.handlers)
	for ; c.index < s; c.index ++ {
		// 执行函数
		c.handlers[c.index](c) 
	}
}
```
### 中间件的洋葱模型
对于服务端来说，如果要使用中间件，`gin`的做法是使用`Use`函数并在函数内调用`c.Next()`，假设所有的请求有两个中间件`part1`和`part2`，那么其函数的执行顺序就是：
`part1 => part2 => user => part2 => part1`，可以看出这就是一个递归的过程，是因为在中间件中调用了`c.Next()`去循序执行所有的中间件，也就是说
```Go
func part1() {
    // do sth

    c.Next() // 就是去执行part2

    // do sth
}

```
正好形成了递归的调用结构

### 请求实例
可以使用`Use`方法来传递所有的中间件函数，`Use`就是将所有的中间件函数去添加到该路由组的中间件列表中，在中间件里面去调用`c.Next()`去执行下一个中间件函数。
中间件是支持分组的，所以就可以给某个路由组的所有子路由添加到中间件，作用范围是整个路由组。

## Day6
### 模板渲染
模板渲染是指服务端负责将数据渲染为完整的结构然后返回给浏览器，这其实是指前后端不分离的项目。在前后端分离的项目中，后端只需要考虑数据怎么返回给前端即可，前端只需要考虑如何更好的渲染来自后端的数据即可。如果服务端负责模板渲染，那么返回前端的就是直接可以被浏览器解析的数据，从而可以很快的将页面加载出来，提升用户体验。
### 文件映射
服务器在对外提供文件资源的时候，必须将用户请求的`url`映射为本地存在的文件目录。例如，用户请求"assets/css/main.css"，服务器存储的静态文件的路径是"usr/zxp/local/static/css/main.css",因此将`assets`映射为`usr/zxp/local/static/"，而对应后面的实际文件路径与新的地址作拼接，得到一个本地资源的路径。拼接后的地址就是`usr/zxp/local/static/css/main.css`。

## Day7
### 错误恢复
当框架出现错误时，例如数组越界、空指针等`panic`错误的时候，服务就会挂掉，为了不使服务挂掉，可以使用`defer + recover`实现错误恢复，这里起到的作用就是`try {// do sth } catch(e Exception) {错误恢复},这样就不会使服务挂掉。
### 添加recover中间件
可以将错误恢复的功能封装为一个中间件，当出现错误的时候，就使用`defer + recover`来接收。在`recover`里面封装了这两个函数
```Go
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
```

当出现`panic`时，即`recover() != nil`时，这个函数就会去调用`trace`函数去记录出错的`stack`，此时的函数调用链就是
``` txt
runtime -> trace -> Recover(中间件)
0       -> 1     -> 2
```
为了获取直接出错的部分，可以跳过前三个无用的pc，直接记录出错的`pc`。注意，`recover`只能和`defer`一起使用。

# 总结
做完这个七天go-web系列，收获最大的就是关于路由组以及中间件的知识，尤其是中间件链的执行流程这个过程。
