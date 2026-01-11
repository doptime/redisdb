package redisdb

import (
	"fmt"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type IHttpStringKey interface {
	GetKeyType() KeyType
	GetUseModer() bool
	GetValue() interface{}
	ValidDataKey() error
	TimestampFiller(in interface{}) (err error)

	// Context 注入 (核心：用于多租户/Key变换)
	WithContext(key string, RedisDataSource string) IHttpStringKey

	Set(field string, val interface{}, expiration time.Duration) error
	Get(field string) (interface{}, error)
}

var HttpStringKeyMap cmap.ConcurrentMap[string, IHttpStringKey] = cmap.New[IHttpStringKey]()

type HttpStringKey[k comparable, v any] StringKey[k, v]

// --- 内部辅助 ---

func (ctx *HttpStringKey[k, v]) native() *StringKey[k, v] {
	return (*StringKey[k, v])(ctx)
}

// --- 接口实现 ---

func (ctx *HttpStringKey[k, v]) GetKeyType() KeyType {
	return ctx.native().GetKeyType()
}
func (ctx *HttpStringKey[k, v]) GetUseModer() bool {
	return ctx.native().GetUseModer()
}
func (ctx *HttpStringKey[k, v]) GetValue() interface{} {
	var _value v
	return _value
}
func (ctx *HttpStringKey[k, v]) ValidDataKey() error {
	return ctx.native().ValidDataKey()
}
func (ctx *HttpStringKey[k, v]) TimestampFiller(in interface{}) (err error) {
	return ctx.native().TimestampFiller(in)
}

// WithContext 实现：克隆自己，修改 Key 和 DS，返回接口
func (ctx *HttpStringKey[k, v]) WithContext(key string, RedisDataSource string) IHttpStringKey {
	// 1. 获取原始对象的副本 (RedisKey)
	newObj := ctx.native().Duplicate(key, RedisDataSource)
	// 2. 包装回 HttpStringKey
	// 注意：StringKey 结构体中嵌入了 RedisKey，所以可以直接初始化
	newCtx := HttpStringKey[k, v]{RedisKey: newObj}
	return &newCtx
}

// --- 数据操作实现 ---

func (ctx *HttpStringKey[k, v]) Set(field string, val interface{}, expiration time.Duration) (err error) {
	skey := ctx.native()
	// 1. 转换 Key (string -> k)
	var key k
	key, err = skey.toKey([]byte(field))
	if err != nil {
		return err
	}

	// 2. 转换 Value (interface{} -> v)
	// 这里依赖泛型 v 的具体类型，如果类型不匹配会报错
	_v, ok := val.(v)
	if !ok {
		return fmt.Errorf("value type assertion failed: expected %T, got %T", *new(v), val)
	}

	// 3. 调用底层 Set
	return skey.Set(key, _v, expiration)
}

func (ctx *HttpStringKey[k, v]) Get(field string) (val interface{}, err error) {
	skey := ctx.native()
	// 1. 转换 Key (string -> k)
	var key k
	key, err = skey.toKey([]byte(field))
	if err != nil {
		return nil, err
	}

	// 2. 调用底层 Get (返回 v)
	return skey.Get(key)
}

// 工厂方法
func GetHttpStringKey(Key string, rdsName string) (IHttpStringKey, error) {
	_keyscope := KeyScope(Key)
	ikey, ok := HttpStringKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found for scope: %s", _keyscope)
	}
	// 核心修改：必须调用 WithContext 注入具体的 Key 和 DataSource
	return ikey.WithContext(Key, rdsName), nil
}
