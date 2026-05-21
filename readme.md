# RedisDB

`github.com/doptime/redisdb` —— 类型化 Redis 包装,泛型 `[K comparable, V any]`,
msgpack 自动编解,内置 mod 修饰、HTTP 暴露、RediSearch / KNN。8 种 Key。

```bash
go get github.com/doptime/redisdb
```

```go
type User struct {
    UID  string `msgpack:"uid"  mod:"trim,lowercase"`
    Name string `msgpack:"name" validate:"required"`
}

var Users = redisdb.NewHashKey[string, *User](redisdb.WithKey("users")).
    HttpOn(redisdb.HashRead | redisdb.HashWrite)

_, _ = Users.HSet("u1", &User{UID: " U1 ", Name: "Alice"})  // UID 落库时变 "u1"
u, _ := Users.HGet("u1")                                     // u 是 *User
```

**目录**
[StringKey](#stringkey) ·
[HashKey](#hashkey) ·
[ListKey](#listkey) ·
[SetKey](#setkey) ·
[ZSetKey](#zsetkey) ·
[StreamKey](#streamkey) ·
[VectorSetKey](#vectorsetkey) ·
[SearchKey](#searchkey) ·
[公共契约](#common) ·
[HttpOn 权限位](#httpon)

---

<a id="common"></a>
## 公共契约

### 构造

`NewXxxKey(ops ...Option)` 失败**只写日志返回 `nil`**,调用方要 nil 检。三个 helper 可链:

```go
redisdb.WithKey("users")                          // Redis key/前缀;省略则取 V 的类型名
redisdb.WithRds("cache")                          // 选 config.toml 里 [[Redis]] 的 Name,默认 "default"
redisdb.WithModifier(map[string]ModifierFunc{…})  // 注册额外 mod 指令
```

### 派生 key

每种 Key 都有 `ConcatKey(fields ...interface{}) *Self` —— **返回新实例**,原 ctx 不变,
新 key = `oldKey:field1:field2…`。时间分桶 helper(ISO 周):

```go
redisdb.CatYearMonthDay(t)  // "YMD_20260521"
redisdb.CatYearMonth(t)     // "YM_202605"
redisdb.CatYear(t)          // "Y_2026"
redisdb.CatYearWeek(t)      // "YW_202621"
```

### Struct tag

| tag | 用途 |
| --- | --- |
| `msgpack:"…"` | 存储编解(所有非检索类型) |
| `json:"…"` | `VectorSetKey` / `SearchKey` 字段映射 |
| `mod:"…"` | 写入前修饰(见下) |
| `validate:"…"` | go-playground/validator 校验 |

### mod 指令

`tag_modifier.go` 实际只注册了 8 个,**没有 `now`**:

| 指令 | 行为 | 参数 |
| --- | --- | --- |
| `trim` `lowercase` `uppercase` `title` | 字符串变换 | — |
| `default=X` | 零值时填 X(按字段类型解析) | 必需 |
| `unixtime` | 当前 Unix 秒(int64) | 可选 `=ms` 毫秒 |
| `counter` | 数值字段 +1(所有 int/uint/float 宽度) | — |
| `nanoid` | 生成 NanoID | 可选 `=N`,默认 21 |

**默认只在字段为零值时跑**,加 `,force` 后无条件跑;多指令用 `,` 串接,从左到右执行。
手动触发:`redisdb.ApplyModifiers(&v)`。

### CreatedAt / UpdatedAt 自动填充

V 里若有**严格命名**为 `CreatedAt` 或 `UpdatedAt` 且类型为 `time.Time`(非指针)的字段,
框架会在反序列化时自动接管:

- `CreatedAt`:**仅在零值时**填 `time.Now().UTC()`
- `UpdatedAt`:**总是**覆盖为 `time.Now().UTC()`

- 💡 触发点是 **read 时**(`DeserializeToInterface`),**不是 save**;而且只走 HTTP 暴露层的反序列化路径,常规 Go 直接调 `HGet`/`Get`/`LPop` 等**不触发**
- 💡 字段名严格区分大小写;`createdAt` / `CreateAt` / `*time.Time` / `int64` 一律不识别 —— 想要写入时戳还是用 `mod:"unixtime=ms,force"`

### 错误约定

- `redis.Nil` = key/字段不存在,用 `errors.Is` 区分
- 出错时**绝不返回半填的 struct**,一律返回零值 + err
- `HGetAll` / `GetAll` / `HMGET` / `*Scan` 批量读里**单条**解码失败静默跳过(只写日志)

### config.toml

```toml
[[Redis]]
Name = "default"
Host = "127.0.0.1"
Port = 6379
DB   = 0
```

---

<a id="stringkey"></a>
## StringKey `[K comparable, V any]`

一个 ctx 对应**一族**独立 Redis string,实际键名 = `ctx.Key + ":" + serialize(K)`。

```go
func NewStringKey[K comparable, V any](ops ...Option) *StringKey[K, V]
func (c *StringKey[K, V]) ConcatKey(fields ...interface{}) *StringKey[K, V]

func (c *StringKey[K, V]) Set(key K, value V, expiration time.Duration) error
func (c *StringKey[K, V]) Get(field K) (V, error)
func (c *StringKey[K, V]) Del(key K) error

func (c *StringKey[K, V]) GetAll(match string) (map[K]V, error)   // SCAN+GET,上限 1GiB,别在热路径用
func (c *StringKey[K, V]) SetAll(m map[K]V) error                  // Pipeline,会清掉已有 TTL

func (c *StringKey[K, V]) Scan(cursor uint64, match string, count int64) ([]string, uint64, error)
func (c *StringKey[K, V]) Keys() ([]K, error)
func (c *StringKey[K, V]) HttpOn(op StringOp) *StringKey[K, V]
```

- 💡 `Set` 的 `expiration`:`0` = 永不过期,**负值** = 清掉已有 TTL
- 💡 `SetAll` 内部固定传 `-1`,会**清掉同名 key 原有的过期时间**

---

<a id="hashkey"></a>
## HashKey `[K comparable, V any]`

一个 ctx 对应**一个** Redis hash,field=K,value=V。

```go
func NewHashKey[K comparable, V any](ops ...Option) *HashKey[K, V]
func (c *HashKey[K, V]) ConcatKey(fields ...interface{}) *HashKey[K, V]

func (c *HashKey[K, V]) HSet(values ...interface{}) (int64, error)  // (k,v,k,v,...) 或单个 map[K]V
func (c *HashKey[K, V]) HMSet(kvMap map[K]V)        (int64, error)
func (c *HashKey[K, V]) HSetNX(field K, value V)    error
func (c *HashKey[K, V]) Save(value V)               (int64, error)  // 自动从 struct 取 field
func (c *HashKey[K, V]) HDel(fields ...K)           error
func (c *HashKey[K, V]) HIncrBy(field K, inc int64)        error
func (c *HashKey[K, V]) HIncrByFloat(field K, inc float64) error

func (c *HashKey[K, V]) HGet(field K)                (V, error)
func (c *HashKey[K, V]) HMGET(fields ...interface{}) ([]V, error)
func (c *HashKey[K, V]) HGetAll()                    (map[K]V, error)
func (c *HashKey[K, V]) HExists(field K)             (bool, error)
func (c *HashKey[K, V]) HKeys()                      ([]K, error)
func (c *HashKey[K, V]) HVals()                      ([]V, error)
func (c *HashKey[K, V]) HLen()                       (int64, error)

func (c *HashKey[K, V]) HRandField(count int)           ([]K, error)
func (c *HashKey[K, V]) HRandFieldWithValues(count int) ([]K, []V, error)
func (c *HashKey[K, V]) HScan(cursor uint64, match string, count int64)         ([]K, []V, uint64, error)
func (c *HashKey[K, V]) HScanNoValues(cursor uint64, match string, count int64) ([]K, uint64, error)
```

- 💡 `HSet` 散参格式必须**偶数对**,且 k,v 类型严格对齐 K,V,否则运行时报错
- 💡 `HSet` 返回的 `int64` 是**新增**字段数(Redis 语义),覆写的不算
- 💡 `Save` 在构造时自省 V,找第一个类型可赋给 K 的字段当主键;找不到就退化成单参 `HSet`
- 💡 `HIncrBy` 直接操作字段裸字节 —— 这个字段必须是数字字符串,**不能是 msgpack blob**
- 💡 `HDel`:K 是 string 直传;非 string 走 JSON 序列化(和写入时一致)
- 💡 `HRandField` 的 `count`:正数=去重,上限为 hash 大小;负数=可重复,正好 `|count|` 条

---

<a id="listkey"></a>
## ListKey `[V any]`

**单泛型参数** —— K 在底层恒为 string。

```go
func NewListKey[V any](ops ...Option) *ListKey[V]
func (c *ListKey[V]) ConcatKey(fields ...interface{}) *ListKey[V]

func (c *ListKey[V]) RPush(v ...V)  error
func (c *ListKey[V]) LPush(v ...V)  error
func (c *ListKey[V]) RPushX(v ...V) error                  // 仅当 list 已存在
func (c *ListKey[V]) LPushX(v ...V) error
func (c *ListKey[V]) RPop() (V, error)
func (c *ListKey[V]) LPop() (V, error)

func (c *ListKey[V]) BLPop(timeout time.Duration) (V, error)
func (c *ListKey[V]) BRPop(timeout time.Duration) (V, error)
func (c *ListKey[V]) BRPopLPush(dest string, timeout time.Duration) (V, error)

func (c *ListKey[V]) LRange(start, stop int64) ([]V, error)
func (c *ListKey[V]) LIndex(i int64)           (V, error)
func (c *ListKey[V]) LSet(i int64, v V)        error
func (c *ListKey[V]) LLen()                    (int64, error)
func (c *ListKey[V]) LTrim(start, stop int64)  error

func (c *ListKey[V]) LRem(count int64, v V)    error       // +N 头到尾 / -N 尾到头 / 0 全删
func (c *ListKey[V]) LInsertBefore(pivot, v V) error
func (c *ListKey[V]) LInsertAfter(pivot, v V)  error
func (c *ListKey[V]) Sort(sort *redis.Sort) ([]V, error)
```

- 💡 `BLPop/BRPop` 的 `timeout=0` = **永久阻塞**(Redis 语义)
- 💡 `LRange(0, -1)` 取全部;负索引从尾算起
- 💡 `LInsert*` 和 `LRem` 的 pivot/v **按 msgpack 字节匹配** —— 要传完全相同的 struct

---

<a id="setkey"></a>
## SetKey `[K comparable, V any]`

K 当前运行时未用(API 对称占位)。

```go
func NewSetKey[K comparable, V any](ops ...Option) *SetKey[K, V]
func (c *SetKey[K, V]) ConcatKey(fields ...interface{}) *SetKey[K, V]

func (c *SetKey[K, V]) SAdd(members ...V) error            // 变参
func (c *SetKey[K, V]) SRem(members ...V) error            // 变参
func (c *SetKey[K, V]) SIsMember(m V)     (bool, error)
func (c *SetKey[K, V]) SMembers()         ([]V, error)
func (c *SetKey[K, V]) SCard()            (int64, error)
func (c *SetKey[K, V]) SScan(cursor uint64, match string, count int64) ([]V, uint64, error)
```

- 💡 `SRem` 按序列化后字节匹配 —— 传的 struct 必须 msgpack 回完全相同字节才能命中

---

<a id="zsetkey"></a>
## ZSetKey `[K comparable, V any]`

```go
func NewZSetKey[K comparable, V any](ops ...Option) *ZSetKey[K, V]
func (c *ZSetKey[K, V]) ConcatKey(fields ...interface{}) *ZSetKey[K, V]

func (c *ZSetKey[K, V]) ZAdd(members ...redis.Z)            error
func (c *ZSetKey[K, V]) ZRem(members ...interface{})        error
func (c *ZSetKey[K, V]) ZIncrBy(inc float64, m interface{}) (float64, error)  // 返回新分

func (c *ZSetKey[K, V]) ZRemRangeByRank(start, stop int64) error
func (c *ZSetKey[K, V]) ZRemRangeByScore(min, max string)  error

func (c *ZSetKey[K, V]) ZRank(m interface{})       (int64, error)
func (c *ZSetKey[K, V]) ZRevRank(m interface{})    (int64, error)
func (c *ZSetKey[K, V]) ZScore(m interface{})      (float64, error)
func (c *ZSetKey[K, V]) ZCard()                    (int64, error)
func (c *ZSetKey[K, V]) ZCount(min, max string)    (int64, error)
func (c *ZSetKey[K, V]) ZLexCount(min, max string) (int64, error)

func (c *ZSetKey[K, V]) ZRange(start, stop int64)              ([]V, error)
func (c *ZSetKey[K, V]) ZRevRange(start, stop int64)           ([]V, error)
func (c *ZSetKey[K, V]) ZRangeWithScores(start, stop int64)    ([]V, []float64, error)
func (c *ZSetKey[K, V]) ZRevRangeWithScores(start, stop int64) ([]V, []float64, error)

func (c *ZSetKey[K, V]) ZRangeByScore(opt *redis.ZRangeBy)              ([]V, error)
func (c *ZSetKey[K, V]) ZRevRangeByScore(opt *redis.ZRangeBy)           ([]V, error)
func (c *ZSetKey[K, V]) ZRangeByScoreWithScores(opt *redis.ZRangeBy)    ([]V, []float64, error)
func (c *ZSetKey[K, V]) ZRevRangeByScoreWithScores(opt *redis.ZRangeBy) ([]V, []float64, error)

func (c *ZSetKey[K, V]) ZPopMax(count int64) ([]V, []float64, error)
func (c *ZSetKey[K, V]) ZPopMin(count int64) ([]V, []float64, error)

func (c *ZSetKey[K, V]) ZScan(cursor uint64, match string, count int64) ([]V, uint64, error)
```

- 💡 `ZAdd` 会**就地**改写传入切片 —— 把 `redis.Z.Member` 替成 msgpack 字节,原生客户端直接读会拿到二进制
- 💡 `ZIncrBy` 返回 +inc 之后的**新分**,不是 error-only
- 💡 `ZCount` / `ZLexCount` / `ZRemRangeByScore` 的 `min/max` 走 Redis 分数语法:`"-inf"`、`"+inf"`、`"(1.0"`(排他)
- 💡 `ZRem` 内部用 Pipeline 逐条 ZREM,不是单命令多 member

---

<a id="streamkey"></a>
## StreamKey `[K comparable, V any]`

薄包装,**不对条目内字段做类型检查** —— 自己用 `redis.XAddArgs.Values` 装。

```go
func NewStreamKey[K comparable, V any](ops ...Option) *StreamKey[K, V]
func (c *StreamKey[K, V]) ConcatKey(fields ...interface{}) *StreamKey[K, V]

func (c *StreamKey[K, V]) XAdd(args *redis.XAddArgs)              (string, error)  // 返回 entry ID
func (c *StreamKey[K, V]) XDel(ids ...string)                      (int64, error)
func (c *StreamKey[K, V]) XLen()                                   (int64, error)
func (c *StreamKey[K, V]) XRange(start, stop string)               ([]redis.XMessage, error)
func (c *StreamKey[K, V]) XRangeN(start, stop string, n int64)     ([]redis.XMessage, error)
func (c *StreamKey[K, V]) XRevRange(start, stop string)            ([]redis.XMessage, error)
func (c *StreamKey[K, V]) XRevRangeN(start, stop string, n int64)  ([]redis.XMessage, error)
func (c *StreamKey[K, V]) XRead(args *redis.XReadArgs)             ([]redis.XStream, error)
```

- 💡 `XAdd` 会**覆盖** `args.Stream` 为 ctx.Key —— 不用自己填
- 💡 `XRead` 若 `args.Streams` 为空,默认 `[ctx.Key, "$"]`(只读新增)
- 💡 `start/stop` 走 stream ID 语法:`"-"` `"+"` 或 `"<ms>-<seq>"`

---

<a id="vectorsetkey"></a>
## VectorSetKey `[K comparable, V any]`

原生 `FT.*` 透传。索引名 = ctx.Key。V 必须是 struct(带 `json:"…"`)或 `map[string]interface{}`。

```go
func NewVectorSetKey[K comparable, V any](ops ...Option) *VectorSetKey[K, V]
func (c *VectorSetKey[K, V]) ConcatKey(fields ...interface{}) *VectorSetKey[K, V]

func (c *VectorSetKey[K, V]) Create(args ...interface{}) error    // FT.CREATE 尾段,原样透传
func (c *VectorSetKey[K, V]) DropIndex(deleteDocs bool) error     // true 时追加 "DD"
func (c *VectorSetKey[K, V]) Info()                  (map[string]interface{}, error)
func (c *VectorSetKey[K, V]) TagVals(fieldName string) ([]string, error)

func (c *VectorSetKey[K, V]) AliasAdd(alias string)    error
func (c *VectorSetKey[K, V]) AliasUpdate(alias string) error
func (c *VectorSetKey[K, V]) AliasDel(alias string)    error

func (c *VectorSetKey[K, V]) Search(query string, params ...interface{}) (count int64, docs []V, err error)

func (c *VectorSetKey[K, V]) Float32ToBytes(v []float32) []byte
func (c *VectorSetKey[K, V]) BytesToFloat32(b []byte)    ([]float32, error)
func (c *VectorSetKey[K, V]) KNNParamHelper(k int, field string, vec []float32) (string, []interface{})
```

- 💡 `Search` 返回的 `count` 是服务端总匹配数,**不是 `len(docs)`**(分页时不等)
- 💡 文档解析走 JSON round-trip,V 的 tag 必须是 `json:"…"`,**不是** `msgpack:"…"`
- 💡 `KNNParamHelper` 返回 `(queryFragment, params)`,自己拼 `"*=>" + frag` 再传给 `Search`

最小 KNN 工作流:

```go
// 1. 建索引(FT.CREATE 尾段原样透传)
idx.Create(
    "ON", "HASH", "PREFIX", "1", "doc:",
    "SCHEMA",
    "title", "TEXT",
    "vec",   "VECTOR", "HNSW", "6", "TYPE", "FLOAT32", "DIM", "128", "DISTANCE_METRIC", "L2",
)

// 2. KNN 查询(写入用普通 Redis HSET,这里只演示查询)
frag, params := idx.KNNParamHelper(10, "vec", embedding)
count, docs, err := idx.Search("*=>"+frag, params...)
```

---

<a id="searchkey"></a>
## SearchKey `[K comparable, V any]`

VectorSetKey 之上的**自动建索引 + 类型化 KNN** 封装,RAG / AI 场景用。构造时执行 `EnsureIndex()`,幂等。

```go
func NewSearchKey[K comparable, V any](indexName string, ops ...Option) *SearchKey[K, V]
func (c *SearchKey[K, V]) EnsureIndex() error           // 据 V 的 tag 反射建索引,幂等

func (c *SearchKey[K, V]) Put(id K, doc V) error        // struct 打散为 hash 字段,向量转 BLOB
func (c *SearchKey[K, V]) Search(query string, opts ...SearchOption) ([]V, int64, error)
func (c *SearchKey[K, V]) VectorSearch(field string, vec []float32, topK int) ([]V, []float64, error)

// SearchOption 拼装器
func SearchLimit(offset, num int)              SearchOption
func SearchSortBy(field string, asc bool)      SearchOption
func SearchReturn(fields ...string)            SearchOption
func SearchHighlight(openTag, closeTag string) SearchOption
func SearchVerbatim()                          SearchOption
func SearchWithScores()                        SearchOption
```

- 💡 `indexName` 是**第一个位置参数**,不是通过 `WithKey` 传
- 💡 `Put` 反射拆 struct 为多个 hash 字段以适配 RediSearch 倒排索引 —— **存储格式和其他 Key 类型不兼容**,不能用 `HashKey` 读
- 💡 走 `DIALECT 2`,返回值经 JSON round-trip 还原为 `[]V`
- 💡 `Search` 返回顺序 `([]V, total, err)` —— **total 在第二位**(VectorSetKey 在第一位,别搞反)

---

<a id="httpon"></a>
## HttpOn 权限位 (`http_whitelist.go`)

```go
key.HttpOn(redisdb.HashRead | redisdb.HashWrite)   // 或直接 HashAll
```

| 类型 | 读掩码 | 写掩码 | 全掩码 | 单独位 |
| --- | --- | --- | --- | --- |
| Hash | `HashRead` | `HashWrite` | `HashAll` | `HGet HSet HDel HMGET HExists HGetAll HRandField HRandFieldWithValues HLen HKeys HVals HIncrBy HIncrByFloat HSetNX HScan` |
| List | `ListRead` | `ListWrite` | `ListAll` | `RPush RPushX LPush LPushX RPop LPop LRange LRem LSet LIndex LTrim LLen` |
| Set | `SetRead` | `SetWrite` | `SetAll` | `SAdd SCard SRem SIsMember SMembers SScan` |
| ZSet | `ZSetRead` | `ZSetWrite` | `ZSetAll` | `ZAdd ZRem ZRange ZRank ZScore ZCard ZCount ZIncrBy ZScan ZRangeByScore ZRevRange ZRevRangeByScore ZRemRangeByScore ZRangeWithScores ZRevRangeWithScores` |
| String | `StringRead` | `StringWrite` | `StringAll` | `Get Set StringGetAll StringSetAll` |
| Stream | `StreamRead` | `StreamWrite` | `StreamAll` | `XAdd XDel XRange XLen XRead XTrim XInfo` |
| VectorSet | `VectorSetRead` | `VectorSetWrite` | `VectorSetAll` | `FtCreate FtSearch FtAggregate FtDropIndex FtTagVals FtInfo` |
| 通用 | `CommonRead` | `CommonWrite` | — | `Del Exists Expire Persist TTL Type Rename` |
| 系统 | — | — | — | `DBTime DBKeys`(用 `AllowDBOp` / `IsAllowedDBOp`) |

校验函数(注意是 `IsAllowed…`,**有 -ed**):

```go
redisdb.IsAllowedHashOp(key, redisdb.HGet)
redisdb.IsAllowedListOp(key, redisdb.RPush)
redisdb.IsAllowedSetOp(key, redisdb.SAdd)
redisdb.IsAllowedZSetOp(key, redisdb.ZAdd)
redisdb.IsAllowedStringOp(key, redisdb.Get)
redisdb.IsAllowedStreamOp(key, redisdb.XAdd)
redisdb.IsAllowedVectorSetOp(key, redisdb.FtSearch)
redisdb.IsAllowedCommon(key, redisdb.Del)       // 通用位
redisdb.IsAllowedDBOp(redisdb.DBKeys)           // 系统位
```

**权限按 key 前缀(第一个 `:` 之前)聚合** —— `user:profile` 和 `user:settings` 共用一份掩码。
