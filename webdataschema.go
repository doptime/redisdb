package redisdb

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

var WebDataDocsMap cmap.ConcurrentMap[string, *WebDataDocs] = cmap.New[*WebDataDocs]()

var SynWebDataRunOnce = sync.Mutex{}
var KeyWebDataDocs = NewHashKey[string, *WebDataDocs](WithKey("Docs:Data"))

// func initializeFields(value reflect.Value) (ret interface{}) {
// 	switch value.Kind() {
// 	case reflect.String:
// 		ret = ""
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 		ret = 0
// 	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// 		ret = 0
// 	case reflect.Float32, reflect.Float64:
// 		ret = 0.0
// 	case reflect.Bool:
// 		ret = false
// 	case reflect.Slice:
// 		if value.IsNil() {
// 			value.Set(reflect.MakeSlice(value.Type(), 0, 0))
// 		}
// 		elemType := value.Type().Elem()
// 		if elemType.Kind() == reflect.Invalid {
// 			return value.IsNil()
// 		}
// 		elementValue := initializeFields(reflect.New(elemType).Elem())
// 		newSlice := reflect.Append(value, reflect.ValueOf(elementValue))
// 		return newSlice.Interface()

//		case reflect.Struct:
//			for i := 0; i < value.NumField(); i++ {
//				field := value.Field(i)
//				if field.CanSet() {
//					fieldValue := initializeFields(field)
//					field.Set(reflect.ValueOf(fieldValue))
//				}
//			}
//			return value.Interface()
//		case reflect.Map:
//			if value.IsNil() {
//				value.Set(reflect.MakeMap(value.Type()))
//			}
//			keyType := value.Type().Key()
//			valType := value.Type().Elem()
//			if keyType.Kind() != reflect.Invalid && valType.Kind() != reflect.Invalid {
//				mapKey := initializeFields(reflect.New(keyType).Elem())
//				mapValue := initializeFields(reflect.New(valType).Elem())
//				value.SetMapIndex(reflect.ValueOf(mapKey), reflect.ValueOf(mapValue))
//			}
//			return value.Interface()
//		case reflect.Ptr:
//			if value.IsNil() {
//				value.Set(reflect.New(value.Type().Elem()))
//			}
//			initializeFields(value.Elem())
//			return value.Interface()
//		case reflect.Interface:
//			if !value.IsNil() {
//				elem := value.Elem()
//				return initializeFields(elem)
//			}
//		}
//		return value.Interface()
//	}
func initializeType(t reflect.Type) reflect.Value {
	switch t.Kind() {
	case reflect.Ptr:
		elemValue := initializeType(t.Elem())
		ptrValue := reflect.New(t.Elem())
		ptrValue.Elem().Set(elemValue)
		return ptrValue
	case reflect.Slice:
		elemType := t.Elem()
		elemValue := initializeType(elemType)
		sliceValue := reflect.MakeSlice(t, 1, 1)
		sliceValue.Index(0).Set(elemValue)
		return sliceValue
	case reflect.Array:
		elemType := t.Elem()
		elemValue := initializeType(elemType)
		arrayValue := reflect.New(t).Elem()
		for i := 0; i < t.Len(); i++ {
			arrayValue.Index(i).Set(elemValue)
		}
		return arrayValue
	case reflect.Map:
		keyType := t.Key()
		elemType := t.Elem()
		keyValue := initializeType(keyType)
		elemValue := initializeType(elemType)
		mapValue := reflect.MakeMap(t)
		mapValue.SetMapIndex(keyValue, elemValue)
		return mapValue
	case reflect.Struct:
		structValue := reflect.New(t).Elem()
		for i := 0; i < t.NumField(); i++ {
			field := structValue.Field(i)
			fieldType := t.Field(i)
			if field.CanSet() && fieldType.PkgPath == "" { // 确保字段是可导出的
				fieldValue := initializeType(fieldType.Type)
				field.Set(fieldValue)
			}
		}
		return structValue
	default:
		return reflect.Zero(t)
	}
}

type WebDataDocs struct {
	KeyName string
	// string, hash, list, set, zset, stream
	KeyType         string
	UpdateAt        int64
	CreateFromLocal bool `msgpack:"-"`
	Instance        interface{}
}

func (ctx *RedisKey[k, v]) RegisterWebData() {
	var validRdsKeyTypes = map[string]bool{"string": true, "list": true, "set": true, "hash": true, "zset": true, "stream": true}
	if _, ok := validRdsKeyTypes[ctx.KeyType]; !ok {
		return
	}

	// check if type of v can be instantiated
	vType := reflect.TypeOf((*v)(nil)).Elem()
	if vType.Kind() == reflect.Invalid {
		fmt.Println("vType is not valid, vType: ", vType)
		return
	}

	rootKey := strings.Split(ctx.Key, ":")[0]
	obj := initializeType(vType).Interface()
	dataSchema := &WebDataDocs{
		KeyName:         rootKey,
		KeyType:         ctx.KeyType,
		Instance:        obj,
		UpdateAt:        time.Now().Unix(),
		CreateFromLocal: true,
	}
	WebDataDocsMap.Set(ctx.Key, dataSchema)
	if SynWebDataRunOnce.TryLock() {
		go syncWebDataToRedis()
	}
}

func syncWebDataToRedis() {
	//wait arrival of other schema to be store in map
	time.Sleep(time.Second)
	for {
		now := time.Now().Unix()
		//only update local defined data to redis
		WebDataDocsMap.IterCb(func(key string, value *WebDataDocs) {
			if value.CreateFromLocal {
				value.UpdateAt = now
			}
		})
		if WebDataDocsMap.Count() > 0 {
			KeyWebDataDocs.HSet(WebDataDocsMap.Items())
		}

		//for the purpose of checking the data schema
		//copy all data schema to local ,but do not cover the local data
		if vals, err := KeyWebDataDocs.HGetAll(); err == nil {
			for k, v := range vals {
				//if defined in non local, allow to cover the definition
				//that is ,only local defined data schema can not be covered
				if v, ok := WebDataDocsMap.Get(k); ok && v.CreateFromLocal {
					continue
				}
				WebDataDocsMap.Set(k, v)
			}
		}
		//sleep 10 min to save next time
		time.Sleep(time.Minute * 10)
	}
}
