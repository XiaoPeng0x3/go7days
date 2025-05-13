// 实现一致性哈希
// 结构包括：
// - 哈希环
// - 真正的key值
// - 虚拟结点
// - 可供用户自定义的hash函数
// - 虚拟结点与真实结点的映射

// 请求的流程：
// 首先根据机器的地址去进行编号，例如有三台机器`a`, 'b', 'c'。
// 为了实现负载均衡，我们需要给这些结点创建一系列的虚拟结点，将这些虚拟结点映射到哈希环上。
// 当一个`key`来到时，首先使用`hash`算法计算出`hash`值，然后去`hash`环上进行顺序搜索
// 搜索到的第一个`hash`值就是要添加的主机

package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func([]byte) uint32

type Map struct {
	// 用户自定义函数
	hash Hash
	// 每个真实结点的虚拟结点数
	replicas int
	// 哈希环
	keys []int // Sorted
	// 虚拟结点与真实结点的映射
	hashMap  map[int]string
}

func New(replicas int, fn Hash) *Map {
	m :=  &Map{
		hash: fn,
		replicas: replicas,
		keys: make([]int, 0),
		hashMap: make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	// 遍历所有的keys, 给每个key都创建虚拟结点
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 创建虚拟结点
			hash := int(m.hash([]byte(strconv.Itoa(i) + key))) // 返回的是一个哈希值
			// 建立映射关系
			m.hashMap[hash] = key
			// 添加到哈希环里面
			m.keys = append(m.keys, hash)
		}
	}
	// 排序
	sort.Ints(m.keys) // 可以进行二分查找
}

// Get方法
func (m *Map) Get(key string) string {
	if len(m.hashMap) == 0 { // 哈希环无映射
		return ""
	}
	// 根据key计算出哈希值
	hash := m.hash([]byte(key))
	// 使用二分查找
	idx := sort.Search(len(m.keys), func (i int) bool {
		return m.keys[i] >= int(hash)
	}) // 找到第一个 >= hash的`hash`值的idx，因为是升序，可以使用二分
	return m.hashMap[m.keys[idx % len(m.keys)]]
}

