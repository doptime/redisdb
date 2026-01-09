package redisdb

import (
	cmap "github.com/orcaman/concurrent-map/v2"
)

type IHttpSetKey interface {
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

var HttpSetKeyMap cmap.ConcurrentMap[string, IHttpSetKey] = cmap.New[IHttpSetKey]()

type HttpSetKey[k comparable, v any] SetKey[k, v]

func (ctx *HttpSetKey[k, v]) DeserializeValue(msgpack []byte) (rets interface{}, err error) {
	return ctx.DeserializeToValue(msgpack)
}
func (ctx *HttpSetKey[k, v]) DeserializeValues(msgpacks []string) (rets []interface{}, err error) {
	return ctx.DeserializeToInterfaceSlice(msgpacks)
}

func GetHttpSetKey(keyScope string, rdsName string) (IHttpSetKey, bool) {
	return HttpSetKeyMap.Get(keyScope + ":" + rdsName)
}
