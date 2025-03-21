
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

## Basic Operations

### Setting Hash Fields

To set fields in the hash, use the `HSet` method. This can take either a map of keys to values or a list of key-value pairs.

```go
func (ctx *HashKey[k, v]) HSet(values ...interface{}) error
func (ctx *HashKey[k, v]) HMSet(kvMap map[k]v) error
```

### Getting Hash Fields

Retrieve a specific field from the hash using `HGet`.

```go
func (ctx *HashKey[k, v]) HGet(field k) (value v, err error)
```

### Deleting Hash Fields

Delete a specific field from the hash using `HDel`.

```go
func (ctx *HashKey[k, v]) HDel(fields ...k) error
```

### Checking Field Existence

Check if a field exists in the hash using `HExists`.

```go
func (ctx *HashKey[k, v]) HExists(field k) (bool, error)
```

### Getting All Fields and Values

Get all fields and their corresponding values from the hash.

```go
func (ctx *HashKey[k, v]) HGetAll() (map[k]v, error)
```

### Getting All Hash Fields

Retrieve all fields (keys) from the hash.

```go
func (ctx *HashKey[k, v]) HKeys() ([]k, error)
```

### Getting All Hash Values

Retrieve all values from the hash.

```go
func (ctx *HashKey[k, v]) HVals() ([]v, error)
```

### Getting the Number of Fields

Get the number of fields in the hash.

```go
func (ctx *HashKey[k, v]) HLen() (int64, error)
```

## Modifiers and Data Validation

For information on how to use modifiers with `HashKey`, please refer to the [Modifier Example](doc_mod_example.md).

## Example Usage

### Defining a Struct with Modifiers

```go

type User struct {
    ID   string `msgpack:"id" mod:"trim,lowercase"`
    Name string `msgpack:"name" mod:"trim"`
}
```

### Creating a HashKey Context

```go

keyUser := redisdb.NewHashKey[string, *User](redisdb.WithKey("users"))
```

### Setting and Getting a User

```go
user := &User{ID: "1", Name: "Alice"}
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
