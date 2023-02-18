// test.conflict_test.go
package test

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

/*
有时打印2次，有时4次，有时panic，
是多个goroutine对同一个数据结构竞争造成的
*/
func TestConflict(t *testing.T) {
	var m = make(map[int]bool, 0)
	printOnce := func(n int, m map[int]bool) {
		if _, ok := m[n]; !ok {
			fmt.Println(n)
		}
		m[n] = true
	}
	for i := 0; i < 10; i++ {
		go printOnce(100, m)
	}
	time.Sleep(time.Second * 1)
}

/* 用锁解决数据竞争问题 */
func TestUnconflictWithLock(t *testing.T) {
	var lock sync.Mutex
	var m = make(map[int]bool, 0)
	printOnceWithLock := func(n int, m map[int]bool, lock *sync.Mutex) {
		lock.Lock()
		if _, ok := m[n]; !ok {
			fmt.Println(n)
		}
		m[n] = true
		lock.Unlock()
	}
	for i := 0; i < 10; i++ {
		go printOnceWithLock(100, m, &lock)
	}
	time.Sleep(time.Second * 3)

}
