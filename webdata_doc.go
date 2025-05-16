package redisdb

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type WebDataSchema struct {
	KeyName string
	// string, hash, list, set, zset, stream
	KeyType     string
	UpdateAt    int64
	JSDoc       string
	TSInterface string
	Instance    interface{}

	CreateFromLocal bool `msgpack:"-"`
}

var KeyWebDataSchema = NewHashKey[string, *WebDataSchema](Opt.Key("Docs:WebDataSchema"))

var WebDataSchemaMap = cmap.New[*WebDataSchema]()

func (ctx *RedisKey[k, v]) RegisterWebData() {
	var validRdsKeyTypes = map[string]bool{"string": true, "list": true, "set": true, "hash": true, "zset": true, "stream": true}
	if _, ok := validRdsKeyTypes[ctx.KeyType]; !ok {
		return
	}

	// check if type of v can be instantiated
	_v := (*v)(nil)
	vType := reflect.TypeOf(_v).Elem()
	if vType.Kind() == reflect.Invalid {
		fmt.Println("vType is not valid, vType: ", vType)
		return
	}

	rootKey := strings.Split(ctx.Key, ":")[0]
	obj := initializeType(vType).Interface()
	jsdoc, _ := GenerateAllJSDocTypeDefs(_v)
	typescriptInterface, _ := GoTypeToTypeScriptInterface(_v)
	dataSchema := &WebDataSchema{
		KeyName:         rootKey,
		KeyType:         ctx.KeyType,
		Instance:        obj,
		JSDoc:           jsdoc,
		TSInterface:     typescriptInterface,
		UpdateAt:        time.Now().Unix(),
		CreateFromLocal: true,
	}
	WebDataSchemaMap.Set(rootKey, dataSchema)
}
func init() {
	go syncWebDataToRedis()
}

func syncWebDataToRedis() {
	//wait arrival of other schema to be store in map
	time.Sleep(time.Second)
	for WebDataSchemaMap.Count() > 0 {
		now := time.Now().Unix()
		//only update local defined data to redis
		WebDataSchemaMap.IterCb(func(key string, value *WebDataSchema) {
			if value.CreateFromLocal {
				value.UpdateAt = now
			}
		})
		KeyWebDataSchema.HSet(WebDataSchemaMap.Items())

		//for the purpose of checking the data schema
		//copy all data schema to local ,but do not cover the local data
		if vals, err := KeyWebDataSchema.HGetAll(); err == nil {
			for k, v := range vals {
				if v, ok := WebDataSchemaMap.Get(k); ok && v.CreateFromLocal {
					continue
				}
				WebDataSchemaMap.Set(k, v)
			}
		}
		//sleep 10 min to save next time
		time.Sleep(time.Minute * 10)
	}
}
