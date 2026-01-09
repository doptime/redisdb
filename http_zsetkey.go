package redisdb

import (
	"fmt"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/redis/go-redis/v9"
)

// IHttpZSetKey 定义了 HTTP 层对 ZSet 的所有操作需求
// 返回值使用了 interface{}，这样底层具体的 []Profile 可以在运行时传递给 HTTP 层而不丢失类型信息
type IHttpZSetKey interface {
	// 基础接口
	GetKeyType() KeyType
	GetUseModer() bool
	ValidDataKey() error
	DeserializeValue(msgpack []byte) (rets interface{}, err error)
	DeserializeValues(msgpacks []string) (rets []interface{}, err error)
	TimestampFiller(in interface{}) (err error)

	// Context 注入 (核心：用于多租户/Key变换)
	WithContext(key string, ds string) IHttpZSetKey

	// 数据操作 (对应 startHttp.go 中的调用)
	ZAdd(members ...redis.Z) (err error)
	ZRem(members ...interface{}) (err error)
	ZCard() (int64, error)
	ZCount(min, max string) (int64, error)
	ZLexCount(min, max string) (int64, error)
	ZIncrBy(increment float64, member interface{}) error // member 改为 interface{} 以便通用

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

func (ctx *HttpZSetKey[k, v]) DeserializeValue(msgpack []byte) (rets interface{}, err error) {
	return ctx.native().DeserializeToValue(msgpack)
}
func (ctx *HttpZSetKey[k, v]) DeserializeValues(msgpacks []string) (rets []interface{}, err error) {
	return ctx.native().DeserializeToInterfaceSlice(msgpacks)
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

// ZIncrBy 特殊处理：因为 member 进来是 interface{}，需要转成 v
// 但因为底层 ZIncrBy 接收 v，而 v 可能是 *Profile。
// 这里我们做个妥协：底层 key_zset.go 的 ZIncrBy 如果能接受 interface{} 最好，
// 如果不能，这里需要类型断言。通常 HTTP 传入的 member 已经在 Deserialize 阶段处理过了，
// 但 ZIncrBy 的 member 通常是从 URL 或 Body 解析的简单值。
// 为了简化，建议 key_zset.go 的 ZIncrBy 参数改为 member interface{}，或者这里做简单处理。
// 假设底层 ZIncrBy 的签名是 func (ctx *ZSetKey[k, v]) ZIncrBy(incr float64, member v)
func (ctx *HttpZSetKey[k, v]) ZIncrBy(increment float64, member interface{}) error {
	// 这是一个潜在的痛点。如果 v 是 struct，这里传 interface{} 可能会 panic。
	// 但通常 ZSet 的 member 是 string 或 key。
	// 如果 v 是 interface{} (泛型擦除后)，直接转。
	if vVal, ok := member.(v); ok {
		return ctx.native().ZIncrBy(increment, vVal)
	}
	// Fallback: 如果传进来的是 string/bytes，尝试强转 (视具体业务逻辑而定)
	// 暂时不做处理，让它 panic 或者报错，或者你在 key_zset.go 里修改 ZIncrBy 接受 interface{}
	return fmt.Errorf("ZIncrBy type mismatch: expected %T, got %T", *new(v), member)
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
