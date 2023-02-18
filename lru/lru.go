// lru.lru.go
package lru

import (
	"container/list"
)

/* 定义缓存值的类型 */
type Data interface {
	Len() int
}

/* 双向链表节点内的数据, 也就是*list.Element.Value */
type entry struct {
	// 多保存一份key的好处是淘汰头节点时可以用这个key直接从哈希表里删
	key string
	// 实际数据需要有Len()方法，返回占用内存大小
	data Data
}

/* 缓存结构 */
type Cache struct {
	maxBytes       int64                    // 允许使用的最大内存容量
	nbytes         int64                    // 当前已经使用的内存容量
	doubleLinkList *list.List               // 用来记录访问频率的双向链表
	cacheMap       map[string]*list.Element // 缓存哈希表
	OnEvicted      func(string, Data)       // 当删除缓存时会调用回调函数，可以为nil
}

/* 返回内存占用 */
func (c *Cache) cacheMemory(k string, d Data) int64 {
	return int64(len(k) + d.Len())
}

/* 返回链表长度 */
func (c *Cache) Len() int {
	return c.doubleLinkList.Len()
}

func New(maxBytes int64, OnEvicted func(string, Data)) *Cache {
	return &Cache{
		maxBytes:       maxBytes,
		nbytes:         0,
		doubleLinkList: list.New(),
		cacheMap:       make(map[string]*list.Element),
		OnEvicted:      OnEvicted,
	}
}

/* 新增 & 更新 */
func (c *Cache) Add(newK string, newD Data) {
	if node, ok := c.cacheMap[newK]; ok {
		// 新增缓存如果已存在，那么移动到队尾
		c.doubleLinkList.MoveToFront(node)
		// 更新缓存
		entry := node.Value.(*entry)
		entry.data = newD
		// 更新缓存占用，由于key没变，直接计算data差值即可
		c.nbytes += int64(newD.Len() - entry.data.Len())
	} else {
		// 如果不存在，那么新增到链表尾部
		newNode := c.doubleLinkList.PushFront(&entry{newK, newD})
		c.cacheMap[newK] = newNode
		c.nbytes += c.cacheMemory(newK, newD)
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		// 缓存超过内存限制后，自动删除
		c.RemoveOldest()
	}
}

/* 查找 */
func (c *Cache) Get(k string) (data Data, ok bool) {
	if node, ok := c.cacheMap[k]; ok {
		// 缓存命中，移动到链表尾部
		c.doubleLinkList.MoveToFront(node)
		return node.Value.(*entry).data, ok
	}
	return nil, false
}

/* 删除 */
func (c *Cache) RemoveOldest() {
	// 获取链表Head节点
	head := c.doubleLinkList.Back()
	if head != nil {
		// 删除Head节点
		c.doubleLinkList.Remove(head)
		entry := head.Value.(*entry)
		// 从哈希表里删除
		delete(c.cacheMap, entry.key)
		// 更新已用空间
		c.nbytes -= c.cacheMemory(entry.key, entry.data)
		if c.OnEvicted != nil {
			// 如果设置了回调函数就调用
			c.OnEvicted(entry.key, entry.data)
		}
	}
}
