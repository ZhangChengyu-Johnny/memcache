// ZCache.singleflight.singleflight.go
package singleflight

import "sync"

/* 正在进行中，或已经结束的请求，使用sync.WaitGroup锁避免重入 */
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

/* singleflight主数据结构，管理不同的key的请求(call) */
type Group struct {
	mu sync.Mutex // 保护m不被并发读写
	m  map[string]*call
}

/*
接收2个参数，第1个参数是key，第2个参数是函数fn，
Do的作用是针对相同的key，无论Do被调用多少次，fn只调用一次，等到fn调用结束了，返回返回值或错误
*/
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock() // 循环之前上锁，不让并发修改
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()       // 如果找到key就解锁
		c.wg.Wait()         // 如果正在请求缓存服务器时，则等待
		return c.val, c.err // 等到请求缓存服务器结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)   // 发起请求前加锁
	g.m[key] = c  // 添加到g.m，表示key已经有对应的请求在处理中
	g.mu.Unlock() // 添加完key后解锁，此时如果并发去读会卡在wg锁上

	c.val, c.err = fn() // 调用fn，也就是向缓存服务器发起请求
	c.wg.Done()         // 请求结束后将锁移除

	g.mu.Lock()
	delete(g.m, key) // 更新g.m
	g.mu.Unlock()

	return c.val, c.err // 返回缓存节点的响应结果
}
