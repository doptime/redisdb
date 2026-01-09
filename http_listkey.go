package redisdb

import (
	"fmt"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type IHttpListKey interface {
	GetKeyType() KeyType
	GetUseModer() bool
	ValidDataKey() error
	DeserializeValue(msgpack []byte) (rets interface{}, err error)
	DeserializeValues(msgpacks []string) (rets []interface{}, err error)
	TimestampFiller(in interface{}) (err error)

	LRange(start int64, stop int64) (rets []interface{}, err error)
	LIndex(index int64) (ret interface{}, err error)
	LPop() (ret interface{}, err error)
	RPop() (ret interface{}, err error)
	LPush(vals ...interface{}) (err error)
	RPush(vals ...interface{}) (err error)
	LRem(count int64, val interface{}) (err error)
	LTrim(start int64, stop int64) (err error)
	LLen() (ret int64, err error)
	LSet(index int64, val interface{}) (err error)
	RPushX(val interface{}) (err error)
	LPushX(val interface{}) (err error)
}

var HttpListKeyMap cmap.ConcurrentMap[string, IHttpListKey] = cmap.New[IHttpListKey]()

type HttpListKey[v any] ListKey[v]

func (ctx *HttpListKey[v]) GetKeyType() KeyType {
	return (*ListKey[v])(ctx).GetKeyType()
}
func (ctx *HttpListKey[v]) GetUseModer() bool {
	return (*ListKey[v])(ctx).GetUseModer()
}
func (ctx *HttpListKey[v]) ValidDataKey() error {
	return (*ListKey[v])(ctx).ValidDataKey()
}
func (ctx *HttpListKey[v]) TimestampFiller(in interface{}) (err error) {
	return (*ListKey[v])(ctx).TimestampFiller(in)
}

func (ctx *HttpListKey[v]) DeserializeValue(msgpack []byte) (rets interface{}, err error) {
	return ctx.DeserializeToValue(msgpack)
}
func (ctx *HttpListKey[v]) DeserializeValues(msgpacks []string) (rets []interface{}, err error) {
	return ctx.DeserializeToInterfaceSlice(msgpacks)
}

func (ctx *HttpListKey[v]) LRange(start int64, stop int64) (rets []interface{}, err error) {
	var values []v
	values, err = (*ListKey[v])(ctx).LRange(start, stop)
	if err != nil {
		return nil, err
	}
	for _, val := range values {
		rets = append(rets, val)
	}
	return rets, nil
}
func (ctx *HttpListKey[v]) LIndex(index int64) (ret interface{}, err error) {
	var value v
	value, err = (*ListKey[v])(ctx).LIndex(index)
	if err != nil {
		return nil, err
	}
	return value, nil
}
func (ctx *HttpListKey[v]) LPop() (ret interface{}, err error) {
	var value v
	value, err = (*ListKey[v])(ctx).LPop()
	if err != nil {
		return nil, err
	}
	return value, nil
}
func (ctx *HttpListKey[v]) RPop() (ret interface{}, err error) {
	var value v
	value, err = (*ListKey[v])(ctx).RPop()
	if err != nil {
		return nil, err
	}
	return value, nil
}
func (ctx *HttpListKey[v]) LPush(vals ...interface{}) (err error) {
	var vvals []v
	for _, val := range vals {
		vval, ok := val.(v)
		if !ok {
			return fmt.Errorf("type assertion failed in LPush")
		}
		vvals = append(vvals, vval)
	}
	return (*ListKey[v])(ctx).LPush(vvals...)
}
func (ctx *HttpListKey[v]) RPush(vals ...interface{}) (err error) {
	var vvals []v
	for _, val := range vals {
		vval, ok := val.(v)
		if !ok {
			return fmt.Errorf("type assertion failed in RPush")
		}
		vvals = append(vvals, vval)
	}
	return (*ListKey[v])(ctx).RPush(vvals...)
}
func (ctx *HttpListKey[v]) LRem(count int64, val interface{}) (err error) {
	vval, ok := val.(v)
	if !ok {
		return fmt.Errorf("type assertion failed in LRem")
	}
	return (*ListKey[v])(ctx).LRem(count, vval)
}
func (ctx *HttpListKey[v]) LTrim(start int64, stop int64) (err error) {
	return (*ListKey[v])(ctx).LTrim(start, stop)
}
func (ctx *HttpListKey[v]) LLen() (ret int64, err error) {
	return (*ListKey[v])(ctx).LLen()
}
func (ctx *HttpListKey[v]) LSet(index int64, val interface{}) (err error) {
	vval, ok := val.(v)
	if !ok {
		return fmt.Errorf("type assertion failed in LSet")
	}
	return (*ListKey[v])(ctx).LSet(index, vval)
}
func (ctx *HttpListKey[v]) RPushX(val interface{}) (err error) {
	vval, ok := val.(v)
	if !ok {
		return fmt.Errorf("type assertion failed in RPushX")
	}
	return (*ListKey[v])(ctx).RPushX(vval)
}
func (ctx *HttpListKey[v]) LPushX(val interface{}) (err error) {
	vval, ok := val.(v)
	if !ok {
		return fmt.Errorf("type assertion failed in LPushX")
	}
	return (*ListKey[v])(ctx).LPushX(vval)
}

func GetHttpListKey(Key string, rdsName string) (IHttpListKey, error) {
	_keyscope := KeyScope(Key)
	ikey, ok := HttpListKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found")
	}
	return ikey, nil
}
