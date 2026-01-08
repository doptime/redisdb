package redisdb

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"reflect"

	"github.com/doptime/logger"
)

// -----------------------------------------------------------------------------
//  VectorSetKey Implementation
// -----------------------------------------------------------------------------

// VectorSetKey represents a RediSearch Index (supporting both Text and Vector Search).
// k: Document Key Type (id), usually string.
// v: Document Value Type (payload), struct or map.
type VectorSetKey[k comparable, v any] struct {
	RedisKey[k, v]
}

func NewVectorSetKey[k comparable, v any](ops ...Option) *VectorSetKey[k, v] {
	ctx := &VectorSetKey[k, v]{RedisKey: RedisKey[k, v]{KeyType: KeyTypeVectorSet}}
	if err := ctx.applyOptionsAndCheck(KeyTypeVectorSet, ops...); err != nil {
		logger.Error().Err(err).Msg("redisdb.NewVectorSetKey failed")
		return nil
	}
	ctx.InitFunc()
	return ctx
}

func (ctx *VectorSetKey[k, v]) ConcatKey(fields ...interface{}) *VectorSetKey[k, v] {
	return &VectorSetKey[k, v]{ctx.Duplicate(ConcatedKeys(ctx.Key, fields...), ctx.RdsName)}
}

func (ctx *VectorSetKey[k, v]) HttpOn(op VectorSetOp) *VectorSetKey[k, v] {
	httpAllow(ctx.Key, uint64(op))
	if op != 0 && ctx.Key != "" {
		ctx.RegisterWebDataSchemaDocForWebVisit()
		ctx.RegisterKeyInterfaceForWebVisit()
	}
	return ctx
}

// -----------------------------------------------------------------------------
//  Index Management
// -----------------------------------------------------------------------------

// Create executes FT.CREATE.
func (ctx *VectorSetKey[k, v]) Create(args ...interface{}) error {
	cmdArgs := append([]interface{}{"FT.CREATE", ctx.Key}, args...)
	return ctx.Rds.Do(ctx.Context, cmdArgs...).Err()
}

// DropIndex deletes the index. If deleteDocs is true, it passes DD to delete documents.
func (ctx *VectorSetKey[k, v]) DropIndex(deleteDocs bool) error {
	args := []interface{}{"FT.DROPINDEX", ctx.Key}
	if deleteDocs {
		args = append(args, "DD")
	}
	return ctx.Rds.Do(ctx.Context, args...).Err()
}

// Info retrieves index statistics.
func (ctx *VectorSetKey[k, v]) Info() (map[string]interface{}, error) {
	res, err := ctx.Rds.Do(ctx.Context, "FT.INFO", ctx.Key).Result()
	if err != nil {
		return nil, err
	}
	return ctx.parseRawSliceToMap(res)
}

// AliasAdd adds an alias to the index.
func (ctx *VectorSetKey[k, v]) AliasAdd(alias string) error {
	return ctx.Rds.Do(ctx.Context, "FT.ALIASADD", alias, ctx.Key).Err()
}

// AliasUpdate updates an alias to point to this index.
func (ctx *VectorSetKey[k, v]) AliasUpdate(alias string) error {
	return ctx.Rds.Do(ctx.Context, "FT.ALIASUPDATE", alias, ctx.Key).Err()
}

// AliasDel deletes an alias.
func (ctx *VectorSetKey[k, v]) AliasDel(alias string) error {
	return ctx.Rds.Do(ctx.Context, "FT.ALIASDEL", alias).Err()
}

// TagVals returns the distinct values indexed in a Tag field.
func (ctx *VectorSetKey[k, v]) TagVals(fieldName string) ([]string, error) {
	res, err := ctx.Rds.Do(ctx.Context, "FT.TAGVALS", ctx.Key, fieldName).Result()
	if err != nil {
		return nil, err
	}
	// Convert interface{} slice to string slice
	if slice, ok := res.([]interface{}); ok {
		strs := make([]string, len(slice))
		for i, v := range slice {
			strs[i] = fmt.Sprint(v)
		}
		return strs, nil
	}
	return nil, fmt.Errorf("unexpected format for FT.TAGVALS")
}

// -----------------------------------------------------------------------------
//  Search Operations
// -----------------------------------------------------------------------------

// Search executes FT.SEARCH.
// Returns total count and slice of documents (v).
func (ctx *VectorSetKey[k, v]) Search(query string, params ...interface{}) (count int64, docs []v, err error) {
	args := append([]interface{}{"FT.SEARCH", ctx.Key, query}, params...)

	res, err := ctx.Rds.Do(ctx.Context, args...).Result()
	if err != nil {
		return 0, nil, err
	}

	slice, ok := res.([]interface{})
	if !ok || len(slice) < 1 {
		return 0, nil, fmt.Errorf("unexpected response format from FT.SEARCH")
	}

	// Parse Count
	switch c := slice[0].(type) {
	case int64:
		count = c
	case int:
		count = int64(c)
	default:
		return 0, nil, fmt.Errorf("unexpected count type")
	}

	// Parse Documents (Format: Key, Fields, Key, Fields...)
	docs = make([]v, 0, (len(slice)-1)/2)
	for i := 1; i < len(slice); i += 2 {
		fieldsData := slice[i+1]
		doc, err := ctx.parseDocument(fieldsData)
		if err != nil {
			logger.Warn().Err(err).Msg("failed to parse search document")
			continue
		}
		docs = append(docs, doc)
	}

	return count, docs, nil
}

// parseDocument robustly converts Redis return data (Slice or Map) into Struct 'v'.
func (ctx *VectorSetKey[k, v]) parseDocument(data interface{}) (val v, err error) {
	// 1. Normalize data to map[string]interface{}
	kvMap := make(map[string]interface{})

	if fieldSlice, ok := data.([]interface{}); ok {
		// Format: [field1, val1, field2, val2]
		for j := 0; j < len(fieldSlice); j += 2 {
			if fName, ok := fieldSlice[j].(string); ok && j+1 < len(fieldSlice) {
				kvMap[fName] = fieldSlice[j+1]
			}
		}
	} else if fieldMap, ok := data.(map[interface{}]interface{}); ok {
		for fk, fv := range fieldMap {
			if fks, ok := fk.(string); ok {
				kvMap[fks] = fv
			}
		}
	} else {
		return val, nil // Unable to parse
	}

	// 2. Map to 'v'
	// If 'v' is map[string]interface{}, return directly
	// Note: We use reflection to check type of v, not val (which is zero value)
	vType := reflect.TypeOf(val)
	if vType == nil {
		// v is interface{}, try to return map
		if m, ok := any(kvMap).(v); ok {
			return m, nil
		}
	} else if vType.Kind() == reflect.Map {
		if m, ok := any(kvMap).(v); ok {
			return m, nil
		}
	}

	// 3. If 'v' is Struct, use JSON Round-Trip for robust mapping (handles tags)
	// This maps "field_name" -> Struct Field `json:"field_name"`
	bytes, err := json.Marshal(kvMap)
	if err != nil {
		return val, err
	}

	// We need a pointer to unmarshal
	// If val is a value type (e.g. User), we need &val.
	// If val is a pointer type (e.g. *User), val is nil, we need to allocate.

	// Simple approach: New instance of v
	ptrVal := new(v) // *v
	if err := json.Unmarshal(bytes, ptrVal); err != nil {
		return val, err
	}
	return *ptrVal, nil
}

func (ctx *VectorSetKey[k, v]) parseRawSliceToMap(res interface{}) (map[string]interface{}, error) {
	slice, ok := res.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	info := make(map[string]interface{})
	for i := 0; i < len(slice); i += 2 {
		if key, ok := slice[i].(string); ok && i+1 < len(slice) {
			info[key] = slice[i+1]
		}
	}
	return info, nil
}

// -----------------------------------------------------------------------------
//  Vector Utilities
// -----------------------------------------------------------------------------

func (ctx *VectorSetKey[k, v]) Float32ToBytes(floats []float32) []byte {
	bytes := make([]byte, len(floats)*4)
	for i, f := range floats {
		binary.LittleEndian.PutUint32(bytes[i*4:], math.Float32bits(f))
	}
	return bytes
}

func (ctx *VectorSetKey[k, v]) BytesToFloat32(bytes []byte) ([]float32, error) {
	if len(bytes)%4 != 0 {
		return nil, fmt.Errorf("invalid byte length for float32 vector")
	}
	floats := make([]float32, len(bytes)/4)
	for i := 0; i < len(floats); i++ {
		floats[i] = math.Float32frombits(binary.LittleEndian.Uint32(bytes[i*4:]))
	}
	return floats, nil
}

// KNNParamHelper constructs query syntax for KNN search.
// knum: k nearest neighbors
// vecField: field name in schema
// vector: query vector
func (ctx *VectorSetKey[k, v]) KNNParamHelper(knum int, vecField string, vector []float32) (string, []interface{}) {
	queryPart := fmt.Sprintf("[KNN %d @%s $BLOB]", knum, vecField)
	blob := ctx.Float32ToBytes(vector)
	params := []interface{}{"PARAMS", "2", "BLOB", blob}
	return queryPart, params
}
