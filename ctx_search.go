package redisdb

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strings"
)

// SearchKey 专为 AI 场景设计，支持文本检索和向量检索 (RAG)
type SearchKey[k comparable, v any] struct {
	*HashKey[k, v] // 继承 HashKey 的能力
	IndexName      string
	Prefix         string
}

// NewSearchKey 创建一个支持 RediSearch 的 Key Context
func NewSearchKey[k comparable, v any](indexName string, ops ...Option) *SearchKey[k, v] {
	// 强制使用 HashKey 类型，因为 RediSearch 主要基于 Hash
	ops = append(ops, Option{KeyType: keyTypeHashKey})

	// 初始化基础 HashKey
	baseKey := NewHashKey[k, v](ops...)
	if baseKey == nil {
		return nil
	}

	// IndexName 默认为 Key 的前缀，或者用户指定
	if indexName == "" {
		indexName = "idx:" + baseKey.Key
	}

	sk := &SearchKey[k, v]{
		HashKey:   baseKey,
		IndexName: indexName,
		Prefix:    baseKey.Key,
	}

	// 自动检查并创建索引 (Schema 自省)
	// 注意：在 AI 场景下自动处理更友好，防止因为忘记建索引导致无法搜索
	if err := sk.EnsureIndex(); err != nil {
		// 这里记录错误但不中断，因为可能是连接问题
		fmt.Printf("Warning: Failed to ensure index %s: %v\n", indexName, err)
	}

	return sk
}

// Put 这是一个对 AI 友好的别名，本质是 HSet，但会自动将 Struct 拆解为 Flat Hash
func (ctx *SearchKey[k, v]) Put(id k, doc v) error {
	// 将结构体转换为 map[string]interface{} 以便存储为独立的 Hash 字段
	// 这样 RediSearch 才能索引到具体的字段
	flatFields, err := structToFlatMap(doc)
	if err != nil {
		return err
	}
	// 调用底层的 HMSet
	keyStr, _ := ctx.SerializeKey(id)
	fullKey := ctx.Key + ":" + keyStr

	return ctx.Rds.HSet(ctx.Context, fullKey, flatFields).Err()
}

// Search 执行文本搜索
// Query 示例: "hello world @age:[10 20]"
func (ctx *SearchKey[k, v]) Search(query string, options ...SearchOption) ([]v, int64, error) {
	args := []interface{}{"FT.SEARCH", ctx.IndexName, query}

	// 应用分页等选项 (默认 Limit 0 10)
	args = append(args, "LIMIT", 0, 10)

	// 应用用户自定义选项
	for _, opt := range options {
		opt(&args)
	}

	// DIALECT 2 必须开启以获得更规范的 JSON/Array 响应
	args = append(args, "DIALECT", 2)

	cmd := ctx.Rds.Do(ctx.Context, args...)
	if cmd.Err() != nil {
		return nil, 0, cmd.Err()
	}

	return ctx.parseSearchResponse(cmd.Val())
}

// VectorSearch 执行向量近邻搜索 (KNN)
// vectorField: 结构体中标记为 vector 的字段名 (例如 "Embedding")
// vector: 浮点数向量
// topK: 返回结果数量
func (ctx *SearchKey[k, v]) VectorSearch(vectorField string, vector []float32, topK int) ([]v, []float64, error) {
	// 将 float32 转换为字节切片 (Little Endian)
	vecBytes := float32ToBytes(vector)

	// 构建 KNN 查询语句: "*=>[KNN 10 @vec $BLOB AS score]"
	// 这里的 * 表示全表预过滤，可以结合 filter 来做混合检索
	query := fmt.Sprintf("*=>[KNN %d @%s $BLOB AS score]", topK, vectorField)

	args := []interface{}{
		"FT.SEARCH", ctx.IndexName, query,
		"PARAMS", 2, "BLOB", vecBytes,
		"SORTBY", "score",
		"DIALECT", 2,
	}

	cmd := ctx.Rds.Do(ctx.Context, args...)
	if cmd.Err() != nil {
		return nil, nil, cmd.Err()
	}

	docs, _, err := ctx.parseSearchResponse(cmd.Val())
	// 注意：这里的 score 解析需要根据 RediSearch 具体的返回格式进一步完善
	// 目前为了简化，只返回文档列表，Score 返回 nil
	return docs, nil, err
}

// EnsureIndex 根据泛型 V 的结构体 Tag 自动创建索引
// 实现了幂等性：如果索引已存在则跳过，如果 Tag 变更目前不会自动更新（需要人工 Drop）
func (ctx *SearchKey[k, v]) EnsureIndex() error {
	// 1. 检查索引是否存在 (使用 FT.INFO)
	_, err := ctx.Rds.Do(ctx.Context, "FT.INFO", ctx.IndexName).Result()
	if err == nil {
		return nil // 索引已存在，无需操作
	}

	// 2. 生成 Schema
	schemaArgs := buildSchemaFromType[v]()
	if len(schemaArgs) == 0 {
		return fmt.Errorf("no search tags found in struct %T", *new(v))
	}

	// 3. 构造 FT.CREATE 命令
	// FT.CREATE idx ON HASH PREFIX 1 "prefix:" SCHEMA ...
	args := []interface{}{
		"FT.CREATE", ctx.IndexName,
		"ON", "HASH",
		"PREFIX", 1, ctx.Key + ":",
		"SCHEMA",
	}
	args = append(args, schemaArgs...)

	err = ctx.Rds.Do(ctx.Context, args...).Err()
	if err != nil {
		// 如果是因为并发导致索引已经存在，我们忽略这个错误
		if strings.Contains(err.Error(), "BUSYGROUP") || strings.Contains(err.Error(), "Index already exists") {
			return nil
		}
	}
	return err
}

// parseSearchResponse 解析 FT.SEARCH 的响应 (DIALECT 2)
// 格式: [total_count, key1, [field1, val1, field2, val2...], key2, ...]
func (ctx *SearchKey[k, v]) parseSearchResponse(resp interface{}) (results []v, total int64, err error) {
	arr, ok := resp.([]interface{})
	if !ok || len(arr) < 1 {
		return nil, 0, fmt.Errorf("invalid search response format")
	}

	// 第一个元素是总数
	if n, ok := arr[0].(int64); ok {
		total = n
	} else {
		// 兼容部分环境返回 int
		if n, ok := arr[0].(int); ok {
			total = int64(n)
		}
	}

	if total == 0 {
		return make([]v, 0), 0, nil
	}

	// 遍历结果：Step = 2
	// arr[1] 是 Key, arr[2] 是 FieldsArray
	for i := 1; i < len(arr); i += 2 {
		// keyStr := arr[i].(string) // 如果需要 key 可以在这里获取

		if i+1 >= len(arr) {
			break
		}

		fieldsArr, ok := arr[i+1].([]interface{})
		if !ok {
			continue
		}

		// 将 fields array 转换为 map 用于填充 struct
		fieldsMap := make(map[string]interface{})
		for j := 0; j < len(fieldsArr); j += 2 {
			if j+1 >= len(fieldsArr) {
				break
			}
			keyStr, _ := fieldsArr[j].(string)
			val := fieldsArr[j+1]
			fieldsMap[keyStr] = val
		}

		// 创建新的结构体实例
		newVal := new(v) // 此时 v 应该是指针类型，如 *User

		fillStructFromMap(newVal, fieldsMap)
		results = append(results, *newVal)
	}
	return results, total, nil
}

// fillStructFromMap 简单的填充逻辑，增强了类型兼容性
func fillStructFromMap(ptr interface{}, data map[string]interface{}) {
	val := reflect.ValueOf(ptr).Elem() // Assume ptr is *Struct
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}
		val = val.Elem()
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("msgpack")
		if tag == "" {
			tag = field.Name
		}

		if v, ok := data[tag]; ok {
			fieldVal := val.Field(i)
			if !fieldVal.CanSet() {
				continue
			}

			// Redis 返回的可能是 string 或 []byte
			// 尝试自动转换
			switch fieldVal.Kind() {
			case reflect.String:
				if str, ok := v.(string); ok {
					fieldVal.SetString(str)
				} else if b, ok := v.([]byte); ok {
					fieldVal.SetString(string(b))
				}
			case reflect.Int, reflect.Int64, reflect.Int32:
				// 简单的数字处理 (go-redis 可能会将数字作为字符串返回)
				// 这里为精简代码，仅做简单类型断言，生产环境建议使用 cast 库
				if num, ok := v.(int64); ok {
					fieldVal.SetInt(num)
				}
			case reflect.Slice:
				// Vector 或 Byte 数据
				if b, ok := v.([]byte); ok {
					if fieldVal.Type().Elem().Kind() == reflect.Uint8 {
						fieldVal.SetBytes(b)
					}
					// TODO: 如果是 []float32 (Vector) 反序列化，这里需要 binary 解析
					// 目前 VectorSearch 仅用于检索，返回结果通常不需要 Vector 原文
				}
			default:
				if reflect.TypeOf(v).AssignableTo(fieldVal.Type()) {
					fieldVal.Set(reflect.ValueOf(v))
				}
			}
		}
	}
}

// buildSchemaFromType 通过反射解析 struct tag 生成 FT.CREATE 的参数
func buildSchemaFromType[v any]() []interface{} {
	var args []interface{}
	t := reflect.TypeOf((*v)(nil)).Elem()

	// 处理指针情况
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("search")
		if tag == "" {
			continue
		}

		// 默认使用字段名，或 tag 指定的名称
		parts := strings.Split(tag, ",")
		redisName := field.Tag.Get("msgpack") // 优先复用 msgpack tag 作为存储字段名
		if redisName == "" {
			redisName = field.Name
		}

		fieldType := parts[0] // text, tag, numeric, vector

		args = append(args, redisName, strings.ToUpper(fieldType))

		// 处理 Vector 特有的参数: vector,HNSW,dim=1536,dist=COSINE
		if fieldType == "vector" {
			// 默认值
			algo := "HNSW"
			dim := "1536" // OpenAI default
			dist := "COSINE"
			type_ := "FLOAT32"

			// 解析参数
			for _, p := range parts[1:] {
				kv := strings.Split(p, "=")
				if len(kv) == 1 {
					if strings.ToUpper(kv[0]) == "HNSW" || strings.ToUpper(kv[0]) == "FLAT" {
						algo = strings.ToUpper(kv[0])
					}
				} else if len(kv) == 2 {
					switch kv[0] {
					case "dim":
						dim = kv[1]
					case "dist":
						dist = kv[1]
					case "type":
						type_ = kv[1]
					}
				}
			}

			// 修正后的追加参数逻辑：包含 VECTOR 关键字和 ALGO
			// Syntax: ... VECTOR <ALGO> <NARGS> [TYPE type] [DIM dim] [DISTANCE_METRIC dist]
			args = append(args, "VECTOR", algo, 6, "TYPE", type_, "DIM", dim, "DISTANCE_METRIC", dist)

		} else {
			// 处理 text/tag 的 extra args 比如 WEIGHT, SORTABLE
			for _, p := range parts[1:] {
				args = append(args, strings.ToUpper(p))
			}
		}
	}
	return args
}

// structToFlatMap 将结构体打平为 map，用于 HSET
func structToFlatMap(v interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		// 建议只存 msgpack 标记的字段
		name := field.Tag.Get("msgpack")
		if name == "" {
			name = field.Name
		}

		fieldVal := val.Field(i).Interface()

		// 特殊处理 Vector (slice of float32) -> Byte Slice
		if field.Tag.Get("search") != "" && strings.Contains(field.Tag.Get("search"), "vector") {
			if vec, ok := fieldVal.([]float32); ok {
				out[name] = float32ToBytes(vec)
				continue
			}
		}

		out[name] = fieldVal
	}
	return out, nil
}

func float32ToBytes(floats []float32) []byte {
	bytes := make([]byte, len(floats)*4)
	for i, f := range floats {
		bits := math.Float32bits(f)
		binary.LittleEndian.PutUint32(bytes[i*4:], bits)
	}
	return bytes
}

// SearchOption 定义为一个修改命令参数切片的函数
type SearchOption func(args *[]interface{})

// SearchLimit 分页限制
func SearchLimit(offset, num int) SearchOption {
	return func(args *[]interface{}) {
		*args = append(*args, "LIMIT", offset, num)
	}
}

// SearchSortBy 排序
func SearchSortBy(field string, asc bool) SearchOption {
	return func(args *[]interface{}) {
		direction := "DESC"
		if asc {
			direction = "ASC"
		}
		*args = append(*args, "SORTBY", field, direction)
	}
}

// SearchReturn 指定返回字段
func SearchReturn(fields ...string) SearchOption {
	return func(args *[]interface{}) {
		if len(fields) > 0 {
			*args = append(*args, "RETURN", len(fields))
			for _, f := range fields {
				*args = append(*args, f)
			}
		}
	}
}

// SearchHighlight 高亮匹配字段
func SearchHighlight(openTag, closeTag string) SearchOption {
	return func(args *[]interface{}) {
		*args = append(*args, "HIGHLIGHT")
		if openTag != "" && closeTag != "" {
			*args = append(*args, "TAGS", openTag, closeTag)
		}
	}
}

// SearchVerbatim 禁用查询扩展
func SearchVerbatim() SearchOption {
	return func(args *[]interface{}) {
		*args = append(*args, "VERBATIM")
	}
}

// SearchWithScores 返回相关性分数
func SearchWithScores() SearchOption {
	return func(args *[]interface{}) {
		*args = append(*args, "WITHSCORES")
	}
}
