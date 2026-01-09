package redisdb

import (
	"fmt"

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

func GetHttpStringKey(Key string, rdsName string) (IHttpStringKey, error) {
	_keyscope := KeyScope(Key)
	ikey, ok := HttpStringKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found")
	}
	return ikey, nil
}
