package redisdb

import (
	"fmt"

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

func GetHttpListKey(Key string, rdsName string) (HttpInterfaceList, error) {
	_keyscope := KeyScope(Key)
	ikey, ok := HttpListKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found")
	}
	return ikey, nil
}
