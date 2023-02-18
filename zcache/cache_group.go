// ZCache.zcache.cache_group.go
package zcache

import (
	pb "ZCache/proto"
	singleflight "ZCache/singlefilght"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"

	"google.golang.org/protobuf/proto"
)

/* 获取源数据方法的接口，暴露给用户实现 */
type SourceDataGetter interface {
	SearchData(key string) ([]byte, error)
}

/* 函数类型实现接口 */
type GetterFunc func(key string) ([]byte, error)

/* 让函数具备调用自己的能力来实现接口 */
func (f GetterFunc) SearchData(key string) ([]byte, error) {
	return f(key)
}

// ZCache.zcache.cache_group.go
/* 缓存组 */
type Group struct {
	groupName    string
	source       SourceDataGetter
	mainCache    cache
	nodeSelector NodeSelector

	// 添加缓存击穿组件
	loader *singleflight.Group
}

var (
	mu          sync.RWMutex              // 用来支持缓存组的并发安全
	cacheGroups = make(map[string]*Group) // 缓存组 {缓存组1: cache对象1, 缓存组2: cache对象2, ...}
)

func NewGroup(name string, cacheMaxBytes int64, getter SourceDataGetter) *Group {
	if getter == nil {
		panic("SourceDataGetter is a must")
	}
	mu.Lock()
	defer mu.Unlock()
	group := &Group{
		groupName: name,
		source:    getter,
		mainCache: cache{maxBytes: cacheMaxBytes},
		loader:    &singleflight.Group{},
	}
	cacheGroups[name] = group
	return group
}

/* 注册节点选择器 */
func (g *Group) RegisterNodeSelector(nodeSelector NodeSelector) {
	if g.nodeSelector != nil {
		panic("RegisterNodeSelector called more than once")
	}
	g.nodeSelector = nodeSelector
}

/* 获取缓存组 */
func GetGroup(groupName string) *Group {
	// 用只读锁因为不涉及任何冲突变量的写操作
	mu.RLock()
	g := cacheGroups[groupName]
	mu.RUnlock()
	return g
}

/* 根据缓存键返回数据，如果缓存未命中就调用source获取源数据 */
func (g *Group) Get(cacheKey string) (ByteView, error) {
	if cacheKey == "" {
		return ByteView{}, fmt.Errorf("cache key is required")
	}

	if v, ok := g.mainCache.get(cacheKey); ok {
		log.Println("hit cache")
		return v, nil
	}

	return g.load(cacheKey)
}

/* 获取源数据 */
func (g *Group) load(key string) (value ByteView, err error) {
	// 先调用缓存击穿
	view, err := g.loader.Do(key, func() (interface{}, error) {
		// 分布式缓存，根据缓存键选择缓存服务器，发起请求
		if g.nodeSelector != nil {
			if node, ok := g.nodeSelector.SelectNode(key); ok {
				if byteView, err := g.getFromRemote(node, key); err == nil {
					return byteView, nil
				}
				log.Println("[ZCache] Failed to get from node", err)
			}
		}

		// 单机缓存，从本地获取
		return g.getFromLocal(key)
	})

	if err != nil {
		return
	}
	return view.(ByteView), nil

}

/* 从本地获取源数据并新增到缓存中 */
func (g *Group) getFromLocal(key string) (ByteView, error) {
	// 获取源数据
	sourceBytesData, err := g.source.SearchData(key)
	if err != nil {
		return ByteView{}, err
	}
	// 深拷贝一份，封装成ByteView，防止源数据被污染
	value := ByteView{b: deepCopy(sourceBytesData)}
	// 保存到缓存中
	g.populateCache(key, value)
	return value, nil
}

/* 添加缓存的方法 */
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

/* 从缓存服务器获取缓存 */
func (g *Group) getFromRemote(node NodeGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.groupName,
		Key:   key,
	}
	res := &pb.Response{}
	err := node.Request(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

/* 一个缓存服务器对应一个调用方法 */
type httpGetter struct {
	baseURL string
}

/* 请求缓存服务器 */
func (h *httpGetter) Request(in *pb.Request, out *pb.Response) error {
	// url: http://example.com/_zcache/<groupName>/<cacheKey>
	url := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body %v", err)
	}

	return nil

}

var _ NodeGetter = (*httpGetter)(nil)
