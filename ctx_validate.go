package redisdb

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

func (ctx *RedisKey[k, v]) NewValidator() func(in v) (err error) {
	var (
		validate *validator.Validate = validator.New()
	)
	vType := reflect.TypeOf((*v)(nil)).Elem()
	for vType.Kind() == reflect.Ptr {
		vType = vType.Elem()
	}
	hasValidateTag := func(t reflect.Type) bool {
		for i := 0; i < t.NumField(); i++ {
			if tag := t.Field(i).Tag.Get("validate"); tag != "" {
				return true
			}
		}
		return false
	}
	hasValidTag := vType.Kind() == reflect.Struct && hasValidateTag(vType)
	return func(in v) (err error) {
		if !hasValidTag {
			return nil
		}
		return validate.Struct(in)
	}
}
func (ctx *RedisKey[k, v]) Validate(in interface{}) (err error) {
	if ctx.validator == nil {
		return nil
	}
	_in, ok := in.(v)
	if !ok {
		return fmt.Errorf("invalid type for AutoFill: %T", in)
	}
	return ctx.validator(_in)
}
