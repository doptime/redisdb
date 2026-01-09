package redisdb

import (
	cmap "github.com/orcaman/concurrent-map/v2"
)

type HttpInterfaceList interface {
	GetKeyType() KeyType
	GetUseModer() bool
	ValidDataKey() error
	DeserializeValue(msgpack []byte) (rets interface{}, err error)
	DeserializeValues(msgpacks []string) (rets []interface{}, err error)
	TimestampFiller(in interface{}) (err error)
}

var HttpListKeyMap cmap.ConcurrentMap[string, HttpInterfaceList] = cmap.New[HttpInterfaceList]()

type HttpListKey[v any] ListKey[v]

func (ctx *HttpListKey[v]) DeserializeValue(msgpack []byte) (rets interface{}, err error) {
	return ctx.DeserializeToValue(msgpack)
}
func (ctx *HttpListKey[v]) DeserializeValues(msgpacks []string) (rets []interface{}, err error) {
	return ctx.DeserializeToInterfaceSlice(msgpacks)
}

func GetHttpListKey(keyScope string, rdsName string) (HttpInterfaceList, bool) {
	return HttpListKeyMap.Get(keyScope + ":" + rdsName)
}
