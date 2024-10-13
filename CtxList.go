package redisdb

import (
	"time"

	"github.com/doptime/logger"
	"github.com/redis/go-redis/v9"
)

type CtxList[k comparable, v any] struct {
	Ctx[k, v]
}

func ListKey[k comparable, v any](ops ...opSetter) *CtxList[k, v] {
	ctx := &CtxList[k, v]{}
	ctx.KeyType = "list"
	op := Option{}.buildOptions(ops...)
	if err := ctx.applyOption(op); err != nil {
		logger.Error().Err(err).Msg("data.New failed")
		return nil
	}
	return ctx
}

func (ctx *CtxList[k, v]) ConcatKey(fields ...interface{}) *CtxList[k, v] {
	return &CtxList[k, v]{ctx.Duplicate(ConcatedKeys(ctx.Key, fields...), ctx.RdsName)}
}
func (ctx *CtxList[k, v]) RPush(param ...v) error {
	vals, err := ctx.toValueStrsSlice(param...)
	if err != nil {
		return err
	}
	return ctx.Rds.RPush(ctx.Context, ctx.Key, vals...).Err()
}
func (ctx *CtxList[k, v]) RPushX(param ...v) error {
	vals, err := ctx.toValueStrsSlice(param...)
	if err != nil {
		return err
	}
	return ctx.Rds.RPushX(ctx.Context, ctx.Key, vals...).Err()
}

func (ctx *CtxList[k, v]) LPush(param ...v) error {
	vals, err := ctx.toValueStrsSlice(param...)
	if err != nil {
		return err
	}
	return ctx.Rds.LPush(ctx.Context, ctx.Key, vals...).Err()
}

func (ctx *CtxList[k, v]) RPop() (ret v, err error) {
	cmd := ctx.Rds.RPop(ctx.Context, ctx.Key)
	if err = cmd.Err(); err != nil {
		return ret, err
	}
	data, err := cmd.Bytes()
	if err != nil {
		return ret, err
	}
	return ctx.UnmarshalValue(data)
}

func (ctx *CtxList[k, v]) LPop() (ret v, err error) {
	cmd := ctx.Rds.LPop(ctx.Context, ctx.Key)
	if err := cmd.Err(); err != nil {
		return ret, err
	}
	data, err := cmd.Bytes()
	if err != nil {
		return ret, err
	}
	return ctx.UnmarshalValue(data)
}

func (ctx *CtxList[k, v]) LRange(start, stop int64) ([]v, error) {
	cmd := ctx.Rds.LRange(ctx.Context, ctx.Key, start, stop)
	if err := cmd.Err(); err != nil {
		return nil, err
	}
	values := make([]v, len(cmd.Val()))
	for i, v := range cmd.Val() {
		value, err := ctx.UnmarshalValue([]byte(v))
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

func (ctx *CtxList[k, v]) LRem(count int64, param v) error {
	val, err := ctx.MarshalValue(param)
	if err != nil {
		return err
	}
	return ctx.Rds.LRem(ctx.Context, ctx.Key, count, val).Err()
}

func (ctx *CtxList[k, v]) LSet(index int64, param v) error {
	val, err := ctx.MarshalValue(param)
	if err != nil {
		return err
	}
	return ctx.Rds.LSet(ctx.Context, ctx.Key, index, val).Err()
}

func (ctx *CtxList[k, v]) BLPop(timeout time.Duration) (ret v, err error) {
	cmd := ctx.Rds.BLPop(ctx.Context, timeout, ctx.Key)
	if err := cmd.Err(); err != nil {
		return ret, err
	}
	data, err := cmd.Result()
	if err != nil {
		return ret, err
	}
	return ctx.UnmarshalValue([]byte(data[1]))
}

func (ctx *CtxList[k, v]) BRPop(timeout time.Duration) (ret v, err error) {
	cmd := ctx.Rds.BRPop(ctx.Context, timeout, ctx.Key)
	if err := cmd.Err(); err != nil {
		return ret, err
	}
	data, err := cmd.Result()
	if err != nil {
		return ret, err
	}
	return ctx.UnmarshalValue([]byte(data[1]))
}

func (ctx *CtxList[k, v]) BRPopLPush(destination string, timeout time.Duration) (ret v, err error) {
	cmd := ctx.Rds.BRPopLPush(ctx.Context, ctx.Key, destination, timeout)
	if err := cmd.Err(); err != nil {
		return ret, err
	}
	data, err := cmd.Bytes()
	if err != nil {
		return ret, err
	}
	return ctx.UnmarshalValue(data)
}

func (ctx *CtxList[k, v]) LInsertBefore(pivot, param v) error {
	pivotStr, err := ctx.MarshalValue(pivot)
	if err != nil {
		return err
	}
	valStr, err := ctx.MarshalValue(param)
	if err != nil {
		return err
	}
	return ctx.Rds.LInsertBefore(ctx.Context, ctx.Key, pivotStr, valStr).Err()
}

func (ctx *CtxList[k, v]) LInsertAfter(pivot, param v) error {
	pivotStr, err := ctx.MarshalValue(pivot)
	if err != nil {
		return err
	}
	valStr, err := ctx.MarshalValue(param)
	if err != nil {
		return err
	}
	return ctx.Rds.LInsertAfter(ctx.Context, ctx.Key, pivotStr, valStr).Err()
}
func (ctx *CtxList[k, v]) Sort(sort *redis.Sort) ([]v, error) {
	cmd := ctx.Rds.Sort(ctx.Context, ctx.Key, sort)
	if err := cmd.Err(); err != nil {
		return nil, err
	}
	values := make([]v, len(cmd.Val()))
	for i, v := range cmd.Val() {
		value, err := ctx.UnmarshalValue([]byte(v))
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

func (ctx *CtxList[k, v]) LTrim(start, stop int64) error {
	return ctx.Rds.LTrim(ctx.Context, ctx.Key, start, stop).Err()
}

func (ctx *CtxList[k, v]) LIndex(index int64) (ret v, err error) {
	cmd := ctx.Rds.LIndex(ctx.Context, ctx.Key, index)
	if err := cmd.Err(); err != nil {
		return ret, err
	}
	data, err := cmd.Bytes()
	if err != nil {
		return ret, err
	}
	return ctx.UnmarshalValue(data)
}

func (ctx *CtxList[k, v]) LLen() (int64, error) {
	return ctx.Rds.LLen(ctx.Context, ctx.Key).Result()
}
