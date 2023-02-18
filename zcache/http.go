// ZCache.zcache.http.go
package zcache

import (
	"ZCache/consistenthash"
	pb "ZCache/proto"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
)

const (
	defaultBasePath     = "/_zcache/"
	defaultVirtualCount = 50
)

type CacheServerPool struct {
	baseServerAddr  string
	basePath        string
	mu              sync.Mutex
	consistentHash  *consistenthash.ConsistentHash // 一致性哈希实例
	nodeHttpGetters map[string]*httpGetter         // 映射缓存节点和节点的请求方法
}

/* 实例化一个缓存服务器连接池 */
func NewCacheServerPool(addr string) *CacheServerPool {
	return &CacheServerPool{
		baseServerAddr: addr,
		basePath:       defaultBasePath,
	}
}

func (p *CacheServerPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.baseServerAddr, fmt.Sprintf(format, v...))
}

/* 在主节点实例化一致性哈希算法，注册缓存节点 */
func (p *CacheServerPool) Set(nodes ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.consistentHash = consistenthash.New(defaultVirtualCount, nil)
	p.consistentHash.Add(nodes...)
	p.nodeHttpGetters = make(map[string]*httpGetter, len(nodes))
	for _, node := range nodes {
		p.nodeHttpGetters[node] = &httpGetter{baseURL: node + p.basePath}
	}
}

/* 封装了一致性哈希的Get()方法，根据cacheKey选择节点，返回对应节点的HTTP客户端 */
func (p *CacheServerPool) SelectNode(key string) (NodeGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if node := p.consistentHash.Get(key); node != "" && node != p.baseServerAddr {
		p.Log("select node: %s", node)
		return p.nodeHttpGetters[node], true
	}
	return nil, false
}

var _ NodeSelector = (*CacheServerPool)(nil)

/* 缓存节点的服务入口，解析URL，根据缓存组名和缓存键获取数据 */
func (p *CacheServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 约定访问路径为 /<basePath>/<groupName>/<cacheKey>
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		// 验证URL是否以 /_zcache/ 开头
		panic("CacheServerPool serving unexpected path: " + r.URL.Path)
	}
	// p.Log("%s %s", r.Method, r.URL.Path)

	// 将URL切分成 [groupName cacheKey]
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	cacheKey := parts[1]
	group := GetGroup(groupName)

	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	data, err := group.Get(cacheKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: data.Bytes()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}
