/*
将缓存的数据抽象为byteview类型
*/

package geecache

// 将缓存的数据封装为一个只读的byteview数据
type ByteView struct {
	b []byte
}

// 获取长度
func (bv ByteView) Len() int {
	return len(bv.b)
}

// 获取拷贝值
func (bv ByteView) ByteSlice() []byte {
	return cloneBytes(bv.b)
}


// 私有拷贝方法封装
func cloneBytes(b []byte) []byte {
	res := make([]byte, len(b))
	copy(res, b)
	return res
}

// strings
func (bv ByteView) String () string {
	return string(bv.b)
}