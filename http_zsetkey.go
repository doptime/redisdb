package redisdb

import (
	"fmt"

	"github.com/doptime/redisdb/utils"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/redis/go-redis/v9"
)

// IHttpZSetKey 定义了 HTTP 层对 ZSet 的所有操作需求
// 返回值使用了 interface{}，这样底层具体的 []Profile 可以在运行时传递给 HTTP 层而不丢失类型信息
type IHttpZSetKey interface {
	// 基础接口
	GetKeyType() KeyType
	GetUseModer() bool
	GetValue() interface{}
	ValidDataKey() error
	TimestampFiller(in interface{}) (err error)

	// Context 注入 (核心：用于多租户/Key变换)
	WithContext(key string, ds string) IHttpZSetKey

	// 数据操作 (对应 startHttp.go 中的调用)
	ZAdd(members ...redis.Z) (err error)
	ZRem(members ...interface{}) (err error)
	ZCard() (int64, error)
	ZCount(min, max string) (int64, error)
	ZLexCount(min, max string) (int64, error)
	ZIncrBy(increment float64, member interface{}) (float64, error)

	ZScore(member interface{}) (score float64, err error)
	ZRank(member interface{}) (rank int64, err error)
	ZRevRank(member interface{}) (rank int64, err error)

	// Range 类操作 (返回值设为 interface{}，底层是 []v)
	ZRange(start, stop int64) (interface{}, error)
	ZRangeWithScores(start, stop int64) (interface{}, []float64, error)
	ZRevRange(start, stop int64) (interface{}, error)
	ZRevRangeWithScores(start, stop int64) (interface{}, []float64, error)
	ZRangeByScore(opt *redis.ZRangeBy) (interface{}, error)
	ZRangeByScoreWithScores(opt *redis.ZRangeBy) (interface{}, []float64, error)
	ZRevRangeByScore(opt *redis.ZRangeBy) (interface{}, error)
	ZRevRangeByScoreWithScores(opt *redis.ZRangeBy) (interface{}, []float64, error)

	ZPopMax(count int64) (interface{}, []float64, error)
	ZPopMin(count int64) (interface{}, []float64, error)

	ZScan(cursor uint64, match string, count int64) (values interface{}, rcursor uint64, err error)
}

var HttpZSetKeyMap cmap.ConcurrentMap[string, IHttpZSetKey] = cmap.New[IHttpZSetKey]()

// HttpZSetKey 是 ZSetKey 的 Wrapper，实现了 IHttpZSetKey 接口
type HttpZSetKey[k comparable, v any] ZSetKey[k, v]

// --- 基础转换 ---

func (ctx *HttpZSetKey[k, v]) native() *ZSetKey[k, v] {
	return (*ZSetKey[k, v])(ctx)
}

func (ctx *HttpZSetKey[k, v]) GetKeyType() KeyType {
	return ctx.native().GetKeyType()
}
func (ctx *HttpZSetKey[k, v]) GetUseModer() bool {
	return ctx.native().GetUseModer()
}
func (ctx *HttpZSetKey[k, v]) GetValue() interface{} {
	return utils.CreateNonNilInstance[v]()
}
func (ctx *HttpZSetKey[k, v]) ValidDataKey() error {
	return ctx.native().ValidDataKey()
}

// WithContext 实现：克隆自己，修改 Key 和 DS，返回接口
func (ctx *HttpZSetKey[k, v]) WithContext(key string, RedisDataSource string) IHttpZSetKey {
	// 1. 获取原始对象的副本 (浅拷贝结构体)
	newObj := ctx.native().Duplicate(key, RedisDataSource)
	newCtx := HttpZSetKey[k, v]{RedisKey: newObj}
	return &newCtx
}

// --- 数据操作实现 (转发给底层 ZSetKey) ---

func (ctx *HttpZSetKey[k, v]) ZAdd(members ...redis.Z) error {
	return ctx.native().ZAdd(members...)
}

func (ctx *HttpZSetKey[k, v]) ZRem(members ...interface{}) error {
	return ctx.native().ZRem(members...)
}

func (ctx *HttpZSetKey[k, v]) ZCard() (int64, error) {
	return ctx.native().ZCard()
}
func (ctx *HttpZSetKey[k, v]) ZCount(min, max string) (int64, error) {
	return ctx.native().ZCount(min, max)
}
func (ctx *HttpZSetKey[k, v]) ZLexCount(min, max string) (int64, error) {
	return ctx.native().ZLexCount(min, max)
}

func (ctx *HttpZSetKey[k, v]) ZIncrBy(increment float64, member interface{}) (float64, error) {
	return ctx.native().ZIncrBy(increment, member)
}

func (ctx *HttpZSetKey[k, v]) ZScore(member interface{}) (float64, error) {
	if vVal, ok := member.(v); ok {
		return ctx.native().ZScore(vVal)
	}
	// 尝试处理 member 是 string 的情况 (如果是 raw key)
	// 这里的逻辑取决于你的 key_zset.go 实现是否严格要求 v
	// 建议：key_zset.go 里的 ZScore 参数最好也是 interface{}，因为它主要是用来做 Key
	return 0, fmt.Errorf("ZScore type mismatch")
}

func (ctx *HttpZSetKey[k, v]) ZRank(member interface{}) (int64, error) {
	return ctx.native().ZRank(member)
}
func (ctx *HttpZSetKey[k, v]) ZRevRank(member interface{}) (int64, error) {
	return ctx.native().ZRevRank(member)
}

// Range 操作 - 直接返回，因为 interface{} 可以容纳 []v
func (ctx *HttpZSetKey[k, v]) ZRange(start, stop int64) (interface{}, error) {
	return ctx.native().ZRange(start, stop)
}
func (ctx *HttpZSetKey[k, v]) ZRangeWithScores(start, stop int64) (interface{}, []float64, error) {
	return ctx.native().ZRangeWithScores(start, stop)
}
func (ctx *HttpZSetKey[k, v]) ZRevRange(start, stop int64) (interface{}, error) {
	return ctx.native().ZRevRange(start, stop)
}
func (ctx *HttpZSetKey[k, v]) ZRevRangeWithScores(start, stop int64) (interface{}, []float64, error) {
	return ctx.native().ZRevRangeWithScores(start, stop)
}
func (ctx *HttpZSetKey[k, v]) ZRangeByScore(opt *redis.ZRangeBy) (interface{}, error) {
	return ctx.native().ZRangeByScore(opt)
}
func (ctx *HttpZSetKey[k, v]) ZRangeByScoreWithScores(opt *redis.ZRangeBy) (interface{}, []float64, error) {
	return ctx.native().ZRangeByScoreWithScores(opt)
}
func (ctx *HttpZSetKey[k, v]) ZRevRangeByScore(opt *redis.ZRangeBy) (interface{}, error) {
	return ctx.native().ZRevRangeByScore(opt)
}
func (ctx *HttpZSetKey[k, v]) ZRevRangeByScoreWithScores(opt *redis.ZRangeBy) (interface{}, []float64, error) {
	return ctx.native().ZRevRangeByScoreWithScores(opt)
}
func (ctx *HttpZSetKey[k, v]) ZPopMax(count int64) (interface{}, []float64, error) {
	return ctx.native().ZPopMax(count)
}
func (ctx *HttpZSetKey[k, v]) ZPopMin(count int64) (interface{}, []float64, error) {
	return ctx.native().ZPopMin(count)
}
func (ctx *HttpZSetKey[k, v]) ZScan(cursor uint64, match string, count int64) (interface{}, uint64, error) {
	return ctx.native().ZScan(cursor, match, count)
}

// 工厂方法
func GetHttpZSetKey(Key string, rdsName string) (IHttpZSetKey, error) {
	_keyscope := KeyScope(Key)
	// 这里返回的是 IHttpZSetKey 接口，底层可能是 HttpZSetKey[string, *Profile]
	ikey, ok := HttpZSetKeyMap.Get(_keyscope + ":" + rdsName)
	if !ok {
		return nil, fmt.Errorf("key schema not found for scope: %s", _keyscope)
	}
	return ikey.WithContext(Key, rdsName), nil
}
