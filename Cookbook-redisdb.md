# RedisDB Context for LLM

**Package:** `github.com/doptime/redisdb`
**Core Philosophy:** Type-safe Redis wrapper using Go Generics `[K, V]` with integrated Web Permissions, Auto-Serialization, and RediSearch/Vector support.

---

## 1. Setup & Factory Pattern

**Config (`config.toml`):**

```toml
[[Redis]]
Host = "127.0.0.1"
Port = 6379
DB = 0
Name = "default"
Password = "password"

```

**Initialization:**
All keys are created via `New{Type}Key[K, V](options...)`.

* `WithKey("prefix:name")`: Defines the static key or prefix.
* `WithRds("default")`: Selects the Redis connection block.
* `HttpOn(Op)`: Registers allowed Web/HTTP operations (Fluent API).

---

## 2. Models & Struct Tags

Models used as `V` (Values) support specific tags for storage and validation.

| Tag | Context | Description | Example |
| --- | --- | --- | --- |
| `msgpack` | Storage | Field name in Redis (Hash/String) | `msgpack:"uid"` |
| `json` | Search | Field name for **VectorSet/Search** mapping | `json:"title"` |
| `mod` | Pre-Save | Modifiers applied before saving | `mod:"trim,lowercase"` |
| `validate` | Validation | Rules via `go-playground/validator` | `validate:"required,email"` |

*Note: `CreatedAt`/`UpdatedAt` (`time.Time`) are auto-populated on save.*

---

## 3. Permission System (Web/HTTP)

RedisDB enforces operation whitelisting based on **Key Scope** (prefix before the first `:`).

**Registration:**

```go
// Allow Reading Hash data for keys starting with "user:"
var UserKey = redisdb.NewHashKey[string, *User](
    redisdb.WithKey("user:profile"),
).HttpOn(redisdb.HashRead)

```

**Verification Global Functions:**

* `IsAllowHashOp(key, op)`
* `IsAllowListOp(key, op)`
* `IsAllowSetOp(key, op)`
* `IsAllowZSetOp(key, op)`
* `IsAllowStringOp(key, op)`
* `IsAllowStreamOp(key, op)`
* `IsAllowVectorSetOp(key, op)`

**Op Constants Groups:** `HashRead`, `HashWrite`, `HashAll`, `VectorSetAll`, etc.

---

## 4. API Reference: Standard Data Types

### A. StringKey

**Signature:** `StringKey[k comparable, v any]`

```go
// Init
kStr := redisdb.NewStringKey[string, *User](redisdb.WithKey("u:str"))

// API Signatures
func (c *StringKey[k, v]) Set(key k, value v, expiration time.Duration) error
func (c *StringKey[k, v]) Get(key k) (v, error)
func (c *StringKey[k, v]) Del(key k) error
func (c *StringKey[k, v]) GetAll(match string) (map[k]v, error) // Scan & Get
func (c *StringKey[k, v]) SetAll(kvMap map[k]v) error           // Pipeline Set

```

### B. HashKey

**Signature:** `HashKey[k comparable, v any]`
*Note: `Save(v)` auto-detects the Primary Key if `v` is a struct and a field matches type `k`.*

```go
// Init
kHash := redisdb.NewHashKey[string, *User](redisdb.WithKey("u:hash"))

// API Signatures
func (c *HashKey[k, v]) HSet(values ...interface{}) (int64, error)
func (c *HashKey[k, v]) HMSet(kvMap map[k]v) (int64, error)
func (c *HashKey[k, v]) HGet(field k) (v, error)
func (c *HashKey[k, v]) HGetAll() (map[k]v, error)
func (c *HashKey[k, v]) HDel(fields ...k) error
func (c *HashKey[k, v]) HExists(field k) (bool, error)
func (c *HashKey[k, v]) HLen() (int64, error)
func (c *HashKey[k, v]) HKeys() ([]k, error)
func (c *HashKey[k, v]) HVals() ([]v, error)
func (c *HashKey[k, v]) HIncrBy(field k, increment int64) error
func (c *HashKey[k, v]) HIncrByFloat(field k, increment float64) error
func (c *HashKey[k, v]) HSetNX(field k, value v) error
func (c *HashKey[k, v]) HScan(cursor uint64, match string, count int64) ([]k, []v, uint64, error)
func (c *HashKey[k, v]) Save(value v) (int64, error) // Auto PK detection

```

### C. ListKey

**Signature:** `ListKey[v any]` (Key is always string)

```go
// Init
kList := redisdb.NewListKey[*Task](redisdb.WithKey("q:tasks"))

// API Signatures
func (c *ListKey[v]) RPush(values ...v) error
func (c *ListKey[v]) LPush(values ...v) error
func (c *ListKey[v]) RPop() (v, error)
func (c *ListKey[v]) LPop() (v, error)
func (c *ListKey[v]) LRange(start, stop int64) ([]v, error)
func (c *ListKey[v]) LRem(count int64, value v) error
func (c *ListKey[v]) LSet(index int64, value v) error
func (c *ListKey[v]) LIndex(index int64) (v, error)
func (c *ListKey[v]) BLPop(timeout time.Duration) (v, error)
func (c *ListKey[v]) BRPop(timeout time.Duration) (v, error)
func (c *ListKey[v]) LTrim(start, stop int64) error
func (c *ListKey[v]) LLen() (int64, error)

```

### D. SetKey

**Signature:** `SetKey[k comparable, v any]`

```go
// Init
kSet := redisdb.NewSetKey[string, string](redisdb.WithKey("u:tags"))

// API Signatures
func (c *SetKey[k, v]) SAdd(member v) error
func (c *SetKey[k, v]) SRem(member v) error
func (c *SetKey[k, v]) SIsMember(member v) (bool, error)
func (c *SetKey[k, v]) SMembers() ([]v, error)
func (c *SetKey[k, v]) SCard() (int64, error)
func (c *SetKey[k, v]) SScan(cursor uint64, match string, count int64) ([]v, uint64, error)

```

### E. ZSetKey

**Signature:** `ZSetKey[k comparable, v any]`
*Dependency: Uses `github.com/redis/go-redis/v9` structs.*

```go
// Init
kZSet := redisdb.NewZSetKey[string, *User](redisdb.WithKey("leaderboard"))

// API Signatures
func (c *ZSetKey[k, v]) ZAdd(members ...redis.Z) error
func (c *ZSetKey[k, v]) ZRem(members ...interface{}) error
func (c *ZSetKey[k, v]) ZRange(start, stop int64) ([]v, error)
func (c *ZSetKey[k, v]) ZRangeWithScores(start, stop int64) ([]v, []float64, error)
func (c *ZSetKey[k, v]) ZRevRange(start, stop int64) ([]v, error)
func (c *ZSetKey[k, v]) ZRank(member interface{}) (int64, error)
func (c *ZSetKey[k, v]) ZScore(member v) (float64, error)
func (c *ZSetKey[k, v]) ZCard() (int64, error)
func (c *ZSetKey[k, v]) ZCount(min, max string) (int64, error)
func (c *ZSetKey[k, v]) ZIncrBy(increment float64, member v) error
func (c *ZSetKey[k, v]) ZPopMax(count int64) ([]v, []float64, error)
func (c *ZSetKey[k, v]) ZPopMin(count int64) ([]v, []float64, error)

```

---

## 5. API Reference: Vector & Search

### F. VectorSetKey

**Signature:** `VectorSetKey[k comparable, v any]`
**Purpose:** RediSearch (Text) and Vector Similarity (KNN).
**Note:** `v` should be a Struct (with `json` tags) or `map[string]interface{}`.

```go
// Init
kVec := redisdb.NewVectorSetKey[string, *Doc](redisdb.WithKey("idx:docs"))

// API Signatures

// 1. Index Management
func (c *VectorSetKey[k, v]) Create(args ...interface{}) error
func (c *VectorSetKey[k, v]) DropIndex(deleteDocs bool) error
func (c *VectorSetKey[k, v]) Info() (map[string]interface{}, error)
func (c *VectorSetKey[k, v]) AliasAdd(alias string) error
func (c *VectorSetKey[k, v]) AliasUpdate(alias string) error
func (c *VectorSetKey[k, v]) AliasDel(alias string) error
func (c *VectorSetKey[k, v]) TagVals(fieldName string) ([]string, error)

// 2. Search
func (c *VectorSetKey[k, v]) Search(query string, params ...interface{}) (int64, []v, error)

// 3. Helpers
// Float32ToBytes converts []float32 to LittleEndian []byte for BLOB
func (c *VectorSetKey[k, v]) Float32ToBytes(floats []float32) []byte 
// KNNParamHelper constructs query syntax for KNN search
// Returns: query string ("*=>[KNN...]"), params slice
func (c *VectorSetKey[k, v]) KNNParamHelper(knum int, vecField string, vector []float32) (string, []interface{})

```

**Usage Example (Vector Create & Search):**

```go
// Create
kVec.Create(
    "ON", "HASH", "PREFIX", "1", "doc:",
    "SCHEMA", "title", "TEXT", 
    "vec", "VECTOR", "HNSW", "6", "TYPE", "FLOAT32", "DIM", "128", "DISTANCE_METRIC", "L2",
)

// Search
query, params := kVec.KNNParamHelper(10, "vec", []float32{0.1, ...})
count, docs, err := kVec.Search(query, params...)

```

---

## 6. Utilities & Modifiers

### Dynamic Keys

Use `ConcatKey` to create a new key instance derived from a base key.

```go
base := redisdb.NewStringKey[string, any](redisdb.WithKey("user"))
user1 := base.ConcatKey("1001") // Key becomes "user:1001"

```

### Modifiers Reference (`mod` tag)

Directives applied to struct fields before saving.

* `trim`: `strings.TrimSpace`
* `lowercase`: `strings.ToLower`
* `uppercase`: `strings.ToUpper`
* `default=X`: Sets value `X` if field is zero/empty.
* `unixtime`: Converts `time.Time` to `int64` (seconds).
* `unixtime=ms`: Converts `time.Time` to `int64` (milliseconds).
* `now`: Sets current time if zero.
* `force`: Applies modifier even if field is already set.