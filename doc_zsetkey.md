## Introduction

`ZSetKey` is a type provided by the `redisdb` package to interact with Redis sorted sets (zsets) in a type-safe manner. Redis sorted sets are collections of unique strings, each associated with a floating-point value that determines its sorting order. They are ideal for scenarios where you need to maintain a sorted list of items.

In the `redisdb` package, `ZSetKey` is defined as a generic type that allows specifying the value type, making it easier to work with typed data without manual marshaling and unmarshaling. It uses the `msgpack` library for efficient serialization and deserialization of data.

## Creating a ZSetKey Context

To use `ZSetKey`, you need to create a context that represents your sorted set key in Redis. This is done using the `NewZSetKey` function, which accepts optional setters to configure the context.

```go
func NewZSetKey[k comparable, v any](ops ...opSetter) *ZSetKey[k, v]
```

### Optional Parameters

- `WithKey(key string)`: Allows specifying a custom key name for the sorted set. If not provided, the key name is inferred from the type of `v`.

- `WithRds(dataSource string)`: Specifies the Redis data source to use. By default, it uses "default".

- `WithModifier(extraModifiers map[string]ModifierFunc)`: Registers additional modifiers for the fields of `v`.

### Example

```go
type ScoredItem struct {
    ID   string  `msgpack:"id" mod:"trim,lowercase"`
    Score float64 `msgpack:"score" mod:"force"`
}

keyScoredItem := redisdb.NewZSetKey[string, *ScoredItem](redisdb.WithKey("scored_items"))
```

## Basic Operations

### Adding Members with Scores

To add members to the sorted set, use the `ZAdd` method. This method requires `redis.Z` structures, which contain the member and its score.

```go
func (ctx *ZSetKey[k, v]) ZAdd(members ...redis.Z) error
```

#### Example

```go
item := &ScoredItem{ID: "1", Score: 3.5}
keyScoredItem.ZAdd(redis.Z{Member: item, Score: item.Score})
```

### Removing Members

Remove specific members from the sorted set using the `ZRem` method.

```go
func (ctx *ZSetKey[k, v]) ZRem(members ...interface{}) error
```

#### Example

```go
itemToRemove := &ScoredItem{ID: "1", Score: 3.5}
err := keyScoredItem.ZRem(itemToRemove)
if err == nil {
    fmt.Println("Item removed successfully.")
}
```

### Getting the Number of Members

Retrieve the number of members in the sorted set using the `ZCard` method.

```go
func (ctx *ZSetKey[k, v]) ZCard() (int64, error)
```

#### Example

```go
count, err := keyScoredItem.ZCard()
if err == nil {
    fmt.Println("Number of items:", count)
}
```

### Getting Members by Score Range

Retrieve members within a specific score range using the `ZRangeByScore` method.

```go
func (ctx *ZSetKey[k, v]) ZRangeByScore(opt *redis.ZRangeBy) ([]v, error)
```

#### Example

```go
opt := &redis.ZRangeBy{Min: "0", Max: "5"}
items, err := keyScoredItem.ZRangeByScore(opt)
if err == nil {
    for _, item := range items {
        fmt.Println("Item ID:", item.ID, "Score:", item.Score)
    }
}
```

## Advanced Operations

### Incrementing Member Scores

Increment the score of a member by a given amount using the `ZIncrBy` method.

```go
func (ctx *ZSetKey[k, v]) ZIncrBy(increment float64, member v) error
```

#### Example

```go
item := &ScoredItem{ID: "1", Score: 3.5}
err := keyScoredItem.ZIncrBy(1.0, item)
if err == nil {
    fmt.Println("Score incremented successfully.")
}
```

### Getting Members with Scores

Retrieve members along with their scores using the `ZRangeWithScores` or `ZRevRangeWithScores` methods.

```go
func (ctx *ZSetKey[k, v]) ZRangeWithScores(start, stop int64) ([]v, []float64, error)
func (ctx *ZSetKey[k, v]) ZRevRangeWithScores(start, stop int64) ([]v, []float64, error)
```

#### Example

```go
items, scores, err := keyScoredItem.ZRangeWithScores(0, -1)
if err == nil {
    for i, item := range items {
        fmt.Println("Item ID:", item.ID, "Score:", scores[i])
    }
}
```

### Getting Member Ranks

Get the rank of a member in the sorted set using the `ZRank` or `ZRevRank` methods.

```go
func (ctx *ZSetKey[k, v]) ZRank(member interface{}) (int64, error)
func (ctx *ZSetKey[k, v]) ZRevRank(member interface{}) (int64, error)
```

#### Example

```go
item := &ScoredItem{ID: "1", Score: 3.5}
rank, err := keyScoredItem.ZRank(item)
if err == nil {
    fmt.Println("Rank:", rank)
}
```

### Deleting Members by Rank Range

Remove members from the sorted set based on their rank range.

```go
func (ctx *ZSetKey[k, v]) ZRemRangeByRank(start, stop int64) error
```

#### Example

```go
err := keyScoredItem.ZRemRangeByRank(0, 10)
if err == nil {
    fmt.Println("Members removed successfully.")
}
```

## Modifiers and Data Validation

The `redisdb` package supports modifiers that can be applied to the fields of your struct during serialization and deserialization. This is particularly useful for tasks like trimming spaces, converting to lowercase, setting default values, etc.

Modifiers are defined using struct tags. For example:

```go
type ScoredItem struct {
    ID   string  `msgpack:"id" mod:"trim,lowercase"`
    Score float64 `msgpack:"score" mod:"force"`
}
```

In this example, the `ID` field will be trimmed and converted to lowercase before being stored in Redis, and the `Score` field will have its modifiers applied.

### Registering Custom Modifiers

You can register custom modifiers by providing a map of modifier names to modifier functions when creating the `ZSetKey` context.

```go
extraModifiers := map[string]ModifierFunc{
    "custom": func(fieldValue interface{}, tagParam string) (interface{}, error) {
        // Custom modification logic
    },
}
keyScoredItem := redisdb.NewZSetKey[string, *ScoredItem](redisdb.WithModifier(extraModifiers))
```

## Example Usage

### Defining a Struct with Modifiers

```go
type ScoredItem struct {
    ID   string  `msgpack:"id" mod:"trim,lowercase"`
    Score float64 `msgpack:"score" mod:"force"`
}
```

### Creating a ZSetKey Context

```go
keyScoredItem := redisdb.NewZSetKey[string, *ScoredItem](redisdb.WithKey("scored_items"))
```

### Adding and Removing Items

```go
item := &ScoredItem{ID: "1", Score: 3.5}
keyScoredItem.ZAdd(redis.Z{Member: item, Score: item.Score})

itemToRemove := &ScoredItem{ID: "1", Score: 3.5}
err := keyScoredItem.ZRem(itemToRemove)
if err == nil {
    fmt.Println("Item removed successfully.")
}
```

### Getting Item Ranks

```go
item := &ScoredItem{ID: "1", Score: 3.5}
rank, err := keyScoredItem.ZRank(item)
if err == nil {
    fmt.Println("Rank:", rank)
}
```

### Retrieving All Items

```go
items, err := keyScoredItem.ZRange(0, -1)
if err == nil {
    for _, item := range items {
        fmt.Println("Item ID:", item.ID, "Score:", item.Score)
    }
}
```

## Error Handling

Most methods in `ZSetKey` return errors that you should check to handle failures gracefully. For example:

```go
if err := keyScoredItem.ZAdd(redis.Z{Member: item, Score: item.Score}); err != nil {
    fmt.Println("Error adding item:", err)
}
```

Always ensure to handle errors to maintain the robustness of your application.

## Conclusion

The `ZSetKey` type in the `redisdb` package provides a convenient and type-safe way to work with Redis sorted sets in Go. By using generics, it allows you to specify the types of values, reducing the chance of errors and making your code more maintainable. Explore the other methods and features provided by the package to fully leverage its capabilities.
