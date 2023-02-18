package consistenthash

import (
	"fmt"
	"strconv"
	"testing"
)

func testHashFunc(key []byte) uint32 {
	// 为了知道每个key的哈希值，不能使用默认哈希算法，使用简单的测试函数
	i, _ := strconv.Atoi(string(key))
	return uint32(i)
}

func TestHashing(t *testing.T) {
	// 每个节点3个虚拟节点，使用默认hash函数
	hash := New(3, testHashFunc)
	// 添加3个节点，对应虚拟节点是[02 12 22 04 14 24 06 16 26]
	// 排序后获得[2, 4, 6, 12, 14, 16, 22, 24, 26]
	hash.Add("6", "4", "2")

	testCases := map[string]string{
		"2":  "2", // "2" 命中虚拟节点 02
		"11": "2", // "11"命中虚拟节点 12
		"23": "4", // "23"命中虚拟节点 24
		"27": "2", // "27"命中虚拟节点 02
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s, but have %s", k, v, hash.Get(k))
			return
		}
	}
	fmt.Println("test ok!")
}
