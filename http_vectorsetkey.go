package redisdb

import (
	"fmt"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type IHttpVectorSetKey interface {
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

var HttpVectorSetKeyMap cmap.ConcurrentMap[string, IHttpVectorSetKey] = cmap.New[IHttpVectorSetKey]()

type HttpVectorSetKey[k comparable, v any] VectorSetKey[k, v]

func (ctx *HttpVectorSetKey[k, v]) DeserializeValue(msgpack []byte) (rets interface{}, err error) {
	return ctx.DeserializeToValue(msgpack)
}
func (ctx *HttpVectorSetKey[k, v]) DeserializeValues(msgpacks []string) (rets []interface{}, err error) {
	return ctx.DeserializeToInterfaceSlice(msgpacks)
}

func GetHttpVectorSetKey(Key string, rdsName string) (IHttpVectorSetKey, error) {
	_keyscope := KeyScope(Key)
	ikey, ok := HttpVectorSetKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found")
	}
	return ikey, nil
}
