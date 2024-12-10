## Introduction

`SetKey` is a type provided by the `redisdb` package to interact with Redis set keys in a type-safe manner. Redis sets are collections of unique strings, ensuring no duplicates are stored. This makes them ideal for scenarios where uniqueness is crucial, such as storing tags, categories, or any distinct items.

In the `redisdb` package, `SetKey` is defined as a generic type that allows specifying the value type, making it easier to work with typed data without manual marshaling and unmarshaling. It leverages the `msgpack` library for efficient serialization and deserialization of data.

## Creating a SetKey Context

To utilize `SetKey`, you need to create a context that represents your set key in Redis. This is done through the `NewSetKey` function, which accepts optional setters to configure the context.

```go
func NewSetKey[k comparable, v any](ops ...opSetter) *SetKey[k, v]
```

### Optional Parameters

- `WithKey(key string)`: Allows specifying a custom key name for the set. If not provided, the key name is inferred from the type of `v`.

- `WithRds(dataSource string)`: Specifies the Redis data source to use. By default, it uses "default".

- `WithModifier(extraModifiers map[string]ModifierFunc)`: Registers additional modifiers for the members of `v`.

## Basic Operations

### Adding Members to the Set

To add a member to the set, use the `SAdd` method.

```go
func (ctx *SetKey[k, v]) SAdd(param v) error
```

### Getting the Number of Members

Retrieve the number of members in the set using the `SCard` method.

```go
func (ctx *SetKey[k, v]) SCard() (int64, error)
```

### Removing Members from the Set

Remove a specific member from the set using the `SRem` method.

```go
func (ctx *SetKey[k, v]) SRem(param v) error
```

### Checking Member Existence

Check if a member exists in the set using the `SIsMember` method.

```go
func (ctx *SetKey[k, v]) SIsMember(param v) (bool, error)
```

### Getting All Members

Retrieve all members from the set using the `SMembers` method.

```go
func (ctx *SetKey[k, v]) SMembers() ([]v, error)
```

## Advanced Operations

### Incrementally Iterating Through the Set

Use the `SScan` method to incrementally iterate through the set, which is useful for large datasets.

```go
func (ctx *SetKey[k, v]) SScan(cursor uint64, match string, count int64) ([]v, uint64, error)
```

## Modifiers and Data Validation

For information on how to use modifiers with `SetKey`, please refer to the [Modifier Example](doc_mod_example.md).

## Example Usage

### Defining a Struct with Modifiers

```go

type Tag struct {
    ID   string `msgpack:"id" mod:"trim,lowercase"`
    Name string `msgpack:"name" mod:"trim"`
}
```

### Creating a SetKey Context

```go

keyTag := redisdb.NewSetKey[string, *Tag](redisdb.WithKey("tags"))
```

### Adding and Removing Tags

```go
tag := &Tag{ID: "1", Name: "Technology"}
keyTag.SAdd(tag)

tagToRemove := &Tag{ID: "1", Name: "Technology"}
err := keyTag.SRem(tagToRemove)
if err == nil {
    fmt.Println("Tag removed successfully.")
}
```

### Checking Tag Existence

```go
tag := &Tag{ID: "1", Name: "Technology"}
exists, err := keyTag.SIsMember(tag)
if err == nil {
    fmt.Println("Exists:", exists)
}
```

### Getting All Tags

```go
tags, err := keyTag.SMembers()
if err == nil {
    for _, tag := range tags {
        fmt.Println("Tag ID:", tag.ID, "Name:", tag.Name)
    }
}
```

## Error Handling

Most methods in `SetKey` return errors that you should check to handle failures gracefully. For example:

```go
if err := keyTag.SAdd(tag); err != nil {
    fmt.Println("Error adding tag:", err)
}
```

Always ensure to handle errors to maintain the robustness of your application.

## Conclusion

The `SetKey` type in the `redisdb` package provides a convenient and type-safe way to work with Redis set keys in Go. By using generics, it allows you to specify the types of values, reducing the chance of errors and making your code more maintainable. Explore the other methods and features provided by the package to fully leverage its capabilities.
