---
slug: Data
title: HashKey
type:  docs
sidebar_position: 1
---


## HashKey Documentation

`HashKey` is a type provided by the `redisdb` package to interact with Redis hash keys in a type-safe manner. Redis hash keys are maps between string fields and string values, which can be used to represent complex objects or dictionaries.

In the `redisdb` package, `HashKey` is defined as a generic type that allows specifying the key and value types, making it easier to work with typed data without manual marshaling and unmarshaling.

## Creating a HashKey Context

To use `HashKey`, you first need to create a context that represents your hash key in Redis. This is done using the `NewHashKey` function, which takes optional setters to configure the context.

```go
func NewHashKey[k comparable, v any](ops ...opSetter) *HashKey[k, v]
```

### Optional Parameters

- `WithKey(key string)`: Allows specifying a custom key name for the hash. If not provided, the key name is inferred from the type of `v`.

- `WithRds(dataSource string)`: Specifies the Redis data source to use. By default, it uses "default".

- `WithModifier(extraModifiers map[string]ModifierFunc)`: Registers additional modifiers for the fields of `v`.

### Example

```go
type User struct {
    ID   string `msgpack:"id"`
    Name string `msgpack:"name"`
}

keyUser := redisdb.NewHashKey[string, *User](redisdb.WithKey("users"))
```

## Basic Operations

### Setting Hash Fields

To set fields in the hash, use the `HSet` method. This can take either a map of keys to values or a list of key-value pairs.

```go
func (ctx *HashKey[k, v]) HSet(values ...interface{}) error
func (ctx *HashKey[k, v]) HMSet(kvMap map[k]v) error
```

#### Example

```go
user := &User{ID: "1", Name: "Alice"}
keyUser.HSet("1", user)
```

### Getting Hash Fields

Retrieve a specific field from the hash using `HGet`.

```go
func (ctx *HashKey[k, v]) HGet(field k) (value v, err error)
```

#### Example

```go
user, err := keyUser.HGet("1")
if err == nil {
    fmt.Println(user.Name)
}
```

### Deleting Hash Fields

Delete a specific field from the hash using `HDel`.

```go
func (ctx *HashKey[k, v]) HDel(fields ...k) error
```

#### Example

```go
err := keyUser.HDel("1")
if err == nil {
    fmt.Println("Field deleted successfully.")
}
```

### Checking Field Existence

Check if a field exists in the hash using `HExists`.

```go
func (ctx *HashKey[k, v]) HExists(field k) (bool, error)
```

#### Example

```go
exists, err := keyUser.HExists("1")
if err == nil {
    fmt.Println("Exists:", exists)
}
```

## Advanced Operations

### Retrieving All Fields and Values

Get all fields and their corresponding values from the hash.

```go
func (ctx *HashKey[k, v]) HGetAll() (map[k]v, error)
```

#### Example

```go
users, err := keyUser.HGetAll()
if err == nil {
    for id, user := range users {
        fmt.Println("ID:", id, "Name:", user.Name)
    }
}
```

### Getting All Hash Fields

Retrieve all fields (keys) from the hash.

```go
func (ctx *HashKey[k, v]) HKeys() ([]k, error)
```

#### Example

```go
keys, err := keyUser.HKeys()
if err == nil {
    for _, key := range keys {
        fmt.Println(key)
    }
}
```

### Getting All Hash Values

Retrieve all values from the hash.

```go
func (ctx *HashKey[k, v]) HVals() ([]v, error)
```

#### Example

```go
values, err := keyUser.HVals()
if err == nil {
    for _, user := range values {
        fmt.Println(user.Name)
    }
}
```

### Getting the Number of Fields

Get the number of fields in the hash.

```go
func (ctx *HashKey[k, v]) HLen() (int64, error)
```

#### Example

```go
length, err := keyUser.HLen()
if err == nil {
    fmt.Println("Number of fields:", length)
}
```

## Modifiers and Data Validation

The `redisdb` package supports modifiers that can be applied to the fields of your struct during serialization and deserialization. This is particularly useful for tasks like trimming spaces, converting to lowercase, setting default values, etc.

Modifiers are defined using struct tags. For example:

```go
type User struct {
    ID       string    `msgpack:"id" mod:"trim,lowercase"`
    JoinTime int64     `msgpack:"join_time" mod:"unixtime=ms,force"`
}
```

In this example, the `ID` field will be trimmed and converted to lowercase before being stored in Redis, and the `JoinTime` field will be stored as a Unix timestamp in milliseconds.

### Registering Custom Modifiers

You can register custom modifiers by providing a map of modifier names to modifier functions when creating the `HashKey` context.

```go
extraModifiers := map[string]ModifierFunc{
    "custom": func(fieldValue interface{}, tagParam string) (interface{}, error) {
        // Custom modification logic
    },
}
keyUser := redisdb.NewHashKey[string, *User](redisdb.WithModifier(extraModifiers))
```

## Example Usage

### Defining a Struct with Modifiers

```go
type User struct {
    ID       string    `msgpack:"id" mod:"trim,lowercase"`
    Name     string    `msgpack:"name" mod:"trim"`
    JoinTime int64     `msgpack:"join_time" mod:"unixtime=ms,force"`
}
```

### Creating a HashKey Context

```go
keyUser := redisdb.NewHashKey[string, *User](redisdb.WithKey("users"))
```

### Setting and Getting a User

```go
user := &User{ID: "1", Name: "Alice", JoinTime: time.Now().UnixMilli()}
keyUser.HSet("1", user)

retrievedUser, err := keyUser.HGet("1")
if err == nil {
    fmt.Println("Retrieved User:", retrievedUser.Name)
}
```

### Getting All Users

```go
users, err := keyUser.HGetAll()
if err == nil {
    for id, user := range users {
        fmt.Println("ID:", id, "Name:", user.Name)
    }
}
```

## Error Handling

Most methods in `HashKey` return errors that you should check to handle failures gracefully. For example:

```go
if err := keyUser.HSet("1", user); err != nil {
    fmt.Println("Error setting user:", err)
}
```

Always ensure to handle errors to maintain the robustness of your application.

## Conclusion

The `HashKey` type in the `redisdb` package provides a convenient and type-safe way to work with Redis hash keys in Go. By using generics, it allows you to specify the types of keys and values, reducing the chance of errors and making your code more maintainable. Explore the other methods and features provided by the package to fully leverage its capabilities.
