package redisdb

import (
	"fmt"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type IHttpHashKey interface {
	GetKeyType() KeyType
	GetUseModer() bool
	ValidDataKey() error
	GetValue() interface{}
	TimestampFiller(in interface{}) (err error)

	WithContext(key string, RedisDataSource string) IHttpHashKey

	HScanNoValues(cursor uint64, match string, count int64) (keys []string, cursorRet uint64, err error)
	HScan(cursor uint64, match string, count int64) (keys []string, values []interface{}, cursorRet uint64, err error)
	HGet(field string) (interface{}, error)
	HGetAll() (map[string]interface{}, error)
	HSet(field string, val interface{}) (int64, error)
	HMGET(fields ...interface{}) (vals []interface{}, err error)
	HKeys() (keys []string, err error)
	HVals() (vals []interface{}, err error)
	HExists(field string) (exists bool, err error)
	HRandField(count int) (keys []string, err error)
	HRandFieldWithValues(count int) (keyvalueMap map[string]interface{}, err error)
}

var HttpHashKeyMap cmap.ConcurrentMap[string, IHttpHashKey] = cmap.New[IHttpHashKey]()

type HttpHashKey[k comparable, v any] HashKey[k, v]

func (ctx *HttpHashKey[k, v]) GetKeyType() KeyType {
	return (*HashKey[k, v])(ctx).GetKeyType()
}
func (ctx *HttpHashKey[k, v]) GetUseModer() bool {
	return (*HashKey[k, v])(ctx).GetUseModer()
}
func (ctx *HttpHashKey[k, v]) GetValue() interface{} {
	var _value v
	return _value
}
func (ctx *HttpHashKey[k, v]) ValidDataKey() error {
	return (*HashKey[k, v])(ctx).ValidDataKey()
}
func (ctx *HttpHashKey[k, v]) TimestampFiller(in interface{}) (err error) {
	return (*HashKey[k, v])(ctx).TimestampFiller(in)
}

func (ctx *HttpHashKey[k, v]) native() *HashKey[k, v] {
	return (*HashKey[k, v])(ctx)
}

// WithContext 实现：克隆自己，修改 Key 和 DS，返回接口
func (ctx *HttpHashKey[k, v]) WithContext(key string, RedisDataSource string) IHttpHashKey {
	// 1. 获取原始对象的副本 (浅拷贝结构体)
	newObj := ctx.native().Duplicate(key, RedisDataSource)
	newCtx := HttpHashKey[k, v]{RedisKey: newObj}
	return &newCtx
}

func (ctx *HttpHashKey[k, v]) HScanNoValues(cursor uint64, match string, count int64) (keys []string, cursorRet uint64, err error) {
	var (
		keysRet []k
	)
	keysRet, cursorRet, err = (*HashKey[k, v])(ctx).HScanNoValues(cursor, match, count)
	if err != nil {
		return
	}
	for _, key := range keysRet {
		keys = append(keys, fmt.Sprintf("%v", key))
	}
	return
}

func (ctx *HttpHashKey[k, v]) HScan(cursor uint64, match string, count int64) (keys []string, values []interface{}, cursorRet uint64, err error) {
	var (
		keysRet   []k
		valuesRet []v
	)
	keysRet, valuesRet, cursorRet, err = (*HashKey[k, v])(ctx).HScan(cursor, match, count)
	if err != nil {
		return
	}
	for _, key := range keysRet {
		keys = append(keys, fmt.Sprintf("%v", key))
	}
	for _, val := range valuesRet {
		values = append(values, val)
	}
	return
}

func (ctx *HttpHashKey[k, v]) HGet(field string) (val interface{}, err error) {
	hkey := (*HashKey[k, v])(ctx)
	var key k
	key, err = hkey.toKey([]byte(field))
	if err != nil {
		return nil, err
	}
	return hkey.HGet(key)
}

func (ctx *HttpHashKey[k, v]) HGetAll() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	dataMap, err := (*HashKey[k, v])(ctx).HGetAll()
	if err != nil {
		return nil, err
	}
	for key, val := range dataMap {
		result[fmt.Sprintf("%v", key)] = val
	}
	return result, nil
}

func (ctx *HttpHashKey[k, v]) HSet(field string, val interface{}) (int64, error) {
	hkey := (*HashKey[k, v])(ctx)
	var key k
	key, err := hkey.toKey([]byte(field))
	if err != nil {
		return 0, err
	}
	return hkey.HSet(key, val)
}

func (ctx *HttpHashKey[k, v]) HMGET(fields ...interface{}) (vals []interface{}, err error) {
	hkey := (*HashKey[k, v])(ctx)
	var values []v
	values, err = hkey.HMGET(fields...)
	if err != nil {
		return nil, err
	}
	for _, val := range values {
		vals = append(vals, val)
	}
	return vals, nil
}

func (ctx *HttpHashKey[k, v]) HKeys() (keys []string, err error) {
	hkey := (*HashKey[k, v])(ctx)
	var keysRet []k
	keysRet, err = hkey.HKeys()
	if err != nil {
		return nil, err
	}
	for _, key := range keysRet {
		keys = append(keys, fmt.Sprintf("%v", key))
	}
	return
}
func (ctx *HttpHashKey[k, v]) HVals() (vals []interface{}, err error) {
	hkey := (*HashKey[k, v])(ctx)
	var valuesRet []v
	valuesRet, err = hkey.HVals()
	if err != nil {
		return nil, err
	}
	for _, val := range valuesRet {
		vals = append(vals, val)
	}
	return
}

func (ctx *HttpHashKey[k, v]) HExists(field string) (exists bool, err error) {
	hkey := (*HashKey[k, v])(ctx)
	var key k
	key, err = hkey.toKey([]byte(field))
	if err != nil {
		return false, err
	}
	return hkey.HExists(key)
}
func (ctx *HttpHashKey[k, v]) HRandField(count int) (keys []string, err error) {
	hkey := (*HashKey[k, v])(ctx)
	var keysRet []k
	keysRet, err = hkey.HRandField(count)
	if err != nil {
		return nil, err
	}
	for _, key := range keysRet {
		keys = append(keys, fmt.Sprintf("%v", key))
	}
	return
}

func (ctx *HttpHashKey[k, v]) HRandFieldWithValues(count int) (keyvalueMap map[string]interface{}, err error) {
	hkey := (*HashKey[k, v])(ctx)
	var keysRet []k
	var valuesRet []v
	keysRet, valuesRet, err = hkey.HRandFieldWithValues(count)
	if err != nil {
		return nil, err
	}
	keyvalueMap = make(map[string]interface{})
	if len(keysRet) != len(valuesRet) {
		return nil, fmt.Errorf("mismatched keys and values length")
	}

	for i, key := range keysRet {
		keyvalueMap[fmt.Sprintf("%v", key)] = valuesRet[i]
	}
	return keyvalueMap, nil
}

func GetHttpHashKey(Key string, rdsName string) (IHttpHashKey, error) {
	_keyscope := KeyScope(Key)
	ikey, ok := HttpHashKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found")
	}
	return ikey.WithContext(Key, rdsName), nil
}
