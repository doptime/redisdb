package redisdb

import "github.com/doptime/logger"

type SetKey[k comparable, v any] struct {
	RedisKey[k, v]
}

func NewSetKey[k comparable, v any](ops ...Option) *SetKey[k, v] {
	ctx := &SetKey[k, v]{}
	for _, op := range ops {
		if err := ctx.applyOption(keyTypeSetKey, op); err != nil {
			logger.Error().Err(err).Msg("data.New failed")
			return nil
		}
	}
	ctx.InitFunc()
	return ctx
}
func (ctx *SetKey[k, v]) ConcatKey(fields ...interface{}) *SetKey[k, v] {
	return &SetKey[k, v]{ctx.Duplicate(ConcatedKeys(ctx.Key, fields...), ctx.RdsName)}
}

// append to Set
func (ctx *SetKey[k, v]) SAdd(param v) (err error) {
	valStr, err := ctx.SerializeValue(param)
	if err != nil {
		return err
	}
	return ctx.Rds.SAdd(ctx.Context, ctx.Key, valStr).Err()
}

func (ctx *SetKey[k, v]) SCard() (int64, error) {
	return ctx.Rds.SCard(ctx.Context, ctx.Key).Result()
}

func (ctx *SetKey[k, v]) SRem(param v) error {
	valStr, err := ctx.SerializeValue(param)
	if err != nil {
		return err
	}
	return ctx.Rds.SRem(ctx.Context, ctx.Key, valStr).Err()
}
func (ctx *SetKey[k, v]) SIsMember(param v) (bool, error) {
	valStr, err := ctx.SerializeValue(param)
	if err != nil {
		return false, err
	}
	return ctx.Rds.SIsMember(ctx.Context, ctx.Key, valStr).Result()
}

func (ctx *SetKey[k, v]) SMembers() ([]v, error) {
	cmd := ctx.Rds.SMembers(ctx.Context, ctx.Key)
	if err := cmd.Err(); err != nil {
		return nil, err
	}
	return ctx.DeserializeValues(cmd.Val())
}
func (ctx *SetKey[k, v]) SScan(cursor uint64, match string, count int64) ([]v, uint64, error) {
	cmd := ctx.Rds.SScan(ctx.Context, ctx.Key, cursor, match, count)
	if err := cmd.Err(); err != nil {
		return nil, 0, err
	}
	Strs, cursor, err := cmd.Result()
	if err != nil {
		return nil, 0, err
	}
	values, err := ctx.DeserializeValues(Strs)
	if err != nil {
		return nil, 0, err
	}
	return values, cursor, nil
}
