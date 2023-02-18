package test

import (
	"ZCache/zcache"
	"fmt"
	"log"
	"testing"
)

var dbs map[string]string = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

var loadCounts map[string]int = make(map[string]int, len(dbs))

func SearchData(key string) ([]byte, error) {
	log.Println("[SlowDB] search key", key)
	if v, ok := dbs[key]; ok {
		if _, ok := loadCounts[key]; !ok {
			loadCounts[key] = 0
		}
		loadCounts[key] += 1
		return []byte(v), nil
	}
	return nil, fmt.Errorf("%s not exists", key)
}

func TestGet(t *testing.T) {
	zee := zcache.NewGroup("scores", 2<<10, zcache.GetterFunc(SearchData))

	for k, v := range dbs {
		// 新建缓存
		if view, err := zee.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}

		// 从缓存中获取，验证计数器
		if _, err := zee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := zee.Get("unknow"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}

}
