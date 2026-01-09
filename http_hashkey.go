package redisdb

import (
	"fmt"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type IHttpHashKey interface {
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

var HttpHashKeyMap cmap.ConcurrentMap[string, IHttpHashKey] = cmap.New[IHttpHashKey]()

type HttpHashKey[k comparable, v any] HashKey[k, v]

func (ctx *HttpHashKey[k, v]) DeserializeValue(msgpack []byte) (rets interface{}, err error) {
	return ctx.DeserializeToValue(msgpack)
}
func (ctx *HttpHashKey[k, v]) DeserializeValues(msgpacks []string) (rets []interface{}, err error) {
	return ctx.DeserializeToInterfaceSlice(msgpacks)
}
func GetHttpHashKey(Key string, rdsName string) (IHttpHashKey, error) {
	_keyscope := KeyScope(Key)
	ikey, ok := HttpHashKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found")
	}
	return ikey, nil
}
