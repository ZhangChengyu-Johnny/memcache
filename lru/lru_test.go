// lru.lru_test.go
package lru

import (
	"log"
	"reflect"
	"testing"
)

type S string

func (s S) Len() int {
	return len(s)
}

/* 添加 & 查询测试 */
func TestLRU(t *testing.T) {
	cache := New(0, nil) // 实例化一个无内存限制的缓存
	cache.Add("k1", S("v1"))
	f1, f2 := true, true

	if d, ok := cache.Get("k1"); !ok || string(d.(S)) != "v1" {
		t.Fatalf("cache hit 'k1=v1' failed")
		f1 = false
	}

	if _, ok := cache.Get("k2"); ok {
		t.Fatalf("cache miss 'k2' failed")
		f2 = false
	}

	if f1 && f2 {
		log.Println("cache.Add, cache.Get test ok!")
	}
}

/* 删除缓存时回调函数测试 */
func TestLRUOnEvicted(t *testing.T) {
	removeKeys := make([]string, 0)
	callBackF := func(k string, v Data) {
		// 缓存被删除时调用
		removeKeys = append(removeKeys, k)
	}
	cache := New(8, callBackF)

	cache.Add("k1", S("v1"))
	cache.Add("k2", S("v2"))
	// 缓存达到内存上限
	cache.Add("k3", S("v3")) // 删除k1=v1
	cache.Add("k4", S("v4")) // 删除k2=v2

	if !reflect.DeepEqual(removeKeys, []string{"k1", "k2"}) {
		t.Fatalf("call OnEvicted failed, expect keys: %s", []string{"k1", "k2"})
	} else {
		log.Println("cache.OnEvicted test ok!")
	}
}

/* 删除测试 */
func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"
	memoryCap := int64(len(k1 + k2 + v1 + v2))
	cache := New(memoryCap, nil)
	cache.Add(k1, S(v1))
	cache.Add(k2, S(v2))
	cache.Add(k3, S(v3))

	if _, ok := cache.Get("k1"); ok || cache.Len() != 2 {
		t.Fatalf("RemoveOldest 'k1=v1' failed")
	} else {
		log.Println("cache.RemoveOldest test ok!")
	}
}
