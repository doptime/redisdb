package redisdb

import (
	"fmt"

	"github.com/doptime/redisdb/utils"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// IHttpVectorSetKey 定义 HTTP 层对 VectorSet/RediSearch 的操作接口
type IHttpVectorSetKey interface {
	// --- 基础元数据 ---
	GetKeyType() KeyType
	GetUseModer() bool
	GetValue() interface{}
	ValidDataKey() error
	TimestampFiller(in interface{}) (err error)

	// --- 上下文注入 (核心) ---
	WithContext(key string, ds string) IHttpVectorSetKey

	// --- 索引管理 ---
	Create(args ...interface{}) error
	DropIndex(deleteDocs bool) error
	Info() (map[string]interface{}, error)

	// --- 别名管理 ---
	AliasAdd(alias string) error
	AliasUpdate(alias string) error
	AliasDel(alias string) error

	// --- 搜索与查询 ---
	// TagVals 获取 Tag 字段的所有去重值 (用于 Faceted Search)
	TagVals(fieldName string) ([]string, error)

	// Search 执行 FT.SEARCH
	// 返回 docs 为 interface{} (底层是 []v)，startHttp 会自动序列化它
	Search(query string, params ...interface{}) (count int64, docs interface{}, err error)
}

// 全局注册表
var HttpVectorSetKeyMap cmap.ConcurrentMap[string, IHttpVectorSetKey] = cmap.New[IHttpVectorSetKey]()

// HttpVectorSetKey 是 VectorSetKey 的 Wrapper
type HttpVectorSetKey[k comparable, v any] VectorSetKey[k, v]

// --- 内部辅助 ---

func (ctx *HttpVectorSetKey[k, v]) native() *VectorSetKey[k, v] {
	return (*VectorSetKey[k, v])(ctx)
}

// --- 接口实现 ---

func (ctx *HttpVectorSetKey[k, v]) GetKeyType() KeyType {
	return ctx.native().GetKeyType()
}
func (ctx *HttpVectorSetKey[k, v]) GetUseModer() bool {
	return ctx.native().GetUseModer()
}
func (ctx *HttpVectorSetKey[k, v]) GetValue() interface{} {
	return utils.CreateNonNilInstance[v]()
}
func (ctx *HttpVectorSetKey[k, v]) ValidDataKey() error {
	return ctx.native().ValidDataKey()
}
func (ctx *HttpVectorSetKey[k, v]) TimestampFiller(in interface{}) (err error) {
	return ctx.native().TimestampFiller(in)
}

// WithContext 实现：克隆并注入上下文
func (ctx *HttpVectorSetKey[k, v]) WithContext(key string, ds string) IHttpVectorSetKey {
	// 1. 获取底层 RedisKey 的副本
	newObj := ctx.native().Duplicate(key, ds)
	// 2. 包装并返回
	newKey := VectorSetKey[k, v]{RedisKey: newObj}
	newCtx := HttpVectorSetKey[k, v](newKey)
	return &newCtx
}

// --- 操作实现 ---

func (ctx *HttpVectorSetKey[k, v]) Create(args ...interface{}) error {
	return ctx.native().Create(args...)
}

func (ctx *HttpVectorSetKey[k, v]) DropIndex(deleteDocs bool) error {
	return ctx.native().DropIndex(deleteDocs)
}

func (ctx *HttpVectorSetKey[k, v]) Info() (map[string]interface{}, error) {
	return ctx.native().Info()
}

func (ctx *HttpVectorSetKey[k, v]) AliasAdd(alias string) error {
	return ctx.native().AliasAdd(alias)
}

func (ctx *HttpVectorSetKey[k, v]) AliasUpdate(alias string) error {
	return ctx.native().AliasUpdate(alias)
}

func (ctx *HttpVectorSetKey[k, v]) AliasDel(alias string) error {
	return ctx.native().AliasDel(alias)
}

func (ctx *HttpVectorSetKey[k, v]) TagVals(fieldName string) ([]string, error) {
	return ctx.native().TagVals(fieldName)
}

func (ctx *HttpVectorSetKey[k, v]) Search(query string, params ...interface{}) (int64, interface{}, error) {
	// native().Search 返回 (int64, []v, error)
	// 我们直接把 []v 作为 interface{} 返回，JSON Marshal 会处理好它
	return ctx.native().Search(query, params...)
}

// 工厂方法
func GetHttpVectorSetKey(Key string, rdsName string) (IHttpVectorSetKey, error) {
	_keyscope := KeyScope(Key)
	ikey, ok := HttpVectorSetKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found for: %s", _keyscope)
	}
	return ikey.WithContext(Key, rdsName), nil
}
