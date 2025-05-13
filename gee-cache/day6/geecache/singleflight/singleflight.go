// 防止缓存击穿的思路：
// 当有很多个请求到来缓存的时候，假设此时`key`正好过期，那么这些请求就回去到达后端数据库，造成数据库压力巨大。
// 因为许多个请求都是请求一个东西，所以只需要一个人去请求即可，其他人只需要在那里等待请求返回即可

package singleflight

import (
	// "log"
	"sync"
)

// 封装请求结构体
type call struct {
	// 请求组
	wg sync.WaitGroup
	// 数据
	value interface{}
	// err
	err error
}

// 存储请求对象
type CallGroup struct {
	// 请求对象
	calls map[string] *call
	// 锁
	mu sync.Mutex
}

// 定义Do
func (cg *CallGroup) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 首先判断当前请求是否存在于map中
	// log.Println("Do in singleflight")
	cg.mu.Lock()
	if cg.calls == nil {
		cg.calls = make(map[string]*call)
	}
	if call, ok := cg.calls[key]; ok { // 存在，那么其它请求等就可以了
		// log.Println("singleflight.go: 请求已经存在，等待其它请求返回！")
		cg.mu.Unlock()
		call.wg.Wait() // 等待
		return call.value, call.err // 返回数据
	}
	cg.mu.Unlock()
	// 不存在
	call := new(call)
	call.wg.Add(1) // 添加一个线程
	cg.mu.Lock()
	cg.calls[key] = call // 添加进去
	cg.mu.Unlock()

	// 根据回调函数，获取数据
	call.value, call.err = fn()
	// log.Printf("singleflight.go: 调用回调函数获取数据, value= %v, err= %v", call.value, call.err)
	call.wg.Done() // 等待完毕

	// 请求从map里面删除
	cg.mu.Lock()
	// log.Println("singleflight.go: 所有请求调用完毕，删除请求记录!")
	delete(cg.calls, key)
	cg.mu.Unlock()

	// 返回数据
	//log.Panicln(value, err)
	// log.Printf("singlefilght.go 获得的数据为: value= %v, err= %v", call.value, call.err)
	return call.value, call.err
}
