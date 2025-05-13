// 实现结点选择 (peerpick)和结点获得 (peerget)
package geecache

import pb "geecache/geecachepb"

// 结点选择
type Peerpick interface {
	PickPeer(key string)(peer PeerGetter, ok bool)
}

// 结点获取
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}