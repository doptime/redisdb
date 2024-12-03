## ListKey Documentation

`ListKey` is a type provided by the `redisdb` package to interact with Redis list keys in a type-safe manner. Redis lists are collections of strings that are ordered in a FIFO (First In, First Out) or LIFO (Last In, First Out) manner, depending on the operation used.

In the `redisdb` package, `ListKey` is defined as a generic type that allows specifying the key and value types, making it easier to work with typed data without manual marshaling and unmarshaling.

## Creating a ListKey Context

To use `ListKey`, you first need to create a context that represents your list key in Redis. This is done using the `NewListKey` function, which takes optional setters to configure the context.

```go
func NewListKey[k comparable, v any](ops ...opSetter) *ListKey[k, v]
```

### Optional Parameters

- `WithKey(key string)`: Allows specifying a custom key name for the list. If not provided, the key name is inferred from the type of `v`.

- `WithRds(dataSource string)`: Specifies the Redis data source to use. By default, it uses "default".

- `WithModifier(extraModifiers map[string]ModifierFunc)`: Registers additional modifiers for the fields of `v`.

## Basic Operations

### Pushing Elements

To add elements to the list, use `RPush` to append elements to the end or `LPush` to push elements to the beginning.

```go
func (ctx *ListKey[k, v]) RPush(param ...v) error
func (ctx *ListKey[k, v]) LPush(param ...v) error
```

#### Example

```go
item := &Item{ID: "1", Name: "Item1"}
keyList.RPush(item)
keyList.LPush(item)
```

### Popping Elements

Remove and retrieve elements from the list using `RPop` or `LPop`.

```go
func (ctx *ListKey[k, v]) RPop() (ret v, err error)
func (ctx *ListKey[k, v]) LPop() (ret v, err error)
```

#### Example

```go
poppedItem, err := keyList.RPop()
if err == nil {
    fmt.Println("Popped Item:", poppedItem.Name)
}
```

## Advanced Operations

### Retrieving Elements by Index

Retrieve elements from the list between specified indices using `LRange`.

```go
func (ctx *ListKey[k, v]) LRange(start, stop int64) ([]v, error)
```

#### Example

```go
items, err := keyList.LRange(0, -1)
if err == nil {
    for _, item := range items {
        fmt.Println("Item:", item.Name)
    }
}
```

### Removing Elements

Remove specific elements from the list using `LRem`.

```go
func (ctx *ListKey[k, v]) LRem(count int64, param v) error
```

#### Example

```go
err := keyList.LRem(1, &Item{ID: "1", Name: "Item1"})
if err == nil {
    fmt.Println("Element removed successfully.")
}
```

### Trimming the List

Trim the list to retain only elements between specified indices using `LTrim`.

```go
func (ctx *ListKey[k, v]) LTrim(start, stop int64) error
```

#### Example

```go
err := keyList.LTrim(0, 9)
if err == nil {
    fmt.Println("List trimmed to 10 elements.")
}
```

## Modifiers and Data Validation

The `redisdb` package supports modifiers that can be applied to the fields of your struct during serialization and deserialization. This is particularly useful for tasks like trimming spaces, converting to lowercase, setting default values, etc.

Modifiers are defined using struct tags. For example:

```go
type Item struct {
    ID   string `msgpack:"id" mod:"trim,lowercase"`
    Name string `msgpack:"name" mod:"trim"`
}
```

In this example, the `ID` field will be trimmed and converted to lowercase before being stored in Redis, and the `Name` field will be trimmed.

### Registering Custom Modifiers

You can register custom modifiers by providing a map of modifier names to modifier functions when creating the `ListKey` context.

```go
extraModifiers := map[string]ModifierFunc{
    "custom": func(fieldValue interface{}, tagParam string) (interface{}, error) {
        // Custom modification logic
    },
}
keyList := redisdb.NewListKey[string, *Item](redisdb.WithModifier(extraModifiers))
```

## Example Usage

### Defining a Struct with Modifiers

```go
type Item struct {
    ID   string `msgpack:"id" mod:"trim,lowercase"`
    Name string `msgpack:"name" mod:"trim"`
}
```

### Creating a ListKey Context

```go
keyList := redisdb.NewListKey[string, *Item](redisdb.WithKey("items"))
```

### Pushing and Popping Items

```go
item := &Item{ID: "1", Name: "Item1"}
keyList.RPush(item)

poppedItem, err := keyList.RPop()
if err == nil {
    fmt.Println("Popped Item:", poppedItem.Name)
}
```

### Retrieving All Items

```go
items, err := keyList.LRange(0, -1)
if err == nil {
    for _, item := range items {
        fmt.Println("Item:", item.Name)
    }
}
```

## Error Handling

Most methods in `ListKey` return errors that you should check to handle failures gracefully. For example:

```go
if err := keyList.RPush(item); err != nil {
    fmt.Println("Error pushing item:", err)
}
```

Always ensure to handle errors to maintain the robustness of your application.

## Conclusion

The `ListKey` type in the `redisdb` package provides a convenient and type-safe way to work with Redis list keys in Go. By using generics, it allows you to specify the types of keys and values, reducing the chance of errors and making your code more maintainable. Explore the other methods and features provided by the package to fully leverage its capabilities.
