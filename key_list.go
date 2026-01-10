package redisdb

import (
	"time"

	"github.com/doptime/logger"
	"github.com/redis/go-redis/v9"
)

type ListKey[v any] struct {
	RedisKey[string, v]
}

func NewListKey[v any](ops ...Option) *ListKey[v] {
	ctx := &ListKey[v]{RedisKey: RedisKey[string, v]{KeyType: KeyTypeList}}
	if err := ctx.applyOptionsAndCheck(KeyTypeList, ops...); err != nil {
		logger.Error().Err(err).Msg("redisdb.NewListKey failed")
		return nil
	}
	ctx.InitFunc()
	return ctx
}
func (ctx *ListKey[v]) HttpOn(op ListOp) (ctx1 *ListKey[v]) {
	httpAllow(ctx.Key, uint64(op))
	// don't register web data if it fully prepared
	if op != 0 && ctx.Key != "" {
		ctx.RegisterWebDataSchemaDocForWebVisit()
		ctx.RegisterHttpInterface()
	}
	return ctx
}
func (ctx *ListKey[v]) RegisterHttpInterface() {
	// register the key interface for web access
	keyScope := KeyScope(ctx.Key)
	hskey := ListKey[v]{ctx.Duplicate(ctx.Key, ctx.RdsName)}
	IListKey := HttpListKey[v](hskey)
	HttpListKeyMap.Set(keyScope+":"+ctx.RdsName, &IListKey)
}

func (ctx *ListKey[v]) ConcatKey(fields ...interface{}) *ListKey[v] {
	return &ListKey[v]{ctx.Duplicate(ConcatedKeys(ctx.Key, fields...), ctx.RdsName)}
}
func (ctx *ListKey[v]) RPush(param ...v) error {
	vals, err := ctx.toValueStrsSlice(param...)
	if err != nil {
		return err
	}
	return ctx.Rds.RPush(ctx.Context, ctx.Key, vals...).Err()
}

func (ctx *ListKey[v]) LPush(param ...v) error {
	vals, err := ctx.toValueStrsSlice(param...)
	if err != nil {
		return err
	}
	return ctx.Rds.LPush(ctx.Context, ctx.Key, vals...).Err()
}

func (ctx *ListKey[v]) RPushX(param ...v) error {
	vals, err := ctx.toValueStrsSlice(param...)
	if err != nil {
		return err
	}
	return ctx.Rds.RPushX(ctx.Context, ctx.Key, vals...).Err()
}
func (ctx *ListKey[v]) LPushX(param ...v) error {
	vals, err := ctx.toValueStrsSlice(param...)
	if err != nil {
		return err
	}
	return ctx.Rds.LPushX(ctx.Context, ctx.Key, vals...).Err()
}

func (ctx *ListKey[v]) RPop() (ret v, err error) {
	cmd := ctx.Rds.RPop(ctx.Context, ctx.Key)
	if err = cmd.Err(); err != nil {
		return ret, err
	}
	data, err := cmd.Bytes()
	if err != nil {
		return ret, err
	}
	return ctx.DeserializeToValue(data)
}

func (ctx *ListKey[v]) LPop() (ret v, err error) {
	cmd := ctx.Rds.LPop(ctx.Context, ctx.Key)
	if err := cmd.Err(); err != nil {
		return ret, err
	}
	data, err := cmd.Bytes()
	if err != nil {
		return ret, err
	}
	return ctx.DeserializeToValue(data)
}

func (ctx *ListKey[v]) LRange(start, stop int64) ([]v, error) {
	cmd := ctx.Rds.LRange(ctx.Context, ctx.Key, start, stop)
	if err := cmd.Err(); err != nil {
		return nil, err
	}
	values := make([]v, len(cmd.Val()))
	for i, v := range cmd.Val() {
		value, err := ctx.DeserializeToValue([]byte(v))
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

func (ctx *ListKey[v]) LRem(count int64, param v) error {
	val, err := ctx.SerializeValue(param)
	if err != nil {
		return err
	}
	return ctx.Rds.LRem(ctx.Context, ctx.Key, count, val).Err()
}

func (ctx *ListKey[v]) LSet(index int64, param v) error {
	val, err := ctx.SerializeValue(param)
	if err != nil {
		return err
	}
	return ctx.Rds.LSet(ctx.Context, ctx.Key, index, val).Err()
}
func (ctx *ListKey[v]) LIndex(ind int64) (ret v, err error) {
	cmd := ctx.Rds.LIndex(ctx.Context, ctx.Key, ind)
	if err = cmd.Err(); err != nil {
		return ret, err
	}
	data, err := cmd.Bytes()
	if err != nil {
		return ret, err
	}
	return ctx.DeserializeToValue(data)
}

func (ctx *ListKey[v]) BLPop(timeout time.Duration) (ret v, err error) {
	cmd := ctx.Rds.BLPop(ctx.Context, timeout, ctx.Key)
	if err := cmd.Err(); err != nil {
		return ret, err
	}
	data, err := cmd.Result()
	if err != nil {
		return ret, err
	}
	return ctx.DeserializeToValue([]byte(data[1]))
}

func (ctx *ListKey[v]) BRPop(timeout time.Duration) (ret v, err error) {
	cmd := ctx.Rds.BRPop(ctx.Context, timeout, ctx.Key)
	if err := cmd.Err(); err != nil {
		return ret, err
	}
	data, err := cmd.Result()
	if err != nil {
		return ret, err
	}
	return ctx.DeserializeToValue([]byte(data[1]))
}

func (ctx *ListKey[v]) BRPopLPush(destination string, timeout time.Duration) (ret v, err error) {
	cmd := ctx.Rds.BRPopLPush(ctx.Context, ctx.Key, destination, timeout)
	if err := cmd.Err(); err != nil {
		return ret, err
	}
	data, err := cmd.Bytes()
	if err != nil {
		return ret, err
	}
	return ctx.DeserializeToValue(data)
}

func (ctx *ListKey[v]) LInsertBefore(pivot, param v) error {
	pivotStr, err := ctx.SerializeValue(pivot)
	if err != nil {
		return err
	}
	valStr, err := ctx.SerializeValue(param)
	if err != nil {
		return err
	}
	return ctx.Rds.LInsertBefore(ctx.Context, ctx.Key, pivotStr, valStr).Err()
}

func (ctx *ListKey[v]) LInsertAfter(pivot, param v) error {
	pivotStr, err := ctx.SerializeValue(pivot)
	if err != nil {
		return err
	}
	valStr, err := ctx.SerializeValue(param)
	if err != nil {
		return err
	}
	return ctx.Rds.LInsertAfter(ctx.Context, ctx.Key, pivotStr, valStr).Err()
}
func (ctx *ListKey[v]) Sort(sort *redis.Sort) ([]v, error) {
	cmd := ctx.Rds.Sort(ctx.Context, ctx.Key, sort)
	if err := cmd.Err(); err != nil {
		return nil, err
	}
	values := make([]v, len(cmd.Val()))
	for i, v := range cmd.Val() {
		value, err := ctx.DeserializeToValue([]byte(v))
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

func (ctx *ListKey[v]) LTrim(start, stop int64) error {
	return ctx.Rds.LTrim(ctx.Context, ctx.Key, start, stop).Err()
}

func (ctx *ListKey[v]) LLen() (int64, error) {
	return ctx.Rds.LLen(ctx.Context, ctx.Key).Result()
}
