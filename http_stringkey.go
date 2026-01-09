package redisdb

import (
	cmap "github.com/orcaman/concurrent-map/v2"
)

type IHttpStringKey interface {
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

var HttpStringKeyMap cmap.ConcurrentMap[string, IHttpStringKey] = cmap.New[IHttpStringKey]()

type HttpStringKey[k comparable, v any] StringKey[k, v]

func (ctx *HttpStringKey[k, v]) DeserializeValue(msgpack []byte) (rets interface{}, err error) {
	return ctx.DeserializeToValue(msgpack)
}
func (ctx *HttpStringKey[k, v]) DeserializeValues(msgpacks []string) (rets []interface{}, err error) {
	return ctx.DeserializeToInterfaceSlice(msgpacks)
}

func GetHttpStringKey(keyScope string, rdsName string) (IHttpStringKey, bool) {
	return HttpStringKeyMap.Get(keyScope + ":" + rdsName)
}
