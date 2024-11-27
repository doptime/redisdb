package redisdb

import (
	"reflect"
	"strconv"

	"github.com/vmihailenco/msgpack/v5"
)

func (ctx *RedisKey[k, v]) getSerializeFun(typeofv reflect.Kind) func(val interface{}) (valueStr string, err error) {
	//var typeofv = reflect.TypeOf((*v)(nil)).Elem().Kind()
	switch typeofv {
	//type string
	case reflect.String:
		return func(value interface{}) (valueStr string, err error) {
			return value.(string), nil
		}
		//type int
	case reflect.Int:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatInt(int64(value.(int)), 10), nil
		}

	case reflect.Int8:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatInt(int64(value.(int8)), 10), nil
		}
	case reflect.Int16:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatInt(int64(value.(int16)), 10), nil
		}
	case reflect.Int32:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatInt(int64(value.(int32)), 10), nil
		}
	case reflect.Int64:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatInt(value.(int64), 10), nil
		}

		//case uint
	case reflect.Uint:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatUint(uint64(value.(uint)), 10), nil
		}

	case reflect.Uint8:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatUint(uint64(value.(uint8)), 10), nil
		}
	case reflect.Uint16:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatUint(uint64(value.(uint16)), 10), nil
		}
	case reflect.Uint32:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatUint(uint64(value.(uint32)), 10), nil
		}
	case reflect.Uint64:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatUint(value.(uint64), 10), nil
		}

		//case float
	case reflect.Float32:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatFloat(float64(value.(float32)), 'f', -1, 32), nil
		}

	case reflect.Float64:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatFloat(value.(float64), 'f', -1, 64), nil
		}

	case reflect.Bool:
		return func(value interface{}) (valueStr string, err error) {
			return strconv.FormatBool(value.(bool)), nil
		}
	default:
		return func(value interface{}) (valueStr string, err error) {
			bytes, err := msgpack.Marshal(value)
			if err == nil {
				return string(bytes), nil
			}
			return valueStr, err
		}
	}
}

func (ctx *RedisKey[k, v]) toKeyStrs(keys ...interface{}) (KeyStrs []string, err error) {
	var keyStr string
	for _, key := range keys {
		if keyStr, err = ctx.SerializeKey(key); err != nil {
			return nil, err
		}
		KeyStrs = append(KeyStrs, keyStr)
	}
	return KeyStrs, nil
}

func (ctx *RedisKey[k, v]) toValueStrsSlice(values ...v) (ValueStrs []interface{}, err error) {
	var valueStr string
	for _, value := range values {
		if valueStr, err = ctx.SerializeValue(value); err != nil {
			return nil, err
		}
		ValueStrs = append(ValueStrs, valueStr)
	}
	return ValueStrs, nil
}
