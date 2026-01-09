package redisdb

import (
	"fmt"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type IHttpZSetKey interface {
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

var HttpZSetKeyMap cmap.ConcurrentMap[string, IHttpZSetKey] = cmap.New[IHttpZSetKey]()

type HttpZSetKey[k comparable, v any] ZSetKey[k, v]

func (ctx *HttpZSetKey[k, v]) DeserializeValue(msgpack []byte) (rets interface{}, err error) {
	return ctx.DeserializeToValue(msgpack)
}
func (ctx *HttpZSetKey[k, v]) DeserializeValues(msgpacks []string) (rets []interface{}, err error) {
	return ctx.DeserializeToInterfaceSlice(msgpacks)
}

func GetHttpZSetKey(Key string, rdsName string) (IHttpZSetKey, error) {
	_keyscope := KeyScope(Key)
	ikey, ok := HttpZSetKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found")
	}
	return ikey, nil
}
