package redisdb

import (
	cmap "github.com/orcaman/concurrent-map/v2"
)

type IHttpStreamKey interface {
	GetKeyType() KeyType
	GetUseModer() bool
	ValidDataKey() error
	DeserializeValue(msgpack []byte) (rets interface{}, err error)
	DeserializeValues(msgpacks []string) (rets []interface{}, err error)
	TimestampFiller(in interface{}) (err error)

	// HGet(field string) (interface{}, error)
	// HGetAll() (map[string]interface{}, error)
	// HSet(field string, val interface{}) (int64, error)
}

var HttpStreamKeyMap cmap.ConcurrentMap[string, IHttpStreamKey] = cmap.New[IHttpStreamKey]()

type HttpStreamKey[k comparable, v any] StreamKey[k, v]

func (ctx *HttpStreamKey[k, v]) DeserializeValue(msgpack []byte) (rets interface{}, err error) {
	return ctx.DeserializeToValue(msgpack)
}
func (ctx *HttpStreamKey[k, v]) DeserializeValues(msgpacks []string) (rets []interface{}, err error) {
	return ctx.DeserializeToInterfaceSlice(msgpacks)
}

func GetHttpStreamKey(keyScope string, rdsName string) (IHttpStreamKey, bool) {
	return HttpStreamKeyMap.Get(keyScope + ":" + rdsName)
}
