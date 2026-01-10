package redisdb

import (
	"github.com/doptime/logger"
	"github.com/redis/go-redis/v9"
)

type StreamKey[k comparable, v any] struct {
	RedisKey[k, v]
}

func NewStreamKey[k comparable, v any](ops ...Option) *StreamKey[k, v] {
	ctx := &StreamKey[k, v]{RedisKey: RedisKey[k, v]{KeyType: KeyTypeStream}}
	if err := ctx.applyOptionsAndCheck(KeyTypeStream, ops...); err != nil {
		logger.Error().Err(err).Msg("redisdb.NewStreamKey failed")
		return nil
	}
	ctx.InitFunc()
	return ctx
}

func (ctx *StreamKey[k, v]) ConcatKey(fields ...interface{}) *StreamKey[k, v] {
	return &StreamKey[k, v]{ctx.RedisKey.Duplicate(ConcatedKeys(ctx.Key, fields...), ctx.RdsName)}
}

func (ctx *StreamKey[k, v]) HttpOn(op StreamOp) (ctx1 *StreamKey[k, v]) {
	httpAllow(ctx.Key, uint64(op))
	// don't register web data if it fully prepared
	if op != 0 && ctx.Key != "" {
		ctx.RegisterWebDataSchemaDocForWebVisit()
		ctx.RegisterHttpInterface()
	}
	return ctx
}

func (ctx *StreamKey[k, v]) RegisterHttpInterface() {
	// register the key interface for web access
	keyScope := KeyScope(ctx.Key)
	hskey := StreamKey[k, v]{ctx.Duplicate(ctx.Key, ctx.RdsName)}
	IHashKey := HttpStreamKey[k, v](hskey)
	HttpStreamKeyMap.Set(keyScope+":"+ctx.RdsName, &IHashKey)
}

// --- 新增的核心操作方法 ---

func (ctx *StreamKey[k, v]) XAdd(args *redis.XAddArgs) (string, error) {
	// 确保 Stream Key 是正确的 (Context Key)
	args.Stream = ctx.Key
	return ctx.Rds.XAdd(ctx.Context, args).Result()
}

func (ctx *StreamKey[k, v]) XDel(ids ...string) (int64, error) {
	return ctx.Rds.XDel(ctx.Context, ctx.Key, ids...).Result()
}

func (ctx *StreamKey[k, v]) XLen() (int64, error) {
	return ctx.Rds.XLen(ctx.Context, ctx.Key).Result()
}

func (ctx *StreamKey[k, v]) XRange(start, stop string) ([]redis.XMessage, error) {
	return ctx.Rds.XRange(ctx.Context, ctx.Key, start, stop).Result()
}

func (ctx *StreamKey[k, v]) XRangeN(start, stop string, count int64) ([]redis.XMessage, error) {
	return ctx.Rds.XRangeN(ctx.Context, ctx.Key, start, stop, count).Result()
}

func (ctx *StreamKey[k, v]) XRevRange(start, stop string) ([]redis.XMessage, error) {
	return ctx.Rds.XRevRange(ctx.Context, ctx.Key, start, stop).Result()
}

func (ctx *StreamKey[k, v]) XRevRangeN(start, stop string, count int64) ([]redis.XMessage, error) {
	return ctx.Rds.XRevRangeN(ctx.Context, ctx.Key, start, stop, count).Result()
}

func (ctx *StreamKey[k, v]) XRead(args *redis.XReadArgs) ([]redis.XStream, error) {
	// 如果 args.Streams 没有指定，通常需要在上层处理，这里假设调用者会传入完整的 keys
	// 但对于 StreamKey 对象，我们通常只读自己。
	// 如果上层没有传 Streams，我们默认读自己
	if len(args.Streams) == 0 {
		args.Streams = []string{ctx.Key, "$"} // 默认读最新
	}
	return ctx.Rds.XRead(ctx.Context, args).Result()
}

// XInfoGroups 等其他管理命令按需添加...
