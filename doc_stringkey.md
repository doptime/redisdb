---
slug: data-stringkey
title: StringKey
type:  docs
sidebar_position: 2
---

## StringKey Documentation

`StringKey` is a type provided by the `redisdb` package to interact with Redis string keys in a type-safe manner. Redis strings are the simplest data type in Redis, used to store single string values. They can be used for a variety of purposes, such as caching, storing configuration settings, or any scenario where a single piece of data needs to be stored and retrieved efficiently.

In the `redisdb` package, `StringKey` is defined as a generic type that allows specifying the key and value types, making it easier to work with typed data without manual marshaling and unmarshaling.

## Creating a StringKey Context

To use `StringKey`, you first need to create a context that represents your string key in Redis. This is done using the `NewStringKey` function, which takes optional setters to configure the context.

```go
func NewStringKey[k comparable, v any](ops ...opSetter) *StringKey[k, v]
```

### Optional Parameters

- `WithKey(key string)`: Allows specifying a custom key name for the string. If not provided, the key name is inferred from the type of `v`.

- `WithRds(dataSource string)`: Specifies the Redis data source to use. By default, it uses "default".

- `WithModifier(extraModifiers map[string]ModifierFunc)`: Registers additional modifiers for the fields of `v`.

### Example

```go
type Config struct {
    Host     string `msgpack:"host" mod:"trim,lowercase"`
    Port     int    `msgpack:"port" mod:"default=8080"`
    Enabled  bool   `msgpack:"enabled" mod:"default=true"`
}

keyConfig := redisdb.NewStringKey[string, *Config](redisdb.WithKey("configs"))
```

## Basic Operations

### Setting a String Value

To set a string value in Redis, use the `Set` method. This method requires a key and a value, along with an optional expiration time.

```go
func (ctx *StringKey[k, v]) Set(key k, value v, expiration time.Duration) error
```

### Getting a String Value

Retrieve a string value from Redis using the `Get` method.

```go
func (ctx *StringKey[k, v]) Get(field k) (value v, err error)
```

### Deleting a String Value

Delete a specific string value from Redis using the `Del` method.

```go
func (ctx *StringKey[k, v]) Del(key k) error
```

### Example Usage

#### Setting and Getting a Configuration

```go
config := &Config{Host: "example.com", Port: 8080, Enabled: true}
keyConfig.Set("config1", config, time.Hour*24)

retrievedConfig, err := keyConfig.Get("config1")
if err == nil {
    fmt.Println("Retrieved Config:", retrievedConfig.Host, retrievedConfig.Port, retrievedConfig.Enabled)
}
```

## Advanced Operations

### Retrieving All Keys and Values

Get all keys that match a specific pattern and retrieve their corresponding values.

```go
func (ctx *StringKey[k, v]) GetAll(match string) (mapOut map[k]v, err error)
```

### Setting Multiple Keys at Once

Set multiple key-value pairs to Redis strings in a single operation.

```go
func (ctx *StringKey[k, v]) SetAll(_map map[k]v) error
```

### Example Usage

#### Getting All Configurations

```go
configs, err := keyConfig.GetAll("configs:*")
if err == nil {
    for key, config := range configs {
        fmt.Println("Key:", key, "Config:", config.Host, config.Port, config.Enabled)
    }
}
```

## Modifiers and Data Validation

The `redisdb` package supports modifiers that can be applied to the fields of your struct during serialization and deserialization. This is particularly useful for tasks like trimming spaces, converting to lowercase, setting default values, etc.

Modifiers are defined using struct tags. For example:

```go
type Config struct {
    Host     string `msgpack:"host" mod:"trim,lowercase"`
    Port     int    `msgpack:"port" mod:"default=8080"`
    Enabled  bool   `msgpack:"enabled" mod:"default=true"`
}
```

In this example, the `Host` field will be trimmed and converted to lowercase before being stored in Redis, and the `Port` and `Enabled` fields will have default values applied if they are empty.

### Registering Custom Modifiers

You can register custom modifiers by providing a map of modifier names to modifier functions when creating the `StringKey` context.

```go
extraModifiers := map[string]ModifierFunc{
    "custom": func(fieldValue interface{}, tagParam string) (interface{}, error) {
        // Custom modification logic
    },
}
keyConfig := redisdb.NewStringKey[string, *Config](redisdb.WithModifier(extraModifiers))
```

## Example Usage

### Defining a Struct with Modifiers

```go
type Config struct {
    Host     string `msgpack:"host" mod:"trim,lowercase"`
    Port     int    `msgpack:"port" mod:"default=8080"`
    Enabled  bool   `msgpack:"enabled" mod:"default=true"`
}
```

### Creating a StringKey Context

```go
keyConfig := redisdb.NewStringKey[string, *Config](redisdb.WithKey("configs"))
```

### Setting and Getting Configurations

```go
config := &Config{Host: "example.com", Port: 8080, Enabled: true}
keyConfig.Set("config1", config, time.Hour*24)

retrievedConfig, err := keyConfig.Get("config1")
if err == nil {
    fmt.Println("Retrieved Config:", retrievedConfig.Host, retrievedConfig.Port, retrievedConfig.Enabled)
}
```

### Getting All Configurations

```go
configs, err := keyConfig.GetAll("configs:*")
if err == nil {
    for key, config := range configs {
        fmt.Println("Key:", key, "Config:", config.Host, config.Port, config.Enabled)
    }
}
```

## Error Handling

Most methods in `StringKey` return errors that you should check to handle failures gracefully. For example:

```go
if err := keyConfig.Set("config1", config, time.Hour*24); err != nil {
    fmt.Println("Error setting config:", err)
}
```

Always ensure to handle errors to maintain the robustness of your application.

## Conclusion

The `StringKey` type in the `redisdb` package provides a convenient and type-safe way to work with Redis string keys in Go. By using generics, it allows you to specify the types of keys and values, reducing the chance of errors and making your code more maintainable. Explore the other methods and features provided by the package to fully leverage its capabilities.