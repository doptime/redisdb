# RedisDB Context for LLM

**Package:** `github.com/doptime/redisdb`
**Core Function:** Type-safe Redis wrapper with generics `[K, V]`, auto-serialization (msgpack), modifiers, and validation.

## 1. Configuration & Setup

**Config Format (config.toml):**
```toml
[[Redis]]
DB = 0
Host = "host.lan"
Name = "default"
Password = "password"
Port = 6379
Username = ""
````

**Init:**

```go
import (
    "[github.com/doptime/redisdb](https://github.com/doptime/redisdb)" // Import required to use redisdb.New...
)
// cfgredis.Servers.Load("config.toml") must be called before usage
```

## 2\. Struct Tags & Magic Fields

Models used as `V` (Value) support specific tags.

| Tag | Description | Examples |
| :--- | :--- | :--- |
| `msgpack` | Field name in Redis | `msgpack:"user_id"` |
| `mod` | Pre-save modifiers | `mod:"trim,lowercase,default=0"` |
| `validate` | `go-playground/validator` | `validate:"required,email"` |

**Auto-Fill:** `CreatedAt`/`UpdatedAt` (`time.Time`) are automatically set on save.

```go
type User struct {
    ID        string    `msgpack:"id" mod:"trim"`
    Email     string    `msgpack:"email" mod:"trim,lowercase" validate:"email"`
    Count     int       `msgpack:"cnt" mod:"default=1"`
    CreatedAt time.Time // Auto-filled
    UpdatedAt time.Time // Auto-filled
}
```

## 3\. Type Constructors

**Pattern:** `New{Type}Key[KeyType, ValType](options...)`
**Options:**

  * `WithKey(key string)`: Define Redis key name.
  * `WithRds(dsName string)`: Select Redis config block (default "default").
  * `WithModifier(map)`: Register extra modifiers.

## 4\. Usage Patterns & API Signatures

### StringKey (Simple K-V)

```go
// Init: Key=string, Val=*User
kStr := redisdb.NewStringKey[string, *User](redisdb.WithKey("u:str"))

// Usage
err := kStr.Set("u1", &User{ID: "u1"}, time.Hour)
user, err := kStr.Get("u1")
```

**API Signatures:**

```go
Set(key k, value v, expiration time.Duration) error
Get(key k) (v, error)
Del(key k) error
```

### HashKey (Map/Object)

```go
// Init
kHash := redisdb.NewHashKey[string, *User](redisdb.WithKey("u:hash"))

// Usage
err := kHash.HSet("u1", &User{ID: "A"}) 
err := kHash.HMSet(map[string]*User{"u2": {ID: "B"}})
u, err := kHash.HGet("u1")
all, err := kHash.HGetAll() // Returns map[string]*User
```

**API Signatures:**

```go
HSet(values ...interface{}) error // Standard redis HSet (keys/values interleaved)
HMSet(kvMap map[k]v) error        // Type-safe map setter
HGet(field k) (v, error)
HGetAll() (map[k]v, error)
HDel(fields ...k) error
HExists(field k) (bool, error)
HKeys() ([]k, error)
HVals() ([]v, error)
HLen() (int64, error)
```

### ListKey (Queue/Stack)

```go
// Init
kList := redisdb.NewListKey[string, *User](redisdb.WithKey("u:list"))

// Usage
err := kList.RPush(&User{ID: "A"}) 
u, err := kList.LPop() 
users, err := kList.LRange(0, -1)
```

**API Signatures:**

```go
RPush(values ...v) error
LPush(values ...v) error
RPop() (v, error)
LPop() (v, error)
LRange(start, stop int64) ([]v, error) // 0, -1 for all
LRem(count int64, value v) error       // Note: Count is FIRST argument
LTrim(start, stop int64) error
```

### SetKey (Unique Collection)

```go
// Init
kSet := redisdb.NewSetKey[string, *User](redisdb.WithKey("u:set"))

// Usage
err := kSet.SAdd(&User{ID: "A"})
exists, err := kSet.SIsMember(&User{ID: "A"})
users, err := kSet.SMembers()
```

**API Signatures:**

```go
SAdd(param v) error
SRem(param v) error
SIsMember(param v) (bool, error)
SMembers() ([]v, error)
SCard() (int64, error)
SScan(cursor uint64, match string, count int64) ([]v, uint64, error)
```

### ZSetKey (Sorted Set)

```go
// Init
kZSet := redisdb.NewZSetKey[string, *ScoreItem](redisdb.WithKey("u:zset"))

// Usage
err := kZSet.ZAdd(redis.Z{Score: 100, Member: item})
items, err := kZSet.ZRange(0, -1)
```

**API Signatures:**

```go
// Depends on [github.com/redis/go-redis/v9](https://github.com/redis/go-redis/v9)
ZAdd(members ...redis.Z) error 
ZRange(start, stop int64) ([]v, error)
```

## 5\. Modifiers Reference

Applied pre-serialization via `mod` tag.

  * `trim`: Strip whitespace.
  * `lowercase`: Convert to lower.
  * `default={val}`: Set if zero-value.
  * `unixtime=ms`: Convert time to int64 ms.
  * `counter`: Auto-increment.
  * `force`: Apply even if field is set.

<!-- end list -->

```go
// Manual Invocation
redisdb.ApplyModifiers(obj) 
```

```