package redisdb

import (
	"strings"

	"github.com/doptime/logger"
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
	keyScope := strings.ToLower(KeyScope(ctx.Key))
	hskey := StreamKey[k, v]{ctx.Duplicate(ctx.Key, ctx.RdsName)}
	IHashKey := HttpStreamKey[k, v](hskey)
	HttpStreamKeyMap.Set(keyScope+":"+ctx.RdsName, &IHashKey)
}
