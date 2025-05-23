package redisdb

import (
	"fmt"
	"reflect"
	"time"
)

func (ctx *RedisKey[k, v]) TimestampFill(in interface{}) (err error) {
	if ctx.timestampFiller == nil {
		return nil
	}
	_in, ok := in.(v)
	if !ok {
		return fmt.Errorf("invalid type for AutoFill: %T", in)
	}
	return ctx.timestampFiller(_in)
}
func (ctx *RedisKey[k, v]) NewTimestampFiller() func(in v) (err error) {
	vType := reflect.TypeOf((*v)(nil)).Elem()

	// 剥离指针直到非指针类型
	for vType.Kind() == reflect.Ptr {
		vType = vType.Elem()
	}

	// 如果不是结构体，跳过
	if vType.Kind() != reflect.Struct {
		return nil
	}

	createAtIndex := -1
	updateAtIndex := -1

	for i := 0; i < vType.NumField(); i++ {
		field := vType.Field(i)
		if field.Type == reflect.TypeOf(time.Time{}) {
			if field.Name == "CreatedAt" {
				createAtIndex = i
			} else if field.Name == "UpdatedAt" {
				updateAtIndex = i
			}
		}
	}

	//return nil function if no field found
	if createAtIndex == -1 && updateAtIndex == -1 {
		return nil
	}

	// 创建闭包，该闭包在每次调用时直接使用预先计算的索引
	return func(in v) (err error) {
		// 如果没有需要自动填充的字段，直接返回 nil
		if createAtIndex == -1 && updateAtIndex == -1 {
			return nil
		}
		val := reflect.ValueOf(&in).Elem()
		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		now := time.Now().UTC()
		if createAtIndex != -1 {
			field := val.Field(createAtIndex)
			if field.CanSet() && field.Type() == reflect.TypeOf(time.Time{}) {
				if field.Interface().(time.Time).IsZero() {
					field.Set(reflect.ValueOf(now))
				}
			}
		}
		if updateAtIndex != -1 {
			field := val.Field(updateAtIndex)
			if field.CanSet() && field.Type() == reflect.TypeOf(time.Time{}) {
				field.Set(reflect.ValueOf(now))
			}
		}
		return nil
	}
}
