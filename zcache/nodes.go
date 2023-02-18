// ZCache.zcache.nodes.go
package zcache

import pb "ZCache/proto"

/* 根据cacheKey选择对应节点 */
type NodeSelector interface {
	SelectNode(cacheKey string) (getter NodeGetter, ok bool)
}

/* 调用远程节点的接口 */
type NodeGetter interface {
	Request(parameter *pb.Request, out *pb.Response) error
}
