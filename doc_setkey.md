## Introduction

`SetKey` is a type provided by the `redisdb` package to interact with Redis set keys in a type-safe manner. Redis sets are collections of unique strings, ensuring no duplicates are stored. This makes them ideal for scenarios where uniqueness is crucial, such as storing tags, categories, or any distinct items.

In the `redisdb` package, `SetKey` is defined as a generic type that allows specifying the value type, making it easier to work with typed data without manual marshaling and unmarshaling. It leverages the `msgpack` library for efficient serialization and deserialization of data.

## Creating a SetKey Context

To utilize `SetKey`, you need to create a context that represents your set key in Redis. This is achieved through the `NewSetKey` function, which accepts optional setters to configure the context.

```go
func NewSetKey[k comparable, v any](ops ...opSetter) *SetKey[k, v]
```

### Optional Parameters

- `WithKey(key string)`: Allows specifying a custom key name for the set. If not provided, the key name is inferred from the type of `v`.

- `WithRds(dataSource string)`: Specifies the Redis data source to use. By default, it uses "default".

- `WithModifier(extraModifiers map[string]ModifierFunc)`: Registers additional modifiers for the members of `v`.

### Example

```go
type Tag struct {
    ID   string `msgpack:"id" mod:"trim,lowercase"`
    Name string `msgpack:"name" mod:"trim"`
}

keyTag := redisdb.NewSetKey[string, *Tag](redisdb.WithKey("tags"))
```

## Basic Operations

### Adding Members to the Set

To add a member to the set, use the `SAdd` method.

```go
func (ctx *SetKey[k, v]) SAdd(param v) error
```

#### Example

```go
tag := &Tag{ID: "1", Name: "Technology"}
keyTag.SAdd(tag)
```

### Getting the Number of Members

Retrieve the number of members in the set using the `SCard` method.

```go
func (ctx *SetKey[k, v]) SCard() (int64, error)
```

#### Example

```go
count, err := keyTag.SCard()
if err == nil {
    fmt.Println("Number of tags:", count)
}
```

### Removing Members from the Set

Remove a specific member from the set using the `SRem` method.

```go
func (ctx *SetKey[k, v]) SRem(param v) error
```

#### Example

```go
tagToRemove := &Tag{ID: "1", Name: "Technology"}
err := keyTag.SRem(tagToRemove)
if err == nil {
    fmt.Println("Tag removed successfully.")
}
```

### Checking Member Existence

Check if a member exists in the set using the `SIsMember` method.

```go
func (ctx *SetKey[k, v]) SIsMember(param v) (bool, error)
```

#### Example

```go
tag := &Tag{ID: "1", Name: "Technology"}
exists, err := keyTag.SIsMember(tag)
if err == nil {
    fmt.Println("Exists:", exists)
}
```

### Getting All Members

Retrieve all members from the set using the `SMembers` method.

```go
func (ctx *SetKey[k, v]) SMembers() ([]v, error)
```

#### Example

```go
tags, err := keyTag.SMembers()
if err == nil {
    for _, tag := range tags {
        fmt.Println("Tag ID:", tag.ID, "Name:", tag.Name)
    }
}
```

## Advanced Operations

### Incrementally Iterating Through the Set

Use the `SScan` method to incrementally iterate through the set, which is useful for large datasets.

```go
func (ctx *SetKey[k, v]) SScan(cursor uint64, match string, count int64) ([]v, uint64, error)
```

#### Example

```go
var cursor uint64 = 0
var tags []v
for {
    _tags, newCursor, err := keyTag.SScan(cursor, "*", 10)
    if err != nil {
        break
    }
    tags = append(tags, _tags...)
    if newCursor == 0 {
        break
    }
    cursor = newCursor
}
```

## Modifiers and Data Validation

The `redisdb` package supports modifiers that can be applied to the members of your struct during serialization and deserialization. This is particularly useful for tasks like trimming spaces, converting to lowercase, setting default values, etc.

Modifiers are defined using struct tags. For example:

```go
type Tag struct {
    ID   string `msgpack:"id" mod:"trim,lowercase"`
    Name string `msgpack:"name" mod:"trim"`
}
```

In this example, the `ID` field will be trimmed and converted to lowercase before being stored in Redis, and the `Name` field will be trimmed.

### Registering Custom Modifiers

You can register custom modifiers by providing a map of modifier names to modifier functions when creating the `SetKey` context.

```go
extraModifiers := map[string]ModifierFunc{
    "custom": func(fieldValue interface{}, tagParam string) (interface{}, error) {
        // Custom modification logic
    },
}
keyTag := redisdb.NewSetKey[string, *Tag](redisdb.WithModifier(extraModifiers))
```

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