package redisdb

import (
	"strings"
	"time"

	"github.com/doptime/logger"
)

type StringKey[k comparable, v any] struct {
	RedisKey[k, v]
}

func NewStringKey[k comparable, v any](ops ...Option) *StringKey[k, v] {
	ctx := &StringKey[k, v]{RedisKey: RedisKey[k, v]{KeyType: KeyTypeString}}
	if err := ctx.applyOptionsAndCheck(KeyTypeString, ops...); err != nil {
		logger.Error().Err(err).Msg("redisdb.NewStringKey failed")
		return nil
	}
	ctx.InitFunc()
	return ctx
}

func (ctx *StringKey[k, v]) ConcatKey(fields ...interface{}) *StringKey[k, v] {
	return &StringKey[k, v]{ctx.RedisKey.Duplicate(ConcatedKeys(ctx.Key, fields...), ctx.RdsName)}
}

func (ctx *StringKey[k, v]) HttpOn(op StringOp) (ctx1 *StringKey[k, v]) {
	HttpPermissions.Set(KeyScope(ctx.Key), uint64(op))
	// don't register web data if it fully prepared
	if op != 0 && ctx.Key != "" {
		ctx.RegisterWebData()
		RediskeyForWeb.Set(ctx.Key+":"+ctx.RdsName, ctx)
	}
	return ctx
}

func (ctx *StringKey[k, v]) Get(Field k) (value v, err error) {
	FieldStr, err := ctx.SerializeKey(Field)
	if err != nil {
		return value, err
	}
	var keyFields []string
	if len(ctx.Key) > 0 {
		keyFields = append(keyFields, ctx.Key)
	}
	if len(FieldStr) > 0 {
		keyFields = append(keyFields, FieldStr)
	}

	cmd := ctx.Rds.Get(ctx.Context, strings.Join(keyFields, ":"))
	if err := cmd.Err(); err != nil {
		return value, err
	}
	data, err := cmd.Bytes()
	if err != nil {
		return value, err
	}
	return ctx.DeserializeToValue(data)
}

func (ctx *StringKey[k, v]) Set(key k, value v, expiration time.Duration) error {
	keyStr, err := ctx.SerializeKey(key)
	if err != nil {
		return err
	}
	valStr, err := ctx.SerializeValue(value)
	if err != nil {
		return err
	}
	return ctx.Rds.Set(ctx.Context, ctx.Key+":"+keyStr, valStr, expiration).Err()
}

func (ctx *StringKey[k, v]) Del(key k) error {
	keyStr, err := ctx.SerializeKey(key)
	if err != nil {
		return err
	}
	return ctx.Rds.Del(ctx.Context, ctx.Key+":"+keyStr).Err()
}

// get all keys that match the pattern, and return a map of key->value
func (ctx *StringKey[k, v]) GetAll(match string) (mapOut map[k]v, err error) {
	var (
		keys []string = []string{match}
		val  []byte
	)
	if keys, _, err = ctx.Scan(0, match, 1024*1024*1024); err != nil {
		return nil, err
	}
	mapOut = make(map[k]v, len(keys))
	var result error
	for _, key := range keys {
		if val, result = ctx.Rds.Get(ctx.Context, key).Bytes(); result != nil {
			err = result
			continue
		}
		//use default prefix to avoid confict of hash key
		//k is start with ctx.Key, remove it
		if len(ctx.Key) > 0 && (len(key) >= len(ctx.Key)) && key[:len(ctx.Key)] == ctx.Key {
			key = key[len(ctx.Key)+1:]
		}

		k, err := ctx.toKey([]byte(key))
		if err != nil {
			logger.Info().AnErr("GetAll: key unmarshal error:", err).Msgf("Key: %s", ctx.Key)
			continue
		}
		v, err := ctx.DeserializeToValue(val)
		if err != nil {
			logger.Info().AnErr("GetAll: value unmarshal error:", err).Msgf("Key: %s", ctx.Key)
			continue
		}
		mapOut[k] = v
	}
	return mapOut, err
}

// set each key value of _map to redis string type key value
func (ctx *StringKey[k, v]) SetAll(_map map[k]v) (err error) {
	//HSet each element of _map to redis
	pipe := ctx.Rds.Pipeline()
	for k, v := range _map {
		keyStr, err := ctx.SerializeKey(k)
		if err != nil {
			return err
		}
		valStr, err := ctx.SerializeValue(v)
		if err != nil {
			return err
		}

		pipe.Set(ctx.Context, ctx.Key+":"+keyStr, valStr, -1)
	}
	_, err = pipe.Exec(ctx.Context)
	return err
}
