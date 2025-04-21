package redisdb

// type CtxInterface interface {
// 	// MsgpackUnmarshalValue(msgpack []byte) (rets interface{}, err error)
// 	// MsgpackUnmarshalKeyValues(msgpack []byte) (rets interface{}, err error)
// 	CheckDataSchema(msgpackBytes []byte) (val interface{}, err error)
// 	GetUseModer() bool
// 	Validate() error
// }

// var hKeyMap cmap.ConcurrentMap[string, CtxInterface] = cmap.New[CtxInterface]()
// var nonKey = NewRedisKey[string, interface{}]()

// func CtxWithValueSchemaChecked(key, keyType string, RedisDataSource string, msgpackData []byte) (db *RedisKey[string, interface{}], value interface{}, err error) {
// 	useModer, originalKey := false, key
// 	originalKey = strings.SplitN(key, "@", 2)[0]
// 	originalKey = strings.SplitN(originalKey, ":", 2)[0]

// 	hashInterface, exists := hKeyMap.Get(originalKey + ":" + RedisDataSource)
// 	if hashInterface != nil && exists {
// 		useModer = hashInterface.GetUseModer()
// 		if msgpackData != nil {
// 			value, err = hashInterface.CheckDataSchema(msgpackData)
// 		}
// 	} else if msgpackData != nil {
// 		err = msgpack.Unmarshal(msgpackData, &value)
// 	}

// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	if disallowed, found := DisAllowedDataKeyNames[key]; found && disallowed {
// 		return nil, nil, fmt.Errorf("key name is disallowed: " + key)
// 	}
// 	ctx := RedisKey[string, interface{}]{context.Background(), RedisDataSource, nil, key, keyType,
// 		nonKey.SerializeKey, nonKey.SerializeValue, nonKey.DeserializeValue, nonKey.DeserializeValues, useModer, -1}
// 	if ctx.Rds, exists = cfgredis.Servers.Get(RedisDataSource); !exists {
// 		return nil, nil, fmt.Errorf("rds item unconfigured: " + RedisDataSource)
// 	}
// 	return &ctx, value, nil
// }

// func HashCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *HashKey[string, interface{}], value interface{}, err error) {
// 	var ctx *RedisKey[string, interface{}]
// 	ctx, value, err = CtxWithValueSchemaChecked(key, "hash", RedisDataSource, msgpackData)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	return &HashKey[string, interface{}]{*ctx}, value, nil
// }
// func StringCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *StringKey[string, interface{}], value interface{}, err error) {
// 	var ctx *RedisKey[string, interface{}]
// 	ctx, value, err = CtxWithValueSchemaChecked(key, "string", RedisDataSource, msgpackData)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	return &StringKey[string, interface{}]{*ctx}, value, nil
// }

// func (ctx *RedisKey[k, v]) Validate() error {
// 	if disallowed, found := DisAllowedDataKeyNames[ctx.Key]; found && disallowed {
// 		return fmt.Errorf("key name is disallowed: " + ctx.Key)
// 	}
// 	if _, ok := cfgredis.Servers.Get(ctx.RdsName); !ok {
// 		return fmt.Errorf("rds item unconfigured: " + ctx.RdsName)
// 	}
// 	return nil
// }

// func (ctx *RedisKey[k, v]) CheckDataSchema(msgpackBytes []byte) (val interface{}, err error) {
// 	if len(msgpackBytes) == 0 {
// 		return nil, fmt.Errorf("msgpackBytes is empty")
// 	}

// 	var vInstance v

// 	if err = msgpack.Unmarshal(msgpackBytes, &vInstance); err != nil {
// 		return nil, err
// 	}

// 	return vInstance, nil
// }
