package redisdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/doptime/config/cfgredis"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type CtxInterface interface {
	GetUseModer() bool
	ValidDataKey() error
	UnmarshalValue(msgpack []byte) (rets interface{}, err error)
}

var RediskeyForWeb cmap.ConcurrentMap[string, CtxInterface] = cmap.New[CtxInterface]()
var nonKey = NewRedisKey[string, interface{}]()

func CtxWithValueSchemaChecked(key, keyType string, RedisDataSource string, msgpackData []byte) (db *RedisKey[string, interface{}], value interface{}, err error) {
	useModer, originalKey := false, key
	originalKey = strings.SplitN(key, "@", 2)[0]
	originalKey = strings.SplitN(originalKey, ":", 2)[0]

	hashInterface, exists := RediskeyForWeb.Get(originalKey + ":" + RedisDataSource)
	if hashInterface != nil && exists {
		useModer = hashInterface.GetUseModer()
	}
	value, err = hashInterface.UnmarshalValue(msgpackData)

	if err != nil {
		return nil, nil, err
	}

	if disallowed, found := DisAllowedDataKeyNames[key]; found && disallowed {
		return nil, nil, fmt.Errorf("key name is disallowed: " + key)
	}
	ctx := RedisKey[string, interface{}]{context.Background(), RedisDataSource, nil, key, keyType,
		nonKey.SerializeKey, nonKey.SerializeValue, nonKey.DeserializeValue, nonKey.DeserializeValues, nonKey.autoFiller, nonKey.validator,
		useModer, -1}
	if ctx.Rds, exists = cfgredis.Servers.Get(RedisDataSource); !exists {
		return nil, nil, fmt.Errorf("rds item unconfigured: " + RedisDataSource)
	}
	return &ctx, value, nil
}

func HashCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *HashKey[string, interface{}], value interface{}, err error) {
	var ctx *RedisKey[string, interface{}]
	ctx, value, err = CtxWithValueSchemaChecked(key, "hash", RedisDataSource, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &HashKey[string, interface{}]{*ctx}, value, nil
}
func StringCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *StringKey[string, interface{}], value interface{}, err error) {
	var ctx *RedisKey[string, interface{}]
	ctx, value, err = CtxWithValueSchemaChecked(key, "string", RedisDataSource, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &StringKey[string, interface{}]{*ctx}, value, nil
}
func ListCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *ListKey[interface{}], value interface{}, err error) {
	var ctx *RedisKey[string, interface{}]
	ctx, value, err = CtxWithValueSchemaChecked(key, "list", RedisDataSource, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &ListKey[interface{}]{*ctx}, value, nil
}

func (ctx *RedisKey[k, v]) ValidDataKey() error {
	if disallowed, found := DisAllowedDataKeyNames[ctx.Key]; found && disallowed {
		return fmt.Errorf("key name is disallowed: " + ctx.Key)
	}
	if _, ok := cfgredis.Servers.Get(ctx.RdsName); !ok {
		return fmt.Errorf("rds item unconfigured: " + ctx.RdsName)
	}
	return nil
}
