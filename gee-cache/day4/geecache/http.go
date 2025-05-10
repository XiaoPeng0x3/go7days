// 这个类是为了实现分布式结点之间的通信
// 通信协议使用http来实现
package geecache

import (
	"log"
	"net/http"
	"strings"
)

const baseUrl string = "/_geecache/"

type HTTPPool struct {
	// 自己的结点/主机号
	self string
	// 请求的统一前缀
	baseUrl string
}

// 封装当前地址和路径前缀的对象
// 结点之间的通信能力来自于http
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self: self,
		baseUrl: baseUrl,
	}
}

// 实现ServeHTTP
// 要求请求的 url必须符合 self/baseUrl/<groupName>/<key>
func (h *HTTPPool) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 首先判断请求是否合法
	// 获取url, 不包含主机
	if !strings.HasPrefix(req.URL.Path, h.baseUrl) { // 前缀不符合，我们认为这是一个错误请求
		http.Error(w, "Bad url request", http.StatusBadRequest)
		return // bug fix: http.Error并不会导致逻辑终结
	}
	// 解析 <groupName> 和 <key>
	parts := strings.SplitN(req.URL.Path[len(baseUrl):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "Request params error", http.StatusBadRequest)
		return // bug fix: http.Error并不会导致逻辑终结
	}
	// 得到groupName
	groupName := parts[0]
	// 得到请求的key
	key := parts[1]
	// 根据groupName找对应的group
	group := GetGroup(groupName)
	log.Println(group)
	if group == nil { // 找不到对应集群, not found
		http.Error(w, "Cant find group", http.StatusNotFound)
		return // bug fix: http.Error并不会导致逻辑终结
	}
	views, err := group.Get(key)
	if err != nil { // 缓存中不存在该数据
		http.Error(w, "No such Value", http.StatusInternalServerError)
		return // bug fix: http.Error并不会导致逻辑终结
	}
	// 把views数据写回去
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(views.ByteSilces())
}