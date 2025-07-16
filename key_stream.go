package redisdb

import (
	"github.com/doptime/logger"
)

type StreamKey[k comparable, v any] struct {
	RedisKey[k, v]
}

func NewStreamKey[k comparable, v any](ops ...Option) *StreamKey[k, v] {
	ctx := &StreamKey[k, v]{}
	for _, op := range ops {
		if err := ctx.applyOption(keyTypeStreamKey, op); err != nil {
			logger.Error().Err(err).Msg("data.New failed")
			return nil
		}
	}
	ctx.InitFunc()
	return ctx
}

func (ctx *StreamKey[k, v]) ConcatKey(fields ...interface{}) *StreamKey[k, v] {
	return &StreamKey[k, v]{ctx.Duplicate(ConcatedKeys(ctx.Key, fields...), ctx.RdsName)}
}
