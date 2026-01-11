package redisdb

import (
	"fmt"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/redis/go-redis/v9"
)

// IHttpStreamKey 定义 HTTP 层对 Stream 的操作需求
type IHttpStreamKey interface {
	// --- 基础元数据 ---
	GetKeyType() KeyType
	GetUseModer() bool
	GetValue() interface{}
	ValidDataKey() error
	TimestampFiller(in interface{}) (err error)

	// --- 上下文注入 (核心) ---
	WithContext(key string, ds string) IHttpStreamKey

	// --- 数据操作 ---
	XLen() (int64, error)
	// XAdd 参数简化：HTTP 层传 ID 和 Values (通常是 map 或 slice)
	XAdd(id string, values interface{}) (string, error)
	XDel(ids ...string) error

	// Range 类：统一返回 interface{} (实际是 []redis.XMessage)
	XRange(start, stop string, count int64) (interface{}, error)
	XRevRange(start, stop string, count int64) (interface{}, error)

	// Read 类
	XRead(streams []string, count int64, block time.Duration) (interface{}, error)
}

// 全局注册表
var HttpStreamKeyMap cmap.ConcurrentMap[string, IHttpStreamKey] = cmap.New[IHttpStreamKey]()

// HttpStreamKey 是 StreamKey 的 Wrapper
type HttpStreamKey[k comparable, v any] StreamKey[k, v]

// --- 内部辅助 ---

func (ctx *HttpStreamKey[k, v]) native() *StreamKey[k, v] {
	return (*StreamKey[k, v])(ctx)
}

// --- 接口实现 ---

func (ctx *HttpStreamKey[k, v]) GetKeyType() KeyType {
	return ctx.native().GetKeyType()
}
func (ctx *HttpStreamKey[k, v]) GetUseModer() bool {
	return ctx.native().GetUseModer()
}
func (ctx *HttpStreamKey[k, v]) GetValue() interface{} {
	var _value v
	return _value
}
func (ctx *HttpStreamKey[k, v]) ValidDataKey() error {
	return ctx.native().ValidDataKey()
}
func (ctx *HttpStreamKey[k, v]) TimestampFiller(in interface{}) (err error) {
	return ctx.native().TimestampFiller(in)
}

// WithContext: 克隆对象并注入上下文
func (ctx *HttpStreamKey[k, v]) WithContext(key string, ds string) IHttpStreamKey {
	// 1. 获取原始对象的副本 (浅拷贝 RedisKey)
	newObj := ctx.native().Duplicate(key, ds)
	// 2. 包装并返回
	newCtx := HttpStreamKey[k, v]{RedisKey: newObj}
	return &newCtx
}

// --- 数据操作实现 ---

func (ctx *HttpStreamKey[k, v]) XLen() (int64, error) {
	return ctx.native().XLen()
}

func (ctx *HttpStreamKey[k, v]) XAdd(id string, values interface{}) (string, error) {
	// 构造 XAddArgs
	args := &redis.XAddArgs{
		ID:     id,
		Values: values,
		// Stream Key 在 key_stream.go 的 XAdd 内部会自动填充为 ctx.Key
	}
	return ctx.native().XAdd(args)
}

func (ctx *HttpStreamKey[k, v]) XDel(ids ...string) error {
	_, err := ctx.native().XDel(ids...)
	return err
}

func (ctx *HttpStreamKey[k, v]) XRange(start, stop string, count int64) (interface{}, error) {
	// 根据 count 判断调用哪个底层方法
	if count > 0 {
		return ctx.native().XRangeN(start, stop, count)
	}
	return ctx.native().XRange(start, stop)
}

func (ctx *HttpStreamKey[k, v]) XRevRange(start, stop string, count int64) (interface{}, error) {
	if count > 0 {
		return ctx.native().XRevRangeN(start, stop, count)
	}
	return ctx.native().XRevRange(start, stop)
}

func (ctx *HttpStreamKey[k, v]) XRead(streams []string, count int64, block time.Duration) (interface{}, error) {
	args := &redis.XReadArgs{
		Streams: streams,
		Count:   count,
		Block:   block,
	}
	return ctx.native().XRead(args)
}

// 工厂方法
func GetHttpStreamKey(Key string, rdsName string) (IHttpStreamKey, error) {
	_keyscope := KeyScope(Key)
	ikey, ok := HttpStreamKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found for: %s", _keyscope)
	}
	// 核心：调用 WithContext
	return ikey.WithContext(Key, rdsName), nil
}
