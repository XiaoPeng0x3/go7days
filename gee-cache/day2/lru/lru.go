package lru

import (
	"container/list"
)

// 创建cache结构体
type Cache struct {
	// 当前cache支持的最大存储
	maxBytes int64
	// 已经使用的最大内存数
	nbytes   int64
	// 双向链表的指针
	ll       *list.List
	// key是需要存储的key，val是对应的链表结点
	cache    map[string]*list.Element
	// optional and executed when an entry is purged.
	// 回调函数
	OnEvicted func(key string, value Value)
}

type entry struct {
	key string
	value Value
}

// 使用interface可以使Value接受任意类型的数据
type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache{
	return &Cache {
		maxBytes: maxBytes,
		nbytes: 0,
		ll: list.New(),
		cache: make(map[string] *list.Element),
		OnEvicted: onEvicted,
	}
}

// 实现查找
func (c *Cache) Get(key string) (value Value, ok bool) {
	// 根据key去cache里面查找
	if ele, ok := c.cache[key]; ok { // cache hit
		// 将element移动到队首
		c.ll.MoveToFront(ele)
		// 获取element的数据
		kv := ele.Value.(*entry) // 类型断言，也就是强制类型转换
		return kv.value, true
	}
	// cache miss
	return
}

// 实现插入新增逻辑
// 1. 判断是否已经存在，存在需要更新value值和占用空间值
// 2. 不存在需要新建这个值，并且移动到队首
// 3. 如果此时插入的值过多，需要使用LRU来删除队尾元素
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { // cache hit
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		kv.value = value
		// 更新占用值
		c.nbytes += (int64(value.Len())) - (int64(kv.value.Len())) // 新增数据的值 - 旧的数据值就是两次数据的变化值
	} else { // cache miss
		// 新建结点值
		newVal := &entry{key: key, value: value}
		ele := c.ll.PushFront(newVal)
		// 更新map
		c.cache[key] = ele
		// 更新占用空间
		// 占用空间包括 key的size和实际的数据大小
		// TODO 可以考虑内存对齐
		c.nbytes += int64(value.Len()) + int64(len(key))
	}
	// LRU淘汰
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// 删除的时候可能要有日志操作
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil { // 可以判断链表是否为空
		// 删除最后一个
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// 包括 实际存储+key
		c.nbytes -= int64(kv.value.Len()) + int64(len(kv.key))
		// 还需要从map里面删除
		delete(c.cache, kv.key)
		// 回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value) // 日志记录
		}
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}