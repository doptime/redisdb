# RedisDB

类型化 Redis 包装。`github.com/doptime/redisdb`,8 种 Key,统一 `*XxxKey[K, V]`,
msgpack 自动编解,带 mod 修饰符和 HTTP 暴露。

```bash
go get github.com/doptime/redisdb
```

```go
type User struct {
    UID  string `msgpack:"uid"  mod:"trim,lowercase"`
    Name string `msgpack:"name" mod:"trim"`
}
users := redisdb.NewHashKey[string, *User](redisdb.WithKey("users"))
_, _ = users.HSet("u1", &User{UID: "U1 ", Name: " Alice "})  // mod 在写入前跑
u, _ := users.HGet("u1")                                      // u 是 *User
```

## 文档

全部 API 见 [`doc_keys.md`](doc_keys.md),单文件、显式锚点。

| Key 类型 | 锚点 | 适用 |
| --- | --- | --- |
| `StringKey[K,V]`    | [`#stringkey`](doc_keys.md#stringkey)       | 一键一值带 TTL |
| `HashKey[K,V]`      | [`#hashkey`](doc_keys.md#hashkey)           | 类型化 map |
| `ListKey[V]`        | [`#listkey`](doc_keys.md#listkey)           | 队列、阻塞弹出 |
| `SetKey[K,V]`       | [`#setkey`](doc_keys.md#setkey)             | 去重集合 |
| `ZSetKey[K,V]`      | [`#zsetkey`](doc_keys.md#zsetkey)           | 排行榜、按分数/排名查 |
| `StreamKey[K,V]`    | [`#streamkey`](doc_keys.md#streamkey)       | 事件流 |
| `VectorSetKey[K,V]` | [`#vectorsetkey`](doc_keys.md#vectorsetkey) | 原生 `FT.*` |
| `SearchKey[K,V]`    | [`#searchkey`](doc_keys.md#searchkey)       | 自动建索引 + KNN(RAG/AI) |

通用部分:
[选项](doc_keys.md#common) · [mod 修饰符](doc_keys.md#common) · [ConcatKey](doc_keys.md#common) · [HTTP 权限位](doc_keys.md#op-constants) · [config.toml](doc_keys.md#common)

旧文档(`doc_stringkey.md` / `doc_hashkey.md` / `doc_listKey.md` / `doc_setkey.md` /
`doc_zsetkey.md` / `doc_mod_example.md` / Cookbook)全部并入 `doc_keys.md`。
迁移对照 + 已修正项见 [`#migration`](doc_keys.md#migration)。
