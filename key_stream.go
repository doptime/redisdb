package redisdb

import (
	"github.com/doptime/logger"
)

type StreamKey[k comparable, v any] struct {
	RedisKey[k, v]
}

func NewStreamKey[k comparable, v any](ops ...Option) *StreamKey[k, v] {
	ctx := &StreamKey[k, v]{RedisKey: RedisKey[k, v]{KeyType: keyTypeStreamKey}}
	if err := ctx.applyOptionsAndCheck(keyTypeStreamKey, ops...); err != nil {
		logger.Error().Err(err).Msg("redisdb.NewStreamKey failed")
		return nil
	}
	ctx.InitFunc()
	return ctx
}

func (ctx *StreamKey[k, v]) ConcatKey(fields ...interface{}) *StreamKey[k, v] {
	return &StreamKey[k, v]{ctx.Duplicate(ConcatedKeys(ctx.Key, fields...), ctx.RdsName)}
}
