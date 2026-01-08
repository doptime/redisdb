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
	// don't register web data if it fully prepared
	if op != 0 && ctx.Key != "" {
		ctx.RegisterWebDataSchemaDocForWebVisit()
		ctx.RegisterKeyInterfaceForWebVisit()
	}
	return ctx
}

func (ctx *ZSetKey[k, v]) ZAdd(members ...redis.Z) (err error) {
	//MarshalRedisZ
	for i := range members {
		if members[i].Member != nil {
			members[i].Member, _ = msgpack.Marshal(members[i].Member)
		}
	}
	return ctx.Rds.ZAdd(ctx.Context, ctx.Key, members...).Err()
}

func (ctx *ZSetKey[k, v]) ZRem(members ...interface{}) (err error) {
	//msgpack marshal members to slice of bytes
	var bytes = make([][]byte, len(members))
	for i, member := range members {
		if bytes[i], err = msgpack.Marshal(member); err != nil {
			return err
		}
	}
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
func (ctx *ZSetKey[k, v]) ZRank(member interface{}) (rank int64, err error) {
	memberBytes, err := msgpack.Marshal(member)
	if err != nil {
		return 0, err
	}
	cmd := ctx.Rds.ZRank(ctx.Context, ctx.Key, string(memberBytes))
	return cmd.Val(), cmd.Err()
}
func (ctx *ZSetKey[k, v]) ZRevRank(member interface{}) (rank int64, err error) {
	memberBytes, err := msgpack.Marshal(member)
	//marshal member using msgpack
	if err != nil {
		return 0, err
	}
	cmd := ctx.Rds.ZRevRank(ctx.Context, ctx.Key, string(memberBytes))
	return cmd.Val(), cmd.Err()
}

func (ctx *ZSetKey[k, v]) ZScore(member v) (score float64, err error) {
	//marshal member using msgpack
	memberBytes, err := msgpack.Marshal(member)
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

func (ctx *ZSetKey[k, v]) ZIncrBy(increment float64, member v) error {
	memberBytes, err := ctx.SerializeValue(member)
	if err != nil {
		return err
	}
	return ctx.Rds.ZIncrBy(ctx.Context, ctx.Key, increment, memberBytes).Err()
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

func (ctx *ZSetKey[k, v]) UnmarshalToSlice(members []string) (out []v, err error) {
	out = make([]v, 0, len(members))
	//unmarshal each member in cmd.Result() using msgpack,to the type of element of out
	elemType := reflect.TypeOf(out).Elem()
	//don't set elemType to elemType.Elem() again, because out is a slice of pointer
	for _, member := range members {
		elem := reflect.New(elemType).Interface()
		if err := msgpack.Unmarshal([]byte(member), elem); err != nil {
			return out, err
		}
		out = append(out, *elem.(*v))
	}

	return out, nil
}

func (ctx *ZSetKey[k, v]) UnmarshalRedisZ(members []redis.Z) (out []v, scores []float64, err error) {
	var (
		str string
		ok  bool
	)
	out = make([]v, 0, len(members))
	//unmarshal each member in cmd.Result() using msgpack,to the type of element of out
	elemType := reflect.TypeOf(out).Elem()
	scores = make([]float64, len(members))
	for i, member := range members {
		if str, ok = member.Member.(string); !ok || str == "" {
			continue
		}
		elem := reflect.New(elemType).Interface()
		if err := msgpack.Unmarshal([]byte(str), elem); err != nil {
			return nil, nil, err
		}
		out = append(out, *elem.(*v))

		scores[i] = member.Score
	}
	return out, scores, nil
}
