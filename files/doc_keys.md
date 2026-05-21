---
slug: data-keys
title: RedisDB Keys Reference
---

<a id="top"></a>

# RedisDB Keys Reference

`github.com/doptime/redisdb` —— 8 种类型化 Redis Key,统一签名 `*XxxKey[K, V]`,
值用 msgpack 编解。本文档单文件、显式锚点,README 可深链。

* [1. StringKey](#stringkey) — 一键一值,带 TTL
* [2. HashKey](#hashkey) — 类型化 map,自动主键
* [3. ListKey](#listkey) — 队列、阻塞弹出
* [4. SetKey](#setkey) — 去重集合
* [5. ZSetKey](#zsetkey) — 排行榜,按分数/排名查
* [6. StreamKey](#streamkey) — 事件流
* [7. VectorSetKey](#vectorsetkey) — 原生 `FT.*`
* [8. SearchKey](#searchkey) — 自动建索引 + KNN(AI 场景)
* [附 A: 公共契约 / 选项 / 修饰符](#common)
* [附 B: HttpOn 权限位](#op-constants)
* [附 C: 旧文档迁移与已修正项](#migration)

---

<a id="common"></a>

## 附 A · 公共契约

### A.1 构造

所有 `NewXxxKey` 收 `...Option`,失败返回 `nil`(错只写日志,**调用方要 nil 检**)。
三个 helper,可链:

```go
redisdb.WithKey("users")                          // Redis key/前缀;省略则取 V 的类型名
redisdb.WithRds("cache")                          // 选 config.toml 里 [[Redis]] 的 Name,默认 "default"
redisdb.WithModifier(map[string]ModifierFunc{…})  // 注册额外 mod: 指令
```

### A.2 派生 key

每种 Key 都有 `ConcatKey(fields ...interface{}) *Self` —— **返回新实例**,原 ctx 不变,
新 key = `oldKey:field1:field2…`。时间分桶 helper:

```go
redisdb.CatYearMonthDay(t)  // "YMD_20260521"
redisdb.CatYearMonth(t)     // "YM_202605"
redisdb.CatYear(t)          // "Y_2026"
redisdb.CatYearWeek(t)      // "YW_202621"  (ISO week)
```

### A.3 Struct tag

| tag | 用途 |
| --- | --- |
| `msgpack:"…"` | 存储编解(几乎所有类型) |
| `json:"…"` | `VectorSetKey` / `SearchKey` 字段映射 |
| `mod:"…"` | 保存前修饰,见 A.4 |
| `validate:"…"` | go-playground/validator 校验 |

### A.4 mod 指令

`tag_modifier.go` 注册表实际只有 8 个,**没有 `now`**:

| 指令 | 行为 | 参数 |
| --- | --- | --- |
| `trim` `lowercase` `uppercase` `title` | 字符串变换 | — |
| `default=X` | 零值时填 X(按字段类型解析) | 必需 |
| `unixtime` | 当前 Unix 秒(int64) | 可选 `=ms` 毫秒 |
| `counter` | 数值字段 +1(所有 int/uint/float 宽度) | — |
| `nanoid` | 生成 NanoID | 可选 `=N`,默认 21 |

**默认只在字段为零值时跑**,加 `,force` 后无条件跑。多个用 `,` 串接,从左到右执行。
手动触发:`redisdb.ApplyModifiers(&v)`。

### A.5 错误约定

* `redis.Nil` = key/字段不存在,用 `errors.Is` 区分
* 返回零值 + err 时,**绝不返回半填的 struct**
* `HGetAll` / `GetAll` / `HMGET` / `*Scan` 系列里**单条**解码失败会被静默跳过(只写日志)

### A.6 config.toml

```toml
[[Redis]]
Name = "default"
Host = "127.0.0.1"
Port = 6379
DB   = 0
```

[↑](#top)

---

<a id="stringkey"></a>

## 1 · StringKey `[K comparable, V any]`

一个 ctx 对应**一族**独立 Redis string,实际键名 = `ctx.Key + ":" + serialize(K)`。
适合配置、缓存、计数。

```go
func NewStringKey[K comparable, V any](ops ...Option) *StringKey[K, V]
func (c *StringKey[K, V]) ConcatKey(fields ...interface{}) *StringKey[K, V]

// CRUD
func (c *StringKey[K, V]) Set(key K, value V, expiration time.Duration) error
func (c *StringKey[K, V]) Get(field K) (V, error)
func (c *StringKey[K, V]) Del(key K) error

// 批量
func (c *StringKey[K, V]) GetAll(match string) (map[K]V, error)   // SCAN+GET,1GiB 上限,别在热路径用
func (c *StringKey[K, V]) SetAll(m map[K]V) error                  // Pipeline,会清掉已有 TTL

// 来自基类
func (c *StringKey[K, V]) Scan(cursor uint64, match string, count int64) ([]string, uint64, error)
func (c *StringKey[K, V]) Keys() ([]K, error)
func (c *StringKey[K, V]) HttpOn(op StringOp) *StringKey[K, V]
```

* 💡 `Set` 的 `expiration`:`0` = 永不过期,**负值** = 清掉已有 TTL(go-redis 语义)
* 💡 `SetAll` 内部固定传 `-1`,会**清掉同名 key 原有的过期时间**

[↑](#top)

---

<a id="hashkey"></a>

## 2 · HashKey `[K comparable, V any]`

一个 ctx 对应**一个** Redis hash,field=K, value=V。

```go
func NewHashKey[K comparable, V any](ops ...Option) *HashKey[K, V]
func (c *HashKey[K, V]) ConcatKey(fields ...interface{}) *HashKey[K, V]

// 写
func (c *HashKey[K, V]) HSet(values ...interface{}) (int64, error)  // (k,v,k,v,...) 或单个 map[K]V
func (c *HashKey[K, V]) HMSet(kvMap map[K]V)        (int64, error)
func (c *HashKey[K, V]) HSetNX(field K, value V)    error
func (c *HashKey[K, V]) Save(value V)               (int64, error)  // 自动从 struct 取 field
func (c *HashKey[K, V]) HDel(fields ...K)           error
func (c *HashKey[K, V]) HIncrBy(field K, inc int64)    error
func (c *HashKey[K, V]) HIncrByFloat(field K, inc float64) error

// 读
func (c *HashKey[K, V]) HGet(field K)               (V, error)
func (c *HashKey[K, V]) HMGET(fields ...interface{}) ([]V, error)
func (c *HashKey[K, V]) HGetAll()                   (map[K]V, error)
func (c *HashKey[K, V]) HExists(field K)            (bool, error)
func (c *HashKey[K, V]) HKeys()                     ([]K, error)
func (c *HashKey[K, V]) HVals()                     ([]V, error)
func (c *HashKey[K, V]) HLen()                      (int64, error)

// 采样 / 迭代
func (c *HashKey[K, V]) HRandField(count int)           ([]K, error)
func (c *HashKey[K, V]) HRandFieldWithValues(count int) ([]K, []V, error)
func (c *HashKey[K, V]) HScan(cursor uint64, match string, count int64)        ([]K, []V, uint64, error)
func (c *HashKey[K, V]) HScanNoValues(cursor uint64, match string, count int64) ([]K, uint64, error)
```

* 💡 `HSet` 散参格式**必须**偶数对,且 `k,v` 类型严格对齐 `K,V`,否则运行时报错
* 💡 `HSet` 返回的 `int64` 是**新增**字段数(Redis 语义),覆写的不算
* 💡 `Save` 在构造时自省 `V`,找第一个类型可赋给 `K` 的字段记下来当主键;找不到就退化为单参 `HSet`
* 💡 `HIncrBy` 直接操作字段裸字节 —— 这个字段必须是数字字符串,**不能是 msgpack blob**
* 💡 `HDel`:`K` 是 string 直传;非 string 走 JSON 序列化(和写入时一致)
* 💡 `HRandField` 的 `count`:正数=去重,上限为 hash 大小;负数=可重复,正好 `|count|` 条

[↑](#top)

---

<a id="listkey"></a>

## 3 · ListKey `[V any]`

**只有一个泛型参数** —— K 恒为 string。

```go
func NewListKey[V any](ops ...Option) *ListKey[V]
func (c *ListKey[V]) ConcatKey(fields ...interface{}) *ListKey[V]

// 压入/弹出
func (c *ListKey[V]) RPush(v ...V)  error
func (c *ListKey[V]) LPush(v ...V)  error
func (c *ListKey[V]) RPushX(v ...V) error    // 仅当 list 已存在
func (c *ListKey[V]) LPushX(v ...V) error
func (c *ListKey[V]) RPop() (V, error)
func (c *ListKey[V]) LPop() (V, error)

// 阻塞
func (c *ListKey[V]) BLPop(timeout time.Duration) (V, error)
func (c *ListKey[V]) BRPop(timeout time.Duration) (V, error)
func (c *ListKey[V]) BRPopLPush(dest string, timeout time.Duration) (V, error)

// 索引 / 范围
func (c *ListKey[V]) LRange(start, stop int64) ([]V, error)
func (c *ListKey[V]) LIndex(i int64)           (V, error)
func (c *ListKey[V]) LSet(i int64, v V)        error
func (c *ListKey[V]) LLen()                    (int64, error)
func (c *ListKey[V]) LTrim(start, stop int64)  error

// 改写
func (c *ListKey[V]) LRem(count int64, v V)    error    // +N 头到尾 / -N 尾到头 / 0 全删
func (c *ListKey[V]) LInsertBefore(pivot, v V) error
func (c *ListKey[V]) LInsertAfter(pivot, v V)  error
func (c *ListKey[V]) Sort(sort *redis.Sort)   ([]V, error)
```

* 💡 `BLPop/BRPop` 的 `timeout=0` = **永久阻塞**(Redis 语义)
* 💡 `LRange(0, -1)` 取全部;负索引从尾算起
* 💡 `LInsert*` 和 `LRem` 的 pivot/v **按 msgpack 字节匹配** —— 要传完全相同的 struct

[↑](#top)

---

<a id="setkey"></a>

## 4 · SetKey `[K comparable, V any]`

K 当前运行时未用(API 对称占位)。

```go
func NewSetKey[K comparable, V any](ops ...Option) *SetKey[K, V]
func (c *SetKey[K, V]) ConcatKey(fields ...interface{}) *SetKey[K, V]

func (c *SetKey[K, V]) SAdd(members ...V) error    // 变参
func (c *SetKey[K, V]) SRem(members ...V) error    // 变参
func (c *SetKey[K, V]) SIsMember(m V)     (bool, error)
func (c *SetKey[K, V]) SMembers()         ([]V, error)
func (c *SetKey[K, V]) SCard()            (int64, error)
func (c *SetKey[K, V]) SScan(cursor uint64, match string, count int64) ([]V, uint64, error)
```

* 💡 `SRem` 按序列化后字节匹配 —— 删除时传的 struct 必须能 msgpack 回完全相同的字节

[↑](#top)

---

<a id="zsetkey"></a>

## 5 · ZSetKey `[K comparable, V any]`

```go
func NewZSetKey[K comparable, V any](ops ...Option) *ZSetKey[K, V]
func (c *ZSetKey[K, V]) ConcatKey(fields ...interface{}) *ZSetKey[K, V]

// 写
func (c *ZSetKey[K, V]) ZAdd(members ...redis.Z)            error
func (c *ZSetKey[K, V]) ZRem(members ...interface{})        error
func (c *ZSetKey[K, V]) ZIncrBy(inc float64, m interface{}) (float64, error)  // 返回新分

// 区间删除
func (c *ZSetKey[K, V]) ZRemRangeByRank(start, stop int64) error
func (c *ZSetKey[K, V]) ZRemRangeByScore(min, max string)  error

// 排名 / 分数 / 计数
func (c *ZSetKey[K, V]) ZRank(m interface{})       (int64, error)
func (c *ZSetKey[K, V]) ZRevRank(m interface{})    (int64, error)
func (c *ZSetKey[K, V]) ZScore(m interface{})      (float64, error)
func (c *ZSetKey[K, V]) ZCard()                    (int64, error)
func (c *ZSetKey[K, V]) ZCount(min, max string)    (int64, error)
func (c *ZSetKey[K, V]) ZLexCount(min, max string) (int64, error)

// 按排名取
func (c *ZSetKey[K, V]) ZRange(start, stop int64)              ([]V, error)
func (c *ZSetKey[K, V]) ZRevRange(start, stop int64)           ([]V, error)
func (c *ZSetKey[K, V]) ZRangeWithScores(start, stop int64)    ([]V, []float64, error)
func (c *ZSetKey[K, V]) ZRevRangeWithScores(start, stop int64) ([]V, []float64, error)

// 按分数取
func (c *ZSetKey[K, V]) ZRangeByScore(opt *redis.ZRangeBy)              ([]V, error)
func (c *ZSetKey[K, V]) ZRevRangeByScore(opt *redis.ZRangeBy)           ([]V, error)
func (c *ZSetKey[K, V]) ZRangeByScoreWithScores(opt *redis.ZRangeBy)    ([]V, []float64, error)
func (c *ZSetKey[K, V]) ZRevRangeByScoreWithScores(opt *redis.ZRangeBy) ([]V, []float64, error)

// 两端弹出
func (c *ZSetKey[K, V]) ZPopMax(count int64) ([]V, []float64, error)
func (c *ZSetKey[K, V]) ZPopMin(count int64) ([]V, []float64, error)

// 迭代
func (c *ZSetKey[K, V]) ZScan(cursor uint64, match string, count int64) ([]V, uint64, error)
```

* 💡 `ZAdd` 会**就地**改写传入切片 —— 把 `redis.Z.Member` 替成 msgpack 字节。原生客户端直接读这个 key 会拿到二进制不可读
* 💡 `ZIncrBy` 返回的是 **+inc 之后的新分**,不是 error-only
* 💡 `ZCount/ZLexCount/ZRemRangeByScore` 的 `min/max` 走 Redis 分数语法:`"-inf"`、`"+inf"`、`"(1.0"`(排他)
* 💡 `ZRem` 内部用 Pipeline 逐条删,不是单条 ZREM 批操作

[↑](#top)

---

<a id="streamkey"></a>

## 6 · StreamKey `[K comparable, V any]`

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

* 💡 `XAdd` 会**覆盖** `args.Stream` 为 ctx.Key —— 不用自己填
* 💡 `XRead` 若 `args.Streams` 为空,默认 `[ctx.Key, "$"]`(只读新增)
* 💡 `start/stop` 走 stream ID 语法:`"-"` `"+"` 或 `"<ms>-<seq>"`

[↑](#top)

---

<a id="vectorsetkey"></a>

## 7 · VectorSetKey `[K comparable, V any]`

原生 `FT.*` 透传。索引名 = ctx.Key。V 必须是 struct(带 `json:"…"`)或 `map[string]interface{}`。

```go
func NewVectorSetKey[K comparable, V any](ops ...Option) *VectorSetKey[K, V]
func (c *VectorSetKey[K, V]) ConcatKey(fields ...interface{}) *VectorSetKey[K, V]

// 索引生命周期
func (c *VectorSetKey[K, V]) Create(args ...interface{}) error    // FT.CREATE 尾段,原样透传
func (c *VectorSetKey[K, V]) DropIndex(deleteDocs bool) error     // true 时追加 "DD"
func (c *VectorSetKey[K, V]) Info() (map[string]interface{}, error)
func (c *VectorSetKey[K, V]) TagVals(fieldName string) ([]string, error)

// 别名
func (c *VectorSetKey[K, V]) AliasAdd(alias string)    error
func (c *VectorSetKey[K, V]) AliasUpdate(alias string) error
func (c *VectorSetKey[K, V]) AliasDel(alias string)    error

// 查询
func (c *VectorSetKey[K, V]) Search(query string, params ...interface{}) (count int64, docs []V, err error)

// 向量工具
func (c *VectorSetKey[K, V]) Float32ToBytes(v []float32)  []byte
func (c *VectorSetKey[K, V]) BytesToFloat32(b []byte)     ([]float32, error)
func (c *VectorSetKey[K, V]) KNNParamHelper(k int, field string, vec []float32) (string, []interface{})
```

* 💡 `Search` 返回的 `count` 是服务端总匹配数,**不是 `len(docs)`**(分页时不等)
* 💡 文档解析走 **JSON round-trip**,V 的 tag 必须是 `json:"…"`,**不是** `msgpack:"…"`
* 💡 `KNNParamHelper` 返回 `(queryFragment, params)`,自己拼 `"*=>" + frag` 再传给 `Search`

[↑](#top)

---

<a id="searchkey"></a>

## 8 · SearchKey `[K comparable, V any]`

VectorSetKey 之上的**自动建索引 + 类型化 KNN** 封装,专给 RAG / AI 场景。
构造时就 `EnsureIndex()`,幂等。

```go
func NewSearchKey[K comparable, V any](indexName string, ops ...Option) *SearchKey[K, V]
func (c *SearchKey[K, V]) EnsureIndex() error              // 据 V 的 tag 反射建索引,幂等

func (c *SearchKey[K, V]) Put(id K, doc V) error           // struct 打散为 hash 字段,向量转 BLOB
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

* 💡 注意构造签名:`indexName` 是**第一个位置参数**,不是通过 `WithKey` 传
* 💡 `Put` 反射拆 struct 为多个 hash 字段,以满足 RediSearch 倒排索引扫描需求 —— **存储格式和其他 Key 类型不兼容**,不能用 `HashKey` 去读 `SearchKey.Put` 写的数据
* 💡 走 `DIALECT 2`,返回值经 JSON round-trip 还原为 `[]V`
* 💡 `Search` 返回顺序 `([]V, total, err)` —— **total 在第二位**(VectorSetKey 在第一位,别搞反)

[↑](#top)

---

<a id="op-constants"></a>

## 附 B · HttpOn 权限位 (`http_whitelist.go`)

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
| 系统 | — | — | — | `DBTime DBKeys` (用 `AllowDBOp` / `IsAllowedDBOp`) |

校验函数(注意是 `IsAllowed…`,**有 -ed**,旧 cookbook 写错过):

```go
redisdb.IsAllowedHashOp(key, redisdb.HGet)
redisdb.IsAllowedListOp(key, redisdb.RPush)
redisdb.IsAllowedSetOp(key, redisdb.SAdd)
redisdb.IsAllowedZSetOp(key, redisdb.ZAdd)
redisdb.IsAllowedStringOp(key, redisdb.Get)
redisdb.IsAllowedStreamOp(key, redisdb.XAdd)
redisdb.IsAllowedVectorSetOp(key, redisdb.FtSearch)
redisdb.IsAllowedCommon(key, redisdb.Del)        // 通用位
redisdb.IsAllowedDBOp(redisdb.DBKeys)            // 系统位
```

**权限按 key 前缀(第一个 `:` 之前)聚合** —— `user:profile` 和 `user:settings` 共用一份掩码。

[↑](#top)

---

<a id="migration"></a>

## 附 C · 旧文档迁移与已修正项

| 旧文件 | 新锚点 |
| --- | --- |
| `doc_stringkey.md` | [`#stringkey`](#stringkey) |
| `doc_hashkey.md`   | [`#hashkey`](#hashkey) |
| `doc_listKey.md`   | [`#listkey`](#listkey) |
| `doc_setkey.md`    | [`#setkey`](#setkey) |
| `doc_zsetkey.md`   | [`#zsetkey`](#zsetkey) |
| `doc_mod_example.md` | [`#common`](#common) §A.4 |
| (无旧文档) | [`#streamkey`](#streamkey) / [`#vectorsetkey`](#vectorsetkey) / [`#searchkey`](#searchkey) 全新 |

**已修正的签名 / 写错**

* 所有 `New…Key` 参数:旧 `...opSetter` → 实际 `...Option`(`opSetter` 这个类型不存在)
* `NewListKey`:旧 `[K, V]` → 实际 `[V]`(K 恒 string)
* `SetKey.SAdd/SRem`:旧单参 → 实际 `(members ...V)` 变参
* `HashKey.HSet/HMSet`:旧 `error` → 实际 `(int64, error)`
* `ZSetKey.ZIncrBy`:旧 `error` → 实际 `(float64, error)` 返回新分
* 权限校验:`IsAllowHashOp` → `IsAllowedHashOp`(有 `-ed`)
* mod 指令:`now` **不存在**,删掉

**之前漏文档的方法**(已并入对应章节)

* HashKey: `Save HSetNX HIncrBy HIncrByFloat HMGET HRandField HRandFieldWithValues HScan HScanNoValues`
* ListKey: `RPushX LPushX LSet LIndex LLen BLPop BRPop BRPopLPush LInsertBefore LInsertAfter Sort`
* ZSetKey: `ZRevRange ZRevRank ZScore ZPopMax ZPopMin ZScan ZRemRangeByScore ZRangeByScoreWithScores ZRevRangeByScore ZRevRangeByScoreWithScores ZLexCount`
* StreamKey / VectorSetKey / SearchKey: 整章全新

[↑](#top)
