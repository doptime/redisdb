package redisdb

import (
	"reflect"

	"github.com/doptime/logger"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

type ZSetKey[k comparable, v any] struct {
	RedisKey[k, v]
}

func NewZSetKey[k comparable, v any](ops ...Option) *ZSetKey[k, v] {
	ctx := &ZSetKey[k, v]{RedisKey: RedisKey[k, v]{KeyType: KeyTypeZSet}}
	if err := ctx.applyOptionsAndCheck(KeyTypeZSet, ops...); err != nil {
		logger.Error().Err(err).Msg("redisdb.NewZSetKey failed")
		return nil
	}
	ctx.InitFunc()
	return ctx
}

func (ctx *ZSetKey[k, v]) ConcatKey(fields ...interface{}) *ZSetKey[k, v] {
	return &ZSetKey[k, v]{ctx.RedisKey.Duplicate(ConcatedKeys(ctx.Key, fields...), ctx.RdsName)}
}

func (ctx *ZSetKey[k, v]) HttpOn(op ZSetOp) (ctx1 *ZSetKey[k, v]) {
	httpAllow(ctx.Key, uint64(op))
	if op != 0 && ctx.Key != "" {
		ctx.RegisterWebDataSchemaDocForWebVisit()
		ctx.RegisterHttpInterface()
	}
	return ctx
}

func (ctx *ZSetKey[k, v]) RegisterHttpInterface() {
	keyScope := KeyScope(ctx.Key)
	hskey := ZSetKey[k, v]{ctx.Duplicate(ctx.Key, ctx.RdsName)}
	IZSetKey := HttpZSetKey[k, v](hskey)
	HttpZSetKeyMap.Set(keyScope+":"+ctx.RdsName, &IZSetKey)
}

// ZAdd: 批量添加
func (ctx *ZSetKey[k, v]) ZAdd(members ...redis.Z) (err error) {
	// 注意：为了不修改外部传入的切片，建议这里处理 carefully
	// 但为了性能，直接修改 members 里的 Member 字段为 []byte
	for i := range members {
		// 如果 Member 是 v 类型，或者是 struct，尝试序列化
		// 如果已经是 []byte 或 string，msgpack.Marshal 也会处理
		if members[i].Member != nil {
			// 这里假设 HTTP 层传进来的是 Struct/Map，需要序列化存储
			b, err := msgpack.Marshal(members[i].Member)
			if err == nil {
				members[i].Member = b
			}
		}
	}
	return ctx.Rds.ZAdd(ctx.Context, ctx.Key, members...).Err()
}

func (ctx *ZSetKey[k, v]) ZRem(members ...interface{}) (err error) {
	var bytes = make([][]byte, len(members))
	var _msgpack string
	for i, member := range members {
		// 尝试断言为 v 以利用自定义序列化(如果有)，否则通用序列化
		if vVal, ok := member.(v); ok {
			if _msgpack, err = ctx.SerializeValue(vVal); err != nil {
				return err
			}
			bytes[i] = []byte(_msgpack)
		} else {
			if bytes[i], err = msgpack.Marshal(member); err != nil {
				return err
			}
		}
	}
	// Pipeline 优化
	var redisPipe = ctx.Rds.Pipeline()
	for _, memberBytes := range bytes {
		redisPipe.ZRem(ctx.Context, ctx.Key, memberBytes)
	}
	_, err = redisPipe.Exec(ctx.Context)
	return err
}

func (ctx *ZSetKey[k, v]) ZRange(start, stop int64) (members []v, err error) {
	cmd := ctx.Rds.ZRange(ctx.Context, ctx.Key, start, stop)
	if err = cmd.Err(); err != nil && err != redis.Nil {
		return nil, err
	}
	return ctx.UnmarshalToSlice(cmd.Val())
}

func (ctx *ZSetKey[k, v]) ZRangeWithScores(start, stop int64) (members []v, scores []float64, err error) {
	cmd := ctx.Rds.ZRangeWithScores(ctx.Context, ctx.Key, start, stop)
	return ctx.UnmarshalRedisZ(cmd.Val())
}
func (ctx *ZSetKey[k, v]) ZRevRangeWithScores(start, stop int64) (members []v, scores []float64, err error) {
	cmd := ctx.Rds.ZRevRangeWithScores(ctx.Context, ctx.Key, start, stop)
	return ctx.UnmarshalRedisZ(cmd.Val())
}

// ZRank: 参数改为 interface{}
func (ctx *ZSetKey[k, v]) ZRank(member interface{}) (rank int64, err error) {
	memberBytes, err := ctx.serializeInterface(member)
	if err != nil {
		return 0, err
	}
	cmd := ctx.Rds.ZRank(ctx.Context, ctx.Key, string(memberBytes))
	return cmd.Val(), cmd.Err()
}

// ZRevRank: 参数改为 interface{}
func (ctx *ZSetKey[k, v]) ZRevRank(member interface{}) (rank int64, err error) {
	memberBytes, err := ctx.serializeInterface(member)
	if err != nil {
		return 0, err
	}
	cmd := ctx.Rds.ZRevRank(ctx.Context, ctx.Key, string(memberBytes))
	return cmd.Val(), cmd.Err()
}

// ZScore: 参数改为 interface{}
func (ctx *ZSetKey[k, v]) ZScore(member interface{}) (score float64, err error) {
	memberBytes, err := ctx.serializeInterface(member)
	if err != nil {
		return 0, err
	}
	cmd := ctx.Rds.ZScore(ctx.Context, ctx.Key, string(memberBytes))
	if err = cmd.Err(); err != nil {
		return 0, err
	}
	return cmd.Result()
}

func (ctx *ZSetKey[k, v]) ZCard() (int64, error) {
	return ctx.Rds.ZCard(ctx.Context, ctx.Key).Result()
}

func (ctx *ZSetKey[k, v]) ZCount(min, max string) (int64, error) {
	return ctx.Rds.ZCount(ctx.Context, ctx.Key, min, max).Result()
}

func (ctx *ZSetKey[k, v]) ZRangeByScore(opt *redis.ZRangeBy) (out []v, err error) {
	cmd := ctx.Rds.ZRangeByScore(ctx.Context, ctx.Key, opt)
	return ctx.UnmarshalToSlice(cmd.Val())
}
func (ctx *ZSetKey[k, v]) ZRangeByScoreWithScores(opt *redis.ZRangeBy) (out []v, scores []float64, err error) {
	cmd := ctx.Rds.ZRangeByScoreWithScores(ctx.Context, ctx.Key, opt)
	if err = cmd.Err(); err != nil {
		return nil, nil, err
	}
	return ctx.UnmarshalRedisZ(cmd.Val())
}

func (ctx *ZSetKey[k, v]) ZRevRangeByScore(opt *redis.ZRangeBy) (out []v, err error) {
	cmd := ctx.Rds.ZRevRangeByScore(ctx.Context, ctx.Key, opt)
	return ctx.UnmarshalToSlice(cmd.Val())
}

func (ctx *ZSetKey[k, v]) ZRevRange(start, stop int64) (out []v, err error) {
	cmd := ctx.Rds.ZRevRange(ctx.Context, ctx.Key, start, stop)
	if err := cmd.Err(); err != nil {
		return nil, err
	}
	return ctx.UnmarshalToSlice(cmd.Val())
}

func (ctx *ZSetKey[k, v]) ZRevRangeByScoreWithScores(opt *redis.ZRangeBy) (out []v, scores []float64, err error) {
	cmd := ctx.Rds.ZRevRangeByScoreWithScores(ctx.Context, ctx.Key, opt)
	if err = cmd.Err(); err != nil {
		return nil, nil, err
	}
	return ctx.UnmarshalRedisZ(cmd.Val())
}
func (ctx *ZSetKey[k, v]) ZRemRangeByRank(start, stop int64) (err error) {
	return ctx.Rds.ZRemRangeByRank(ctx.Context, ctx.Key, start, stop).Err()
}

func (ctx *ZSetKey[k, v]) ZRemRangeByScore(min, max string) error {
	return ctx.Rds.ZRemRangeByScore(ctx.Context, ctx.Key, min, max).Err()
}

// ZIncrBy: 参数改为 interface{}
func (ctx *ZSetKey[k, v]) ZIncrBy(increment float64, member interface{}) (float64, error) {
	memberBytes, err := ctx.serializeInterface(member)
	if err != nil {
		return 0, err
	}
	// ctx.Rds.ZIncrBy 返回 *FloatCmd
	// .Result() 返回 (float64, error)
	return ctx.Rds.ZIncrBy(ctx.Context, ctx.Key, increment, string(memberBytes)).Result()
}

func (ctx *ZSetKey[k, v]) ZPopMax(count int64) (out []v, scores []float64, err error) {
	cmd := ctx.Rds.ZPopMax(ctx.Context, ctx.Key, count)
	if err = cmd.Err(); err != nil {
		return nil, nil, err
	}
	return ctx.UnmarshalRedisZ(cmd.Val())
}
func (ctx *ZSetKey[k, v]) ZPopMin(count int64) (out []v, scores []float64, err error) {
	cmd := ctx.Rds.ZPopMin(ctx.Context, ctx.Key, count)
	if err = cmd.Err(); err != nil {
		return nil, nil, err
	}
	return ctx.UnmarshalRedisZ(cmd.Val())
}
func (ctx *ZSetKey[k, v]) ZLexCount(min, max string) (int64, error) {
	return ctx.Rds.ZLexCount(ctx.Context, ctx.Key, min, max).Result()
}

func (ctx *ZSetKey[k, v]) ZScan(cursor uint64, match string, count int64) (values []v, rcursor uint64, err error) {
	var strs []string
	strs, rcursor, err = ctx.Rds.ZScan(ctx.Context, ctx.Key, cursor, match, count).Result()
	values = make([]v, 0, len(strs))
	for _, s := range strs {
		if _v, err := ctx.DeserializeToValue([]byte(s)); err == nil {
			values = append(values, _v)
		}
	}
	return values, rcursor, err
}

// 辅助：统一序列化 interface{}，优先尝试转为 v
func (ctx *ZSetKey[k, v]) serializeInterface(member interface{}) (string, error) {
	if vVal, ok := member.(v); ok {
		return ctx.SerializeValue(vVal)
	}
	// Fallback
	bytes, err := msgpack.Marshal(member)
	return string(bytes), err
}

func (ctx *ZSetKey[k, v]) UnmarshalToSlice(members []string) (out []v, err error) {
	out = make([]v, 0, len(members))
	// 修正：确保 out 的 slice 类型正确
	// reflect.TypeOf(out).Elem() 是 v
	vType := reflect.TypeOf(out).Elem()

	for _, member := range members {
		// 创建 v 的指针
		elemPtr := reflect.New(vType)
		if err := msgpack.Unmarshal([]byte(member), elemPtr.Interface()); err != nil {
			return out, err
		}
		out = append(out, elemPtr.Elem().Interface().(v))
	}
	return out, nil
}

func (ctx *ZSetKey[k, v]) UnmarshalRedisZ(members []redis.Z) (out []v, scores []float64, err error) {
	out = make([]v, 0, len(members))
	scores = make([]float64, len(members))
	vType := reflect.TypeOf(out).Elem()

	for i, member := range members {
		str, ok := member.Member.(string)
		if !ok || str == "" {
			continue // 或处理错误
		}

		elemPtr := reflect.New(vType)
		if err := msgpack.Unmarshal([]byte(str), elemPtr.Interface()); err != nil {
			return nil, nil, err
		}
		out = append(out, elemPtr.Elem().Interface().(v))
		scores[i] = member.Score
	}
	return out, scores, nil
}
