package geeCache

type PeerPicker interface {
	PickPeer(key string) (peer PeerPicker, ok bool)
}

type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
