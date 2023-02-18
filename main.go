// ZCache.main.go
package main

import (
	"ZCache/zcache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db map[string]string = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func getSourceData(key string) ([]byte, error) {
	log.Println("[DB] search key", key)
	if v, ok := db[key]; ok {
		return []byte(v), nil
	}
	return nil, fmt.Errorf("%s not exist", key)
}

/* 缓存节点服务端 */
func startCacheServer(cacheAddr string, addrs []string, cache *zcache.Group) {
	server := zcache.NewCacheServerPool(cacheAddr)
	log.Println("cache is running at", cacheAddr)
	// 会调用server里的ServerHTTP方法
	err := http.ListenAndServe(cacheAddr[7:], server)
	if err != nil {
		fmt.Println(err)
		return
	}
}

/* 主服务端 */
func startAPIServer(apiAddr string, addrs []string, cache *zcache.Group) {
	server := zcache.NewCacheServerPool(apiAddr)
	server.Set(addrs...) // 实例化一致性哈希算法，加入节点
	cache.RegisterNodeSelector(server)
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := cache.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.Bytes())
		}))
	log.Println("main server is running at", apiAddr)
	err := http.ListenAndServe(apiAddr[7:], nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "cache server port")
	flag.BoolVar(&api, "api", false, "start a api server")
	flag.Parse()

	apiAddr := "http://localhost:9999"

	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	addrs := []string{"http://localhost:8001", "http://localhost:8002", "http://localhost:8003"}

	cache := zcache.NewGroup("scores", 2<<10, zcache.GetterFunc(getSourceData))

	if api {
		// 开启主服务
		startAPIServer(apiAddr, []string(addrs), cache)
	} else {
		// 开启缓存服务节点
		startCacheServer(addrMap[port], []string(addrs), cache)
	}

}
