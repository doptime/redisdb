package redisdb

import (
	"fmt"

	cmap "github.com/orcaman/concurrent-map/v2"
)

// IHttpSetKey 定义 HTTP 层对 Set 的操作接口
type IHttpSetKey interface {
	// --- 基础元数据 ---
	GetKeyType() KeyType
	GetUseModer() bool
	GetValue() interface{}
	ValidDataKey() error
	TimestampFiller(in interface{}) (err error)

	// --- 上下文注入 (核心) ---
	WithContext(key string, ds string) IHttpSetKey

	// --- 数据操作 ---
	// 读操作: 返回 interface{} (底层是 v 或 []v)
	SCard() (int64, error)
	SIsMember(member interface{}) (bool, error)
	SMembers() ([]interface{}, error)
	SScan(cursor uint64, match string, count int64) (values []interface{}, retCursor uint64, err error)

	// 写操作: 接收 interface{} (需断言为 v)
	SAdd(members ...interface{}) error
	SRem(members ...interface{}) error
}

// 全局注册表
var HttpSetKeyMap cmap.ConcurrentMap[string, IHttpSetKey] = cmap.New[IHttpSetKey]()

// HttpSetKey 是 SetKey 的 Wrapper
type HttpSetKey[k comparable, v any] SetKey[k, v]

// --- 内部辅助 ---

func (ctx *HttpSetKey[k, v]) native() *SetKey[k, v] {
	return (*SetKey[k, v])(ctx)
}

// --- 接口实现 ---

func (ctx *HttpSetKey[k, v]) GetKeyType() KeyType {
	return ctx.native().GetKeyType()
}
func (ctx *HttpSetKey[k, v]) GetUseModer() bool {
	return ctx.native().GetUseModer()
}
func (ctx *HttpSetKey[k, v]) GetValue() interface{} {
	var _value v
	return _value
}
func (ctx *HttpSetKey[k, v]) ValidDataKey() error {
	return ctx.native().ValidDataKey()
}
func (ctx *HttpSetKey[k, v]) TimestampFiller(in interface{}) (err error) {
	return ctx.native().TimestampFiller(in)
}

// WithContext 实现：克隆并注入上下文
func (ctx *HttpSetKey[k, v]) WithContext(key string, ds string) IHttpSetKey {
	// 1. 获取底层 RedisKey 的副本
	newObj := ctx.native().Duplicate(key, ds)

	// 2. 重新包装
	newSet := SetKey[k, v]{RedisKey: newObj}
	newCtx := HttpSetKey[k, v](newSet)

	return &newCtx
}

// --- 数据操作实现 ---

func (ctx *HttpSetKey[k, v]) SAdd(members ...interface{}) error {
	var vMembers []v
	for _, m := range members {
		// 类型断言：确保传入的是 v 类型
		if vm, ok := m.(v); ok {
			vMembers = append(vMembers, vm)
		} else {
			return fmt.Errorf("SAdd type mismatch: expected %T, got %T", *new(v), m)
		}
	}
	return ctx.native().SAdd(vMembers...)
}

func (ctx *HttpSetKey[k, v]) SRem(members ...interface{}) error {
	var vMembers []v
	for _, m := range members {
		if vm, ok := m.(v); ok {
			vMembers = append(vMembers, vm)
		} else {
			return fmt.Errorf("SRem type mismatch: expected %T, got %T", *new(v), m)
		}
	}
	return ctx.native().SRem(vMembers...)
}

func (ctx *HttpSetKey[k, v]) SCard() (int64, error) {
	return ctx.native().SCard()
}

func (ctx *HttpSetKey[k, v]) SIsMember(member interface{}) (bool, error) {
	if vm, ok := member.(v); ok {
		return ctx.native().SIsMember(vm)
	}
	return false, fmt.Errorf("SIsMember type mismatch: expected %T", *new(v))
}

func (ctx *HttpSetKey[k, v]) SMembers() ([]interface{}, error) {
	values, err := ctx.native().SMembers()
	if err != nil {
		return nil, err
	}
	// 转换 []v -> []interface{}
	rets := make([]interface{}, len(values))
	for i, val := range values {
		rets[i] = val
	}
	return rets, nil
}

func (ctx *HttpSetKey[k, v]) SScan(cursor uint64, match string, count int64) ([]interface{}, uint64, error) {
	values, newCursor, err := ctx.native().SScan(cursor, match, count)
	if err != nil {
		return nil, 0, err
	}
	// 转换 []v -> []interface{}
	rets := make([]interface{}, len(values))
	for i, val := range values {
		rets[i] = val
	}
	return rets, newCursor, nil
}

// 工厂方法
func GetHttpSetKey(Key string, rdsName string) (IHttpSetKey, error) {
	_keyscope := KeyScope(Key)
	ikey, ok := HttpSetKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found for: %s", _keyscope)
	}
	// 核心：调用 WithContext
	return ikey.WithContext(Key, rdsName), nil
}
