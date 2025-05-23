package redisdb

import (
	"github.com/doptime/logger"
)

type StreamKey[k comparable, v any] struct {
	RedisKey[k, v]
}

func NewStreamKey[k comparable, v any](ops ...Option) *StreamKey[k, v] {
	ctx := &StreamKey[k, v]{}
	op := append(ops, Opt)[0]
	if err := ctx.apply(keyTypeZSetKey, op); err != nil {
		logger.Error().Err(err).Msg("data.New failed")
		return nil
	}
	return ctx
}

func (ctx *StreamKey[k, v]) ConcatKey(fields ...interface{}) *StreamKey[k, v] {
	return &StreamKey[k, v]{ctx.Duplicate(ConcatedKeys(ctx.Key, fields...), ctx.RdsName)}
}
