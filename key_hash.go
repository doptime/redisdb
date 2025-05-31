package redisdb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/doptime/logger"
	"github.com/redis/go-redis/v9"
)

// HashKey represents a Redis hash key.
type HashKey[k comparable, v any] struct {
	RedisKey[k, v]
}

// NewHashKey creates a new HashKey with the given options.
func NewHashKey[k comparable, v any](ops ...Option) *HashKey[k, v] {
	ctx := &HashKey[k, v]{}
	op := append(ops, Opt)[0]
	if err := ctx.apply(keyTypeHashKey, op); err != nil {
		logger.Error().Err(err).Msg("data.New failed")
		return nil
	}
	ctx.getPrimaryKeyFieldIndex()
	return ctx
}
func (ctx *HashKey[K, V]) getPrimaryKeyFieldIndex() {
	ctx.PrimaryKeyFieldIndex = -1

	// 获取 V 的反射类型
	vType := reflect.TypeOf(new(V))
	if vType.Kind() == reflect.Ptr {
		// 如果 V 是指针类型，取其元素类型
		vType = vType.Elem()
	}

	// 解除多级指针/数组/切片包装
	for vType.Kind() == reflect.Ptr || vType.Kind() == reflect.Slice || vType.Kind() == reflect.Array {
		vType = vType.Elem()
	}

	// 确保 V 是结构体类型
	if vType.Kind() != reflect.Struct {
		return
	}

	// 获取 K 的反射类型
	kType := reflect.TypeOf(new(K))
	if kType.Kind() == reflect.Ptr {
		// 如果 K 是指针类型，取其元素类型
		kType = kType.Elem()
	}

	// 遍历字段查找匹配 K 类型的字段
	for i := 0; i < vType.NumField(); i++ {
		field := vType.Field(i)
		if field.Type.AssignableTo(kType) {
			ctx.PrimaryKeyFieldIndex = i
			return
		}
	}
}
func (ctx *HashKey[k, v]) ConcatKey(fields ...interface{}) *HashKey[k, v] {
	return &HashKey[k, v]{ctx.Duplicate(ConcatedKeys(ctx.Key, fields...), ctx.RdsName)}
}

func (ctx *HashKey[k, v]) HGet(field k) (value v, err error) {
	fieldStr, err := ctx.SerializeKey(field)
	if err != nil {
		return value, err
	}
	cmd := ctx.Rds.HGet(ctx.Context, ctx.Key, fieldStr)
	if err := cmd.Err(); err != nil {
		return value, err
	}
	data, err := cmd.Bytes()
	if err != nil {
		return value, err
	}
	return ctx.DeserializeValue(data)
}

// HSet accepts values in following formats:
//
//   - HSet("myhash", "key1", "value1", "key2", "value2")
//
//   - HSet("myhash", map[string]interface{}{"key1": "value1", "key2": "value2"})
func (ctx *HashKey[k, v]) HSet(values ...interface{}) (int64, error) {

	if kvMap, ok := values[0].(map[k]v); ok {
		return ctx.HMSet(kvMap)
	}
	// if Moder is not nil, apply modifiers to the values
	l := len(values)
	if l < 2 || l%2 != 0 || reflect.TypeOf(values).Kind() != reflect.Slice {
		return 0, fmt.Errorf("valid values must be in the format of [key1, value1, key2, value2, ...]")
	}
	for i, l := 0, len(values); i < l; i += 2 {
		_, keyok := values[i].(k)
		value, valueok := values[i+1].(v)
		if !keyok || !valueok {
			return 0, fmt.Errorf("valid values must be in the format of [key1, value1, key2, value2, ...]")
		}

		if ctx.UseModer {
			ApplyModifiers(&value)
		}
	}

	KeyValuesStrs, err := ctx.toKeyValueStrs(values...)
	if err != nil {
		return 0, err
	}
	return ctx.Rds.HSet(ctx.Context, ctx.Key, KeyValuesStrs).Result()
}
func (ctx *HashKey[k, v]) Save(value v) (int64, error) {
	//get first field of v , which type is k
	var fieldN = reflect.TypeOf(value).NumField()
	if fieldN == 0 {
		return ctx.HSet(value)
	}
	if ctx.UseModer {
		ApplyModifiers(&value)
	}
	if ctx.PrimaryKeyFieldIndex > 0 {
		field, ok := reflect.ValueOf(value).Field(ctx.PrimaryKeyFieldIndex).Interface().(k)
		if ok {
			return ctx.HSet(field, value)
		}
	}
	return 0, fmt.Errorf("no field of type k found in value")
}

func (ctx *HashKey[k, v]) HMSet(kvMap map[k]v) (int64, error) {
	// if Moder is not nil, apply modifiers to the values
	if ctx.UseModer {
		for _, value := range kvMap {
			ApplyModifiers(&value)
		}
	}

	KeyValuesStrs, err := ctx.toKeyValueStrs(kvMap)
	if err != nil {
		return 0, err
	}
	return ctx.Rds.HSet(ctx.Context, ctx.Key, KeyValuesStrs).Result()
}
func (ctx *HashKey[k, v]) HExists(field k) (bool, error) {
	fieldStr, err := ctx.SerializeKey(field)
	if err != nil {
		return false, err
	}
	return ctx.Rds.HExists(ctx.Context, ctx.Key, fieldStr).Result()
}

func (ctx *HashKey[k, v]) HGetAll() (map[k]v, error) {
	result, err := ctx.Rds.HGetAll(ctx.Context, ctx.Key).Result()
	if err != nil {
		return nil, err
	}
	mapOut := make(map[k]v)
	for k, v := range result {
		key, err := ctx.toKey([]byte(k))
		if err != nil {
			continue
		}
		value, err := ctx.DeserializeValue([]byte(v))
		if err != nil {
			continue
		}
		mapOut[key] = value
	}
	return mapOut, nil
}

func (ctx *HashKey[k, v]) HRandField(count int) (fields []k, err error) {
	var (
		cmd *redis.StringSliceCmd
	)
	if cmd = ctx.Rds.HRandField(ctx.Context, ctx.Key, count); cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return ctx.toKeys(cmd.Val())
}
func (ctx *HashKey[k, v]) HMGET(fields ...interface{}) (values []v, err error) {
	var (
		cmd          *redis.SliceCmd
		fieldsString []string
		rawValues    []string
	)
	if fieldsString, err = ctx.toKeyStrs(fields...); err != nil {
		return nil, err
	}
	if cmd = ctx.Rds.HMGet(ctx.Context, ctx.Key, fieldsString...); cmd.Err() != nil {
		return nil, cmd.Err()
	}
	rawValues = make([]string, len(cmd.Val()))
	for i, val := range cmd.Val() {
		if val == nil {
			continue
		}
		rawValues[i] = val.(string)
	}
	return ctx.DeserializeValues(rawValues)
}
func (ctx *HashKey[k, v]) HLen() (length int64, err error) {
	cmd := ctx.Rds.HLen(ctx.Context, ctx.Key)
	return cmd.Val(), cmd.Err()
}

func (ctx *HashKey[k, v]) HDel(fields ...k) (err error) {
	var (
		cmd       *redis.IntCmd
		fieldStrs []string
		bytes     []byte
	)
	if len(fields) == 0 {
		return nil
	}
	//if k is  string, then use HDEL directly
	if reflect.TypeOf(fields[0]).Kind() == reflect.String {
		fieldStrs = interface{}(fields).([]string)
	} else {
		//if k is not string, then marshal k to string
		fieldStrs = make([]string, len(fields))
		for i, field := range fields {
			if bytes, err = json.Marshal(field); err != nil {
				return err
			}
			fieldStrs[i] = string(bytes)
		}
	}
	cmd = ctx.Rds.HDel(ctx.Context, ctx.Key, fieldStrs...)
	return cmd.Err()
}

func (ctx *HashKey[k, v]) HKeys() ([]k, error) {
	result, err := ctx.Rds.HKeys(ctx.Context, ctx.Key).Result()
	if err != nil {
		return nil, err
	}
	keys := make([]k, len(result))
	for i, k := range result {
		key, err := ctx.toKey([]byte(k))
		if err != nil {
			continue
		}
		keys[i] = key
	}
	return keys, nil
}

func (ctx *HashKey[k, v]) HVals() ([]v, error) {
	result, err := ctx.Rds.HVals(ctx.Context, ctx.Key).Result()
	if err != nil {
		return nil, err
	}
	values := make([]v, len(result))
	for i, v := range result {
		value, err := ctx.DeserializeValue([]byte(v))
		if err != nil {
			continue
		}
		values[i] = value
	}
	return values, nil
}

func (ctx *HashKey[k, v]) HIncrBy(field k, increment int64) error {
	fieldStr, err := ctx.SerializeKey(field)
	if err != nil {
		return err
	}
	return ctx.Rds.HIncrBy(ctx.Context, ctx.Key, fieldStr, increment).Err()
}

func (ctx *HashKey[k, v]) HIncrByFloat(field k, increment float64) error {
	fieldStr, err := ctx.SerializeKey(field)
	if err != nil {
		return err
	}
	return ctx.Rds.HIncrByFloat(ctx.Context, ctx.Key, fieldStr, increment).Err()
}
func (ctx *HashKey[k, v]) HSetNX(field k, value v) error {
	fieldStr, err := ctx.SerializeKey(field)
	if err != nil {
		return err
	}
	valStr, err := ctx.SerializeValue(value)
	if err != nil {
		return err
	}
	return ctx.Rds.HSetNX(ctx.Context, ctx.Key, fieldStr, valStr).Err()
}

func (ctx *HashKey[k, v]) HScan(cursor uint64, match string, count int64) (keys []k, values []v, cursorRet uint64, err error) {
	var (
		cmd          *redis.ScanCmd
		keyValueStrs []string
	)
	if cmd = ctx.Rds.HScan(ctx.Context, ctx.Key, cursor, match, count); cmd.Err() != nil {
		return nil, nil, 0, cmd.Err()
	}
	keys = make([]k, 0)
	values = make([]v, 0)
	keyValueStrs, cursorRet, err = cmd.Result()
	for i := 0; i < len(keyValueStrs); i += 2 {
		k, err := ctx.toKey([]byte(keyValueStrs[i]))
		v, err1 := ctx.DeserializeValue([]byte(keyValueStrs[i+1]))
		if err != nil || err1 != nil {
			continue
		}
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values, cursorRet, err
}
func (ctx *HashKey[k, v]) HScanNoValues(cursor uint64, match string, count int64) (keys []k, cursorRet uint64, err error) {
	var (
		cmd      *redis.ScanCmd
		keysStrs []string
	)
	if cmd = ctx.Rds.HScanNoValues(ctx.Context, ctx.Key, cursor, match, count); cmd.Err() != nil {
		return nil, 0, cmd.Err()
	}
	keysStrs, cursorRet, err = cmd.Result()
	if err == nil {
		keys, err = ctx.toKeys(keysStrs)
	}
	return keys, cursorRet, err
}
