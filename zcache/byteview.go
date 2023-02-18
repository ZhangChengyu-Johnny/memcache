// ZCache.zcache.byteview.go
package zcache

/* 统一缓存值的类型，用[]byte可以支持图片、字符串等任何类型数据 */
type ByteView struct {
	b []byte
}

/* 返回缓存值占用的内存大小(单位字节)，实现Data(缓存值类型)接口 */
func (v ByteView) Len() int {
	return len(v.b)
}

/* 由于ByteView.b是指针类型，所以需要返回一个数据深拷贝，防止缓存值被外部程序污染 */
func (v ByteView) Bytes() []byte {
	return deepCopy(v.b)
}

/* 提供字符串类型的返回 */
func (v ByteView) String() string {
	return string(v.b)
}

/* 深拷贝一份 */
func deepCopy(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
