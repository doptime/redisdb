## RedisDB Documentation

RedisDB is a Redis based database package in Golang. 
The basic idea beyond is to making code lean.
It provides a high-level concise abstraction for various Redis data structures, hashes, lists, strings... By leveraging generics and the msgpack library, RedisDB ensures efficient serialization and deserialization of data, reducing the likelihood of errors and making your code more maintainable.

### Key Features

- **Type Safety:** RedisDB uses generics to specify key and value types, preventing type mismatches and ensuring that your data is handled correctly.
- **Automatic Serialization:** Data is automatically serialized using msgpack, eliminating the need for manual marshaling and unmarshaling.
- **Modifiers and Data Validation:** Supports modifiers for fields in your structs, allowing for tasks like trimming spaces, converting to lowercase, and setting default values during serialization and deserialization.
- **Documentation Generation:** Automatically generates documentation for your data structures, making it easier to understand and use them in your application.

### Getting Started

To get started with RedisDB, you need to have Redis installed and running. Additionally, ensure that you have the necessary dependencies installed in your Go project.

### Installation

You can install the RedisDB package using the following command:

```bash
go get github.com/doptime/redisdb
```

### Configuration

Before using RedisDB, you need to configure your Redis data sources. This can be done using the `cfgredis` package.

### Example Usage

#### HashKey

HashKey is used to interact with Redis hash keys, which are maps between string fields and string values.

```go
package main

import (
    "fmt"
    "github.com/doptime/redisdb"
)

type User struct {
    ID   string `msgpack:"id"`
    Name string `msgpack:"name"`
}

func main() {
    keyUser :=  redisdb.NewHashKey[string, *User]( redisdb.WithKey("users"))

    user := &User{ID: "1", Name: "Alice"}
    keyUser.HSet("1", user)

    retrievedUser, err := keyUser.HGet("1")
    if err == nil {
        fmt.Println("Retrieved User:", retrievedUser.Name)
    }
}
```

#### ListKey

ListKey is designed for working with Redis list keys, which are ordered collections of strings.

```go
package main

import (
    "fmt"
    "github.com/doptime/redisdb"
)

type Item struct {
    ID   string `msgpack:"id"`
    Name string `msgpack:"name"`
}

func main() {
    keyList :=  redisdb.NewListKey[string, *Item]( redisdb.WithKey("items"))

    item := &Item{ID: "1", Name: "Item1"}
    keyList.RPush(item)

    poppedItem, err := keyList.RPop()
    if err == nil {
        fmt.Println("Popped Item:", poppedItem.Name)
    }
}
```

#### StringKey

StringKey is used for simple key-value pairs in Redis.

```go
package main

import (
    "fmt"
    "github.com/doptime/redisdb"
)

type Config struct {
    Host     string `msgpack:"host"`
    Port     int    `msgpack:"port"`
    Enabled  bool   `msgpack:"enabled"`
}

func main() {
    keyConfig :=  redisdb.NewStringKey[string, *Config]( redisdb.WithKey("configs"))

    config := &Config{Host: "example.com", Port: 8080, Enabled: true}
    keyConfig.Set("config1", config, time.Hour*24)

    retrievedConfig, err := keyConfig.Get("config1")
    if err == nil {
        fmt.Println("Retrieved Config:", retrievedConfig.Host, retrievedConfig.Port, retrievedConfig.Enabled)
    }
}
```

#### SetKey

SetKey is for managing Redis set keys, which store unique unordered strings.

```go
package main

import (
    "fmt"
    "github.com/doptime/redisdb"
)

type Tag struct {
    ID   string `msgpack:"id"`
    Name string `msgpack:"name"`
}

func main() {
    keyTag :=  redisdb.NewSetKey[string, *Tag]( redisdb.WithKey("tags"))

    tag := &Tag{ID: "1", Name: "Technology"}
    keyTag.SAdd(tag)

    tags, err := keyTag.SMembers()
    if err == nil {
        for _, tag := range tags {
            fmt.Println("Tag ID:", tag.ID, "Name:", tag.Name)
        }
    }
}
```

#### ZSetKey

ZSetKey is for working with Redis sorted sets, which are collections of unique strings with associated scores.

```go
package main

import (
    "fmt"
    "github.com/doptime/redisdb"
)

type ScoredItem struct {
    ID   string  `msgpack:"id"`
    Score float64 `msgpack:"score"`
}

func main() {
    keyScoredItem :=  redisdb.NewZSetKey[string, *ScoredItem]( redisdb.WithKey("scored_items"))

    item := &ScoredItem{ID: "1", Score: 3.5}
    keyScoredItem.ZAdd( redisdb.redis.Z{Member: item, Score: item.Score})

    items, err := keyScoredItem.ZRange(0, -1)
    if err == nil {
        for _, item := range items {
            fmt.Println("Item ID:", item.ID, "Score:", item.Score)
        }
    }
}
```

### Documentation

For detailed documentation on each data structure type, refer to the respective documentation files:

- [HashKey Documentation](doc_hashkey.md)
- [ListKey Documentation](doc_listKey.md)
- [SetKey Documentation](doc_setkey.md)
- [StringKey Documentation](doc_stringkey.md)
- [ZSetKey Documentation](doc_zsetkey.md)

### Error Handling

Most methods in RedisDB return errors that you should check to handle failures gracefully. Always ensure to handle errors to maintain the robustness of your application.

### Conclusion

RedisDB provides a powerful and type-safe way to interact with Redis in Go. By using generics and automatic serialization, it reduces the complexity and potential errors in your code. Explore the various data structures and their methods to fully leverage the capabilities of Redis in your applications.