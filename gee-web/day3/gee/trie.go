// 设计一个路由匹配的前缀树
package gee

import (
	"fmt"
	"strings"
)


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

// ToSting方法
func (n *node) String() string {
	return fmt.Sprintf("node.pattern = [%s], node.part = [%s], node.isWield = [%t]", n.pattern, n.part, n.isWiled)
}

// 向Trie树里面插入
// 可以看作构建前缀树的过程 

// 这里面有一个bug
// 假设第一次注册的路由是 /p/:lang/doc
// 然后又注册了一个路由 /p/go/doc
// 那么最后一个结点的pattern就会由 "/p/:lang/doc" 变为 "/p/go/doc"
// 虽然，当一个请求 /p/python/doc来到时依然会去匹配到这个路由， 但是 在router.go 100行 生成key时就会有些问题
// 这个时候生成的key是 ”GET-"/p/go/doc" 而不是 ”GET-"/p/：lang/doc“
// 此时就会把新到来的处理函数“错误绑定”， 变为{”GET-"/p/go/doc": handlePython}
// 而不是 {”GET-"/p/：lang/doc": handlePython}
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height { // 说明插入到了最后一个part，此时pattern构造完毕
		n.pattern = pattern // 构造完毕
		return
	}
	// 找到要插入的地方，也就是找到子节点的位置
	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		// 新建一个child
		child = &node{
			part: part,
			isWiled: part[0] == ':' || part[0] == '*', // 是否是动态路由
		}
		n.child = append(n.child, child)
	}
	child.insert(pattern, parts, height+1)
}

// 根据parts去搜索是否存在于路径中，如果存在，那么就返回那个node结点
func (n *node) search(parts []string, height int) *node {
	// 什么情况才算是可以停止搜索呢
	// 1. len(parts) == height， 表示搜索到最后了
	// 2. parts[height] 中有通配符，所以到通配符这一层就可以不用进行搜索了
	if (len(parts) == height || strings.HasPrefix(n.part, "*")) { 
		if n.pattern == "" { // 如果是完整的路径，那么n.pattern就会记录完整路径表示结尾
			return nil
		}
		return n
	}
	// 找到n的所有孩子
	children := n.matchChildren(parts[height])
	for _, child := range children {
		// child结点去搜索， 看是否可以找到
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}

// 匹配child, 返回找到的结点
func (n *node) matchChild(part string) *node {
	// 根据part来进行查找

	// 获取n所有的孩子结点
	for _, child := range n.child {
		if child.part == part || child.isWiled { // 精确匹配或者模糊匹配
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	// 获取所有的part的子路径
	res := make([]*node, 0)

	for _, child := range n.child {
		if child.part == part || child.isWiled {
			res = append(res, child)
		}
	}
	// 返回所有匹配的路径
	return res
}