// consistenthash.consistenthash.go
package consistenthash

import (
	"hash/crc32"
	"strconv"
)

/* 快速排序 */
func quickSort(arr []int, left, right int) {
	if left >= right {
		return
	}
	mid := partitionSort(arr, left, right)
	quickSort(arr, left, mid-1)
	quickSort(arr, mid+1, right)
}

func partitionSort(arr []int, left, right int) int {
	tmp := arr[left] // 对tmp排序
	for left < right {
		for left < right && arr[right] >= tmp {
			right--
		}
		arr[left] = arr[right]
		for left < right && arr[left] <= tmp {
			left++
		}
		arr[right] = arr[left]
	}
	arr[left] = tmp
	return left
}

/* 二分查找，如果没有就返回可插入位置 */
func binarySearch(arr []int, hashValue int) int {
	i, j := 0, len(arr)
	for i < j {
		mid := (i + j) >> 1
		if arr[mid] < hashValue {
			i = mid + 1
		} else {
			j = mid
		}
	}
	return i
}

/* 用来计算数据的哈希值 */
type Hash func(data []byte) uint32

/* 虚拟节点 */
type ConsistentHash struct {
	hashFunc     Hash           // 哈希函数
	virtualCount int            // 虚拟节点的个数，默认50
	hashRing     []int          // 哈希环
	nodeMap      map[int]string // 虚拟节点和真实节点的映射,{虚拟节点哈希值: 真实节点主机名称}
}

func New(virtualCount int, hashFunc Hash) *ConsistentHash {
	// 允许自定义虚拟节点倍数和哈希函数
	m := &ConsistentHash{
		hashFunc:     hashFunc,
		virtualCount: virtualCount,
		nodeMap:      make(map[int]string),
	}
	if m.hashFunc == nil {
		// 默认用标准库提供的哈希算法
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

/* 可以传入0或多个真实节点的名称 */
func (h *ConsistentHash) Add(nodes ...string) {
	for _, node := range nodes {
		for i := 0; i < h.virtualCount; i++ {
			// 可以为每个节点配置不同的虚拟节点达到负载均衡效果
			hashValue := int(h.hashFunc([]byte(strconv.Itoa(i) + node)))
			h.hashRing = append(h.hashRing, hashValue)
			h.nodeMap[hashValue] = node
		}
	}
	quickSort(h.hashRing, 0, len(h.hashRing)-1) // 对哈希环排序
}

/* 根据缓存键获取节点 */
func (h *ConsistentHash) Get(cacheKey string) string {
	// 对数据计算出真实节点
	if len(h.hashRing) == 0 {
		return ""
	}
	hashValue := int(h.hashFunc([]byte(cacheKey)))
	// 二分查找，找到顺时针第一个虚拟节点
	idx := binarySearch(h.hashRing, hashValue)
	return h.nodeMap[h.hashRing[idx%len(h.hashRing)]]
}
