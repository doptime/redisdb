## RedisDB Cookbook

This cookbook provides concise, step-by-step examples for using RedisDB with different data structures. RedisDB is a Go package that offers type-safe interactions with Redis, automatic serialization, and built-in modifiers for data validation.

## StringKey Example

```go
type Config struct {
    Host     string `msgpack:"host" mod:"trim,lowercase"`
    Port     int    `mod:"default=8080"`
    Enabled  bool   `mod:"default=true"`
}

keyConfig := redisdb.NewStringKey[string, *Config](redisdb.WithKey("configs"))

// Set a configuration
config := &Config{Host: "example.com", Port: 8080, Enabled: true}
keyConfig.Set("config1", config, time.Hour*24)

// Get the configuration
retrievedConfig, err := keyConfig.Get("config1")
if err == nil {
    fmt.Println("Retrieved Config:", retrievedConfig.Host, retrievedConfig.Port, retrievedConfig.Enabled)
}
```

## HashKey Example

```go
type User struct {
    ID   string `msgpack:"id" mod:"trim,lowercase"`
    Name string `msgpack:"name" mod:"trim"`
}

keyUser := redisdb.NewHashKey[string, *User](redisdb.WithKey("users"))

// Set a user
user := &User{ID: "1", Name: "Alice"}
keyUser.HSet("1", user)

// Get the user
retrievedUser, err := keyUser.HGet("1")
if err == nil {
    fmt.Println("Retrieved User:", retrievedUser.Name)
}
```

## ListKey Example

```go
type Item struct {
    ID   string `msgpack:"id" mod:"trim,lowercase"`
    Name string `msgpack:"name" mod:"trim"`
}

keyList := redisdb.NewListKey[string, *Item](redisdb.WithKey("items"))

// Push an item
item := &Item{ID: "1", Name: "Item1"}
keyList.RPush(item)

// Pop an item
poppedItem, err := keyList.RPop()
if err == nil {
    fmt.Println("Popped Item:", poppedItem.Name)
}
```

## SetKey Example

```go
type Tag struct {
    ID   string `msgpack:"id" mod:"trim,lowercase"`
    Name string `msgpack:"name" mod:"trim"`
}

keyTag := redisdb.NewSetKey[string, *Tag](redisdb.WithKey("tags"))

// Add a tag
tag := &Tag{ID: "1", Name: "Technology"}
keyTag.SAdd(tag)

// Check if tag exists
exists, err := keyTag.SIsMember(tag)
if err == nil {
    fmt.Println("Tag exists:", exists)
}
```

## ZSetKey Example

```go
type ScoredItem struct {
    ID   string  `msgpack:"id" mod:"trim,lowercase"`
    Score float64 `msgpack:"score" mod:"force"`
}

keyScoredItem := redisdb.NewZSetKey[string, *ScoredItem](redisdb.WithKey("scored_items"))

// Add item with score
item := &ScoredItem{ID: "1", Score: 3.5}
keyScoredItem.ZAdd(redis.Z{Member: item, Score: item.Score})

// Get rank
rank, err := keyScoredItem.ZRank(item)
if err == nil {
    fmt.Println("Rank:", rank)
}
```

## Modifiers Example

```go
type ExampleStruct struct {
    Name     string    `mod:"trim,lowercase"`
    Age      int       `mod:"default=18"`
    UnixTime int64     `mod:"unixtime=ms,force"`
    Counter  int64     `mod:"counter,force"`
    Email    string    `mod:"lowercase,trim"`
}

example := ExampleStruct{
    Name:     "  John Doe  ",
    Age:      0,
    UnixTime: 0,
    Counter:  0,
    Email:    "  john.doe@domain.com  ",
}

if err := redisdb.ApplyModifiers(&example); err == nil {
    fmt.Println(example)
}
```

## Error Handling

Always check for errors when using RedisDB methods:

```go
if err := keyConfig.Set("config1", config, time.Hour*24); err != nil {
    fmt.Println("Error setting config:", err)
}
```
