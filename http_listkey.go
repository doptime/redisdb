package redisdb

import (
	"fmt"

	cmap "github.com/orcaman/concurrent-map/v2"
)

// IHttpListKey 定义 HTTP 层对 List 的操作接口
type IHttpListKey interface {
	// --- 基础元数据 ---
	GetKeyType() KeyType
	GetUseModer() bool
	ValidDataKey() error
	DeserializeValue(msgpack []byte) (rets interface{}, err error)
	DeserializeValues(msgpacks []string) (rets []interface{}, err error)
	TimestampFiller(in interface{}) (err error)

	// --- 上下文注入 (核心) ---
	WithContext(key string, ds string) IHttpListKey

	// --- 数据操作 ---
	// 读操作：返回 interface{} (底层是 v 或 []v)
	LRange(start int64, stop int64) (rets []interface{}, err error)
	LIndex(index int64) (ret interface{}, err error)
	LPop() (ret interface{}, err error)
	RPop() (ret interface{}, err error)
	LLen() (ret int64, err error)

	// 写操作：接收 interface{} (需断言为 v)
	LPush(vals ...interface{}) (err error)
	RPush(vals ...interface{}) (err error)
	LRem(count int64, val interface{}) (err error)
	LTrim(start int64, stop int64) (err error)
	LSet(index int64, val interface{}) (err error)
	RPushX(val interface{}) (err error)
	LPushX(val interface{}) (err error)
}

// 全局注册表
var HttpListKeyMap cmap.ConcurrentMap[string, IHttpListKey] = cmap.New[IHttpListKey]()

// HttpListKey 是 ListKey 的 Wrapper
type HttpListKey[v any] ListKey[v]

// --- 内部辅助 ---

func (ctx *HttpListKey[v]) native() *ListKey[v] {
	return (*ListKey[v])(ctx)
}

// --- 接口实现 ---

func (ctx *HttpListKey[v]) GetKeyType() KeyType {
	return ctx.native().GetKeyType()
}
func (ctx *HttpListKey[v]) GetUseModer() bool {
	return ctx.native().GetUseModer()
}
func (ctx *HttpListKey[v]) ValidDataKey() error {
	return ctx.native().ValidDataKey()
}
func (ctx *HttpListKey[v]) TimestampFiller(in interface{}) (err error) {
	return ctx.native().TimestampFiller(in)
}

func (ctx *HttpListKey[v]) DeserializeValue(msgpack []byte) (rets interface{}, err error) {
	return ctx.native().DeserializeToValue(msgpack)
}
func (ctx *HttpListKey[v]) DeserializeValues(msgpacks []string) (rets []interface{}, err error) {
	return ctx.native().DeserializeToInterfaceSlice(msgpacks)
}

// WithContext 实现：克隆并注入上下文
func (ctx *HttpListKey[v]) WithContext(key string, ds string) IHttpListKey {
	// 1. 获取底层 RedisKey 的副本 (ListKey Embed 了 RedisKey[string, v])
	// 注意 ListKey 的 Key 类型固定为 string
	newObj := ctx.native().Duplicate(key, ds)

	// 2. 重新包装
	newList := ListKey[v]{RedisKey: newObj}
	newCtx := HttpListKey[v](newList)

	return &newCtx
}

// --- 数据操作实现 ---

func (ctx *HttpListKey[v]) LRange(start int64, stop int64) (rets []interface{}, err error) {
	var values []v
	values, err = ctx.native().LRange(start, stop)
	if err != nil {
		return nil, err
	}
	// 转换 []v -> []interface{}
	rets = make([]interface{}, len(values))
	for i, val := range values {
		rets[i] = val
	}
	return rets, nil
}

func (ctx *HttpListKey[v]) LIndex(index int64) (ret interface{}, err error) {
	return ctx.native().LIndex(index)
}

func (ctx *HttpListKey[v]) LPop() (ret interface{}, err error) {
	return ctx.native().LPop()
}

func (ctx *HttpListKey[v]) RPop() (ret interface{}, err error) {
	return ctx.native().RPop()
}

func (ctx *HttpListKey[v]) LPush(vals ...interface{}) (err error) {
	var vvals []v
	for _, val := range vals {
		// 类型断言：确保传入的是 v 类型
		vval, ok := val.(v)
		if !ok {
			return fmt.Errorf("LPush type mismatch: expected %T, got %T", *new(v), val)
		}
		vvals = append(vvals, vval)
	}
	return ctx.native().LPush(vvals...)
}

func (ctx *HttpListKey[v]) RPush(vals ...interface{}) (err error) {
	var vvals []v
	for _, val := range vals {
		vval, ok := val.(v)
		if !ok {
			return fmt.Errorf("RPush type mismatch: expected %T, got %T", *new(v), val)
		}
		vvals = append(vvals, vval)
	}
	return ctx.native().RPush(vvals...)
}

func (ctx *HttpListKey[v]) LRem(count int64, val interface{}) (err error) {
	vval, ok := val.(v)
	if !ok {
		return fmt.Errorf("LRem type mismatch: expected %T, got %T", *new(v), val)
	}
	return ctx.native().LRem(count, vval)
}

func (ctx *HttpListKey[v]) LTrim(start int64, stop int64) (err error) {
	return ctx.native().LTrim(start, stop)
}

func (ctx *HttpListKey[v]) LLen() (ret int64, err error) {
	return ctx.native().LLen()
}

func (ctx *HttpListKey[v]) LSet(index int64, val interface{}) (err error) {
	vval, ok := val.(v)
	if !ok {
		return fmt.Errorf("LSet type mismatch: expected %T, got %T", *new(v), val)
	}
	return ctx.native().LSet(index, vval)
}

func (ctx *HttpListKey[v]) RPushX(val interface{}) (err error) {
	vval, ok := val.(v)
	if !ok {
		return fmt.Errorf("RPushX type mismatch: expected %T, got %T", *new(v), val)
	}
	return ctx.native().RPushX(vval)
}

func (ctx *HttpListKey[v]) LPushX(val interface{}) (err error) {
	vval, ok := val.(v)
	if !ok {
		return fmt.Errorf("LPushX type mismatch: expected %T, got %T", *new(v), val)
	}
	return ctx.native().LPushX(vval)
}

// 工厂方法
func GetHttpListKey(Key string, rdsName string) (IHttpListKey, error) {
	_keyscope := KeyScope(Key)
	ikey, ok := HttpListKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found for: %s", _keyscope)
	}
	// 核心：调用 WithContext
	return ikey.WithContext(Key, rdsName), nil
}
