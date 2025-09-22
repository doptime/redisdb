package redisdb

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/doptime/config/cfgredis"
	"github.com/doptime/logger"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

type RedisKey[k comparable, v any] struct {
	Context context.Context
	RdsName string
	Rds     *redis.Client

	Key     string
	KeyType KeyType

	SerializeKey         func(value interface{}) (msgpack string, err error)
	SerializeValue       func(value interface{}) (msgpack string, err error)
	DeserializeValue     func(msgpack []byte) (value v, err error)
	DeserializeValues    func(msgpacks []string) (values []v, err error)
	timestampFiller      func(in v) error
	Validator            func(in v) error
	UseModer             bool
	PrimaryKeyFieldIndex int
}

func (ctx *RedisKey[k, v]) V(value v) (ret v) {
	ret = value
	if ctx.UseModer {
		ApplyModifiers(&ret)
	}
	return ret
}

func (ctx *RedisKey[k, v]) Duplicate(newKey, RdsSourceName string) (newCtx RedisKey[k, v]) {
	return RedisKey[k, v]{ctx.Context, RdsSourceName, ctx.Rds, newKey, ctx.KeyType,
		ctx.SerializeKey, ctx.SerializeValue, ctx.DeserializeValue, ctx.DeserializeValues, ctx.timestampFiller, ctx.Validator,
		ctx.UseModer, ctx.PrimaryKeyFieldIndex}
}
func (ctx *RedisKey[k, v]) InitFunc() {
	ctx.Context = context.Background()
	ctx.SerializeKey = ctx.getSerializeFun(reflect.TypeOf((*k)(nil)).Elem().Kind())
	ctx.SerializeValue = ctx.getSerializeFun(reflect.TypeOf((*v)(nil)).Elem().Kind())
	ctx.DeserializeValue = ctx.getDeserializetoValueFunc()
	ctx.DeserializeValues = ctx.toValuesFunc()
	ctx.timestampFiller = ctx.NewTimestampFiller()
	ctx.Validator = ctx.NewValidator()
}

func NewRedisKey[k comparable, v any](ops ...Option) *RedisKey[k, v] {
	ctx := &RedisKey[k, v]{}
	if err := ctx.applyOptionsAndCheck(keyTypeNonKey, ops...); err != nil {
		logger.Error().Err(err).Msg("redisdb.NewRedisKey failed")
		return nil
	}

	ctx.InitFunc()
	return ctx
}
func (ctx *RedisKey[k, v]) Time() (tm time.Time, err error) {
	cmd := ctx.Rds.Time(ctx.Context)
	return cmd.Result()
}
func (ctx *RedisKey[k, v]) GetUseModer() bool {
	return ctx.UseModer
}

// sacn key by pattern
func (ctx *RedisKey[k, v]) Scan(cursorOld uint64, match string, count int64) (keys []string, cursorNew uint64, err error) {
	var (
		cmd   *redis.ScanCmd
		_keys []string
	)
	//scan all keys
	for {

		if cmd = ctx.Rds.Scan(ctx.Context, cursorOld, match, count); cmd.Err() != nil {
			return nil, 0, cmd.Err()
		}
		if _keys, cursorNew, err = cmd.Result(); err != nil {
			return nil, 0, err
		}
		keys = append(keys, _keys...)
		if cursorNew == 0 {
			break
		}
	}
	return keys, cursorNew, nil
}

type KeyType string

const (
	keyTypeNonKey    KeyType = "nonkey"
	keyTypeStringKey KeyType = "string"
	keyTypeHashKey   KeyType = "hash"
	keyTypeListKey   KeyType = "list"
	keyTypeSetKey    KeyType = "set"
	keyTypeZSetKey   KeyType = "zset"
	keyTypeStreamKey KeyType = "stream"
)

func IsValidKeyType(keyType string) bool {
	switch keyType {
	case string(keyTypeNonKey), string(keyTypeStringKey), string(keyTypeHashKey), string(keyTypeListKey), string(keyTypeSetKey), string(keyTypeZSetKey), string(keyTypeStreamKey):
		return true
	default:
		return false
	}
}

func (ctx *RedisKey[k, v]) applyOptionsAndCheck(keyType KeyType, opts ...Option) (err error) {
	//make default Option take effect. "default" as default redis name
	OptionDefault := Option{
		RedisKey:        "",
		KeyType:         keyType,
		RedisDataSource: "default",
		HttpAccess:      false,
		Modifiers:       map[string]ModifierFunc{},
	}
	OptionDefault.RedisKey, _ = GetValidDataKeyName((*v)(nil))
	opts = append([]Option{OptionDefault}, opts...)
	for _, opt := range opts {
		if len(opt.KeyType) > 0 {
			ctx.KeyType = opt.KeyType
		}
		if len(opt.RedisKey) > 0 {
			ctx.Key = opt.RedisKey
		}
		if len(opt.RedisDataSource) > 0 {
			ctx.RdsName = opt.RedisDataSource
		}
		if len(opt.Modifiers) > 0 {
			ctx.UseModer = RegisterStructModifiers(opt.Modifiers, reflect.TypeOf((*v)(nil)).Elem())
		}

		// don't register web data if it fully prepared
		if opt.HttpAccess && ctx.Key != "" {
			ctx.RegisterWebData()
			RediskeyForWeb.Set(ctx.Key+":"+ctx.RdsName, ctx)
		}
	}
	//check if  options are valid
	if len(ctx.Key) == 0 {
		return fmt.Errorf("invalid data.Ctx Key name")
	}
	var exists bool
	if ctx.Rds, exists = cfgredis.Servers.Get(ctx.RdsName); !exists {
		return fmt.Errorf("rds item unconfigured: " + ctx.RdsName)
	}

	return nil
}

func (ctx *RedisKey[k, v]) toKeyValueStrs(keyValue ...interface{}) (keyValStrs []string, err error) {
	var (
		key              k
		value            v
		strkey, strvalue string
	)
	if len(keyValue) == 0 {
		return keyValStrs, fmt.Errorf("key value is nil")
	}
	// if key value is a map, convert it to key value slice
	if kvMap, ok := keyValue[0].(map[k]v); ok {
		for key, value := range kvMap {
			if strkey, err = ctx.SerializeKey(key); err != nil {
				return nil, err
			}
			if strvalue, err = ctx.SerializeValue(value); err != nil {
				return nil, err
			}
			keyValStrs = append(keyValStrs, strkey, strvalue)
		}
	} else if l := len(keyValue); l%2 == 0 {
		for i := 0; i < l; i += 2 {
			if key, ok = keyValue[i].(k); !ok {
				logger.Error().Any(" key must be of type k", key).Any("raw", keyValue[i+1]).Send()
				return nil, fmt.Errorf("invalid key type in toKeyValueStrs")
			}
			if value, ok = keyValue[i+1].(v); !ok {
				logger.Error().Any(" value must be of type v", value).Any("raw", keyValue[i+1]).Send()
				return nil, fmt.Errorf("invalid value type in toKeyValueStrs")
			}
			if strkey, err = ctx.SerializeKey(key); err != nil {
				return nil, err
			}
			if strvalue, err = ctx.SerializeValue(value); err != nil {
				return nil, err
			}
			keyValStrs = append(keyValStrs, strkey, strvalue)
		}
	} else {
		return nil, fmt.Errorf("invalid type key value while converting to strings")
	}
	return keyValStrs, nil
}
func (ctx *RedisKey[k, v]) UnmarshalValue(msgpackBytes []byte) (rets interface{}, err error) {

	if len(msgpackBytes) == 0 {
		return nil, fmt.Errorf("msgpackBytes is empty")
	}

	var vInstance v
	if err = msgpack.Unmarshal(msgpackBytes, &vInstance); err != nil {
		return nil, err
	}

	// auto fill field UpdateAt and CreateAt
	err = ctx.TimestampFill(vInstance)
	if err == nil {
		err = ctx.Validate(vInstance)
	}

	return vInstance, err
}

func (ctx *RedisKey[k, v]) MsgpackUnmarshalKeyValues(msgpack []byte) (rets interface{}, err error) {
	return nil, nil
}

func (ctx *RedisKey[k, v]) Keys() (out []k, err error) {
	var (
		cmd  *redis.StringSliceCmd
		keys []string
	)
	cmd = ctx.Rds.Keys(ctx.Context, ctx.Key+":*")
	if keys, err = cmd.Result(); err != nil {
		return nil, err
	}
	return ctx.toKeys(keys)
}

// for the reason of protection, both ctx.Key & Key are required. the avoid set Hash table to the wrong type , and thus leading to data loss.
func (ctx *RedisKey[k, v]) Set(key k, param v, expiration time.Duration) (err error) {
	var (
		keyStr string
		valStr string
	)
	if keyStr, err = ctx.SerializeKey(key); err != nil {
		return err
	}
	if valStr, err = ctx.SerializeValue(param); err != nil {
		return err
	} else {
		status := ctx.Rds.Set(ctx.Context, ctx.Key+":"+keyStr, valStr, expiration)
		return status.Err()
	}
}

// for the reason of protection, both ctx.Key & Key are required. the avoid set Hash table to the wrong type , and thus leading to data loss.
func (ctx *RedisKey[k, v]) Del(key k) (err error) {
	var (
		keyStr string
	)
	if keyStr, err = ctx.SerializeKey(key); err != nil {
		return err
	}
	status := ctx.Rds.Del(ctx.Context, ctx.Key+":"+keyStr)
	return status.Err()
}
