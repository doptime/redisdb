package redisdb

import (
	"fmt"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type IHttpStringKey interface {
	GetKeyType() KeyType
	GetUseModer() bool
	ValidDataKey() error
	DeserializeValue(msgpack []byte) (rets interface{}, err error)
	DeserializeValues(msgpacks []string) (rets []interface{}, err error)
	TimestampFiller(in interface{}) (err error)

	Set(field string, val interface{}, expiration time.Duration) (int64, error)
	Get(field string) (interface{}, error)
}

var HttpStringKeyMap cmap.ConcurrentMap[string, IHttpStringKey] = cmap.New[IHttpStringKey]()

type HttpStringKey[k comparable, v any] StringKey[k, v]

func (ctx *HttpStringKey[k, v]) GetKeyType() KeyType {
	return (*HashKey[k, v])(ctx).GetKeyType()
}
func (ctx *HttpStringKey[k, v]) GetUseModer() bool {
	return (*HashKey[k, v])(ctx).GetUseModer()
}
func (ctx *HttpStringKey[k, v]) ValidDataKey() error {
	return (*HashKey[k, v])(ctx).ValidDataKey()
}
func (ctx *HttpStringKey[k, v]) TimestampFiller(in interface{}) (err error) {
	return (*HashKey[k, v])(ctx).TimestampFiller(in)
}

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
func (ctx *HttpStringKey[k, v]) Set(field string, val interface{}, expiration time.Duration) (err error) {
	var key k
	key, err = ctx.toKey([]byte(field))
	if err != nil {
		return err
	}
	var _v v
	_v, ok := val.(v)
	if !ok {
		return fmt.Errorf("value type assertion failed")
	}
	return (*StringKey[k, v])(ctx).Set(key, _v, expiration)
}
func (ctx *HttpStringKey[k, v]) Get(field string) (val interface{}, err error) {
	var key k
	key, err = ctx.toKey([]byte(field))
	if err != nil {
		return nil, err
	}
	return (*StringKey[k, v])(ctx).Get(key)
}
