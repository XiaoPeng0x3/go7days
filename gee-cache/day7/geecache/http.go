// 这个类是为了实现分布式结点之间的通信
// 通信协议使用http来实现
package geecache

import (
	"fmt"
	"geecache/consistenthash"
	pb "geecache/geecachepb"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
)

const baseURL string = "/_geecache/"

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	// 自己的结点/主机号
	self string
	// 请求的统一前缀
	baseURL string
	// lock
	mu sync.Mutex
	// peers
	peers *consistenthash.Map
	// getters
	httpGetters map[string] *httpGetter
}

// http请求的客户端
type httpGetter struct {
	baseURL string
}

// 实现PeerGetter的Get
func (h *httpGetter) Get(in *pb.Request, out *pb.Response) (error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()), // 将字符串转换为URL格式
		url.QueryEscape(in.GetKey()), // 字符串转换为URL格式
	)
	// 获取url连接
	resp, err := http.Get(u) // 得到response 对象
	if err != nil {
		return err
	}
	defer resp.Body.Close() // 关闭连接

	if resp.StatusCode != http.StatusOK { // 出错了
		return fmt.Errorf("StatusError: %v", resp.StatusCode)
	}
	// 读取resp Body
	bytes, err := io.ReadAll(resp.Body) // 读取数据
	if err != nil {
		return fmt.Errorf("IO Read Error: %s", err)
	}
	// 对数据进行反序列化
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}


// 封装当前地址和路径前缀的对象
// 结点之间的通信能力来自于http
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self: self,
		baseURL: baseURL,
	}
}

// 实现ServeHTTP
// 要求请求的 url必须符合 self/baseURL/<groupName>/<key>
func (h *HTTPPool) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 首先判断请求是否合法
	// 获取url, 不包含主机
	if !strings.HasPrefix(req.URL.Path, h.baseURL) { // 前缀不符合，我们认为这是一个错误请求
		http.Error(w, "Bad url request", http.StatusBadRequest)
		return // bug fix: http.Error并不会导致逻辑终结
	}
	// 解析 <groupName> 和 <key>
	parts := strings.SplitN(req.URL.Path[len(baseURL):], "/", 2)
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
	// 把得到的数据序列化为proto的形式 
	data, err := proto.Marshal(&pb.Response{Value: views.ByteSlice()});
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// 把views数据写回去
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

// 创建一致性哈希环，将这 peers节点加入。
// 为每个节点配置一个 HTTP 客户端，便于根据 key 请求对应节点获取数据。
func (h *HTTPPool) Set(peers ...string) { // peers一般是host
	// 加锁
	h.mu.Lock()
	defer h.mu.Unlock()

	// 创建新的hash环
	h.peers = consistenthash.New(defaultReplicas, nil) // 使用默认的虚拟节点数和默认的哈希函数
	// 将传入的结点添加到哈希环中
	h.peers.Add(peers...) // 会为这些结点创建虚拟节点
	// 创建http客户端
	h.httpGetters = make(map[string]*httpGetter, len(peers)) // 给每个结点分配一个http客户端
	for _, peer := range peers {
		peerHttpGetter := &httpGetter{baseURL: peer + h.baseURL}
		h.httpGetters[peer] = peerHttpGetter
	}
}

// PickPeer
// 根据key来选择对应的虚拟结点，然后得到 peer的名称
// 根据peer来选择它的对应的http客户端
func (h *HTTPPool) PickPeer(key string) (peer PeerGetter, ok bool) {
	// lock
	h.mu.Lock()
	defer h.mu.Unlock()

	// Get
	if p := h.peers.Get(key); p != "" && p != h.self { // 不能选择自己
		log.Printf("Pick peer : %s", p)
		return h.httpGetters[p], true // 根据具体的结点返回对应的http客户端
	}
	return
}