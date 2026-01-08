package redisdb

import (
	"fmt"

	"github.com/doptime/config/cfgredis"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type CtxInterface interface {
	GetKeyType() KeyType
	GetUseModer() bool
	ValidDataKey() error
	DeserializeToInterface(msgpack []byte) (rets interface{}, err error)
	DeserializeToInterfaceSlice(msgpacks []string) (rets []interface{}, err error)
	TimestampFiller(in interface{}) (err error)
	CloneToRedisKey(newKey, RdsSourceName string) (newCtx *RedisKey[string, interface{}])
}

var rediskeyInterfaceForWebVisit cmap.ConcurrentMap[string, CtxInterface] = cmap.New[CtxInterface]()

func (ctx *RedisKey[k, v]) ValidDataKey() error {
	_keyscope := KeyScope(ctx.Key)
	if disallowed, found := DisAllowedDataKeyNames[_keyscope]; found && disallowed {
		return fmt.Errorf("key name is disallowed: " + ctx.Key)
	}
	if _, ok := cfgredis.Servers.Get(ctx.RdsName); !ok {
		return fmt.Errorf("rds item unconfigured: " + ctx.RdsName)
	}
	return nil
}
