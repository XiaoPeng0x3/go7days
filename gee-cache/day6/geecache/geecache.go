// 负责与外部交互，控制缓存存储和获取的主流程
package geecache

import (
	"fmt"
	"geecache/singleflight"
	"log"
	"sync"
)

/*
下面的设计技巧是“适配器”模式，也就是使用函数去实现这个接口
定义一个函数类型 F，并且实现接口 A 的方法，然后在这个方法中调用自己。
这是 Go 语言中将其他函数（参数返回值定义与 F 一致）转换为接口 A 的常用技巧。
*/

// 回调getter
type Getter interface {
	// 定义一个Getter函数
	Get (string) ([]byte, error)
}

// 使用函数实现这个接口
type GetterFunc func(string) ([]byte, error)

// 实现接口
func (get GetterFunc) Get(key string) (bytes []byte, err error) {
	return get(key)
}

// 创建分组Group
type Group struct {
	// 分组的名称
	name string
	// 当缓存未命中时，需要从远处获取数据
	getter Getter // 其实就是去要求数据库去实现这个接口
	// 缓存
	mainCache cache
	// 当缓存未命中时需要加载数据
	peers Peerpick
	// 加入singleflight策略
	singleflight *singleflight.CallGroup
}

var (
	// 读写锁
	mu sync.RWMutex
	// 分组管理group
	groups = make(map[string] *Group)
)

// 创建新的group
// 主要是做数据分类，例如 name-group, age-group这些
func NewGroup(name string, caheBytes int64, getter Getter) *Group{
	if getter == nil {
		panic("getter is nil")
	}
	// 上锁
	mu.RLock()
	defer mu.RUnlock()
	g := &Group{
		name: name,
		getter: getter,
		mainCache: cache{cacheBytes: caheBytes},
		singleflight: &singleflight.CallGroup{},
	}
	// 上锁是因为map也会有并发问题
	groups[name] = g
	return g
}

// 注册peers
func (g *Group) RegisterPeers(peers Peerpick) {
	if g.peers != nil {
		log.Println("RegisterPeers more than once")
	}
	g.peers = peers
}


// 加入从其它结点调用数据
func (g *Group) load(key string) (value ByteView, err error) {
	// 先从其它结点调用
	views, err := g.singleflight.Do(key, func(key string) (interface{}, error) { // 数据不存在时会触发回调函数
		// 先从远程结点调用
		if g.peers != nil {
			// 选择结点
			log.Println("geecache.go: 选择结点: 88行")
			if peer, ok := g.peers.PickPeer(key); ok {
				if view, err := g.getFromPeer(peer, key); err != nil {
					return view, err
				}
			}
		}
		// 从本地调用
		return g.getLocally(key)
	})
	
	//
	if err == nil { // 调用正常
		return views.(ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (value ByteView, err error) {
	if key == "" {
		log.Fatal("None key")
	}
	if bytes, err := peer.Get(g.name, key); err == nil {
		return ByteView{b: bytes}, nil
	}
	return
}



// 从groups里面获取某个group
func GetGroup(key string) *Group {
	// 上锁
	mu.RLock()
	defer mu.RUnlock()
	return groups[key] // 不存在会返回nil
}

// Get value for a key from cache
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is nil")
	}

	// 从缓存里面获取数据
	// log.Println(g.mainCache)
	// 初次调用的时候可能 lru.lru会未初始化，需要判断
	if val, ok := g.mainCache.get(key); ok {
		// cache hit
		log.Println("cache hit")
		return val, nil
	}
	// 不存在需要从远程加载数据
	return g.load(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// 从数据源获取信息
	bytes, err := g.getter.Get(key)
	if err != nil {
		// 本地也没有
		return ByteView{}, err
	}
	// 有的话需要添加在cache里面, 拷贝bytes
	val := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, val)
	return val, nil
}

func (g *Group) populateCache(key string, byteView ByteView) {
	g.mainCache.add(key, byteView)
}