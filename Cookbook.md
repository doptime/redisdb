# RedisDB Cookbook: Comprehensive Usage Guide

This cookbook provides concise, step-by-step examples for using RedisDB with all data structures in a LLM-friendly format.

## Introduction
RedisDB is a type-safe Go package for interacting with Redis with automatic serialization and built-in modifiers.

## Basic Usage
- Install doptime/redisdb with go get
- Configure Redis connection in config.yml
- Use the package to interact with Redis keys

## 配置说明

在使用RedisDB之前，需要配置Redis连接。Doptime框架支持从配置文件加载Redis配置：

```yaml
db: 
  redis:
    default:
      network: "tcp"
      address: "localhost:6379"
      password: ""
      database: 0
      pool_size: 10
      read_timeout: "5s"
      write_timeout: "5s"
      pool_timeout: "10s"
```

## 项目初始化模板

创建基于RedisDB的新项目的典型目录结构和代码结构示例：

```bash
project/
  |- config/
  |  |- config.yml
  |- redisdb/
  |  |- cookbook.md  # 本参考文档
  |  |- config.go  # 配置常量和其他公共结构体
  |  |- user.go    # 用户相关数据操作
  |  |- main.go
```

```go
// config.go
package redisdb

// 配置常量
const (
    UserKeyPrefix = "users:"
    ConfigKey = "configs"
)
```

```go
// main.go
package main

import (
    "github.com/doptime/redisdb"
    "log"
)

func main() {
    // 初始化Redis配置
    cfgredis.Servers.Load("config.yml")
    
    // 创建并使用示例
    userKey := redisdb.NewHashKey[string, *User](redisdb.WithKey(redisdb.UserKeyPrefix))
    
    // ... 其他功能代码 ...
    log.Println("RedisDB项目初始化成功")
}
```

## Data Structures

### StringKey
Store a single key-value pair with type safety.
```go
type Config struct {
    Host     string `msgpack:"host" mod:"trim,lowercase"
    Port     int    `msgpack:"port" mod:"default=8080"
    Enabled  bool   `msgpack:"enabled" mod:"default=true"
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

### HashKey
Store a hash with key-value pairs for a composite object.
```go
type User struct {
    ID   string `msgpack:"id" mod:"trim,lowercase"
    Name string `msgpack:"name" mod:"trim"
}

keyUser := redisdb.NewHashKey[string, *User](redisdb.WithKey("users"))

// Add a user
user := &User{ID: "1", Name: "Alice"}
keyUser.HSet("1", user)

// Get a user
retrievedUser, err := keyUser.HGet("1")
if err == nil {
    fmt.Println("Retrieved User:", retrievedUser.Name)
}
```

### ListKey
Store a list of elements with FIFO ordering.
```go
type Item struct {
    ID   string `msgpack:"id" mod:"trim,lowercase"
    Name string `msgpack:"name" mod:"trim"
}

keyList := redisdb.NewListKey[string, *Item](redisdb.WithKey("items"))

// Add an item to the end
item := &Item{ID: "1", Name: "Item1"}
keyList.RPush(item)

// Pop an item from the end
poppedItem, err := keyList.RPop()
if err == nil {
    fmt.Println("Popped Item:", poppedItem.Name)
}
```

### SetKey
Store unique elements in an unordered set.
```go
type Tag struct {
    ID   string `msgpack:"id" mod:"trim,lowercase"
    Name string `msgpack:"name" mod:"trim"
}

keyTag := redisdb.NewSetKey[string, *Tag](redisdb.WithKey("tags"))

// Add a tag to the set
tag := &Tag{ID: "1", Name: "Technology"}
keyTag.SAdd(tag)

// Check if a tag exists
exists, err := keyTag.SIsMember(tag)
if err == nil {
    fmt.Println("Exists:", exists)
}
```

### ZSetKey
Store key-value pairs with scores for ordering.
```go
type ScoredItem struct {
    ID   string  `msgpack:"id" mod:"trim,lowercase"
    Score float64 `msgpack:"score" mod:"force"
}

keyScoredItem := redisdb.NewZSetKey[string, *ScoredItem](redisdb.WithKey("scored_items"))

// Add an item with score
item := &ScoredItem{ID: "1", Score: 3.5}
keyScoredItem.ZAdd(redis.Z{Member: item, Score: item.Score})

// Get the rank of an item
rank, err := keyScoredItem.ZRank(item)
if err == nil {
    fmt.Println("Rank:", rank)
}
```

## Modifiers and Validation
Apply built-in modifiers to fields during serialization.
```go
type ExampleStruct struct {
    Name     string    `mod:"trim,lowercase"
    Age      int       `mod:"default=18"
    UnixTime int64     `mod:"unixtime=ms,force"
    Counter  int64     `mod:"counter,force"
    Email    string    `mod:"lowercase,trim"
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
}{%- if tools %}
    {{- '<|im_start|>system\n' }}
    {%- if messages[0].role == 'system' %}
        {{- messages[0].content + '\n\n' }}
    {%- endif %}
    {{- "# Tools\n\nYou may call one or more functions to assist with the user query.\n\nYou are provided with function signatures within <tools></tools> XML tags:\n<tools>" }}
    {%- for tool in tools %}
        {{- "\n" }}
        {{- tool | tojson }}
    {%- endfor %}
    {{- "\n</tools>\n\nFor each function call, return a json object with function name and arguments within <tool_call></tool_call> XML tags:\n<tool_call>\n{\"name\": <function-name>, \"arguments\": <args-json-object>}\n</tool_call><|im_end|>\n" }}
{%- else %}
    {%- if messages[0].role == 'system' %}
        {{- '<|im_start|>system\n' + messages[0].content + '<|im_end|>\n' }}
    {%- endif %}
{%- endif %}
{%- set ns = namespace(multi_step_tool=true, last_query_index=messages|length - 1) %}
{%- for message in messages[::-1] %}
    {%- set index = (messages|length - 1) - loop.index0 %}
    {%- if ns.multi_step_tool and message.role == "user" and message.content is string and not(message.content.startswith('<tool_response>') and message.content.endswith('</tool_response>')) %}
        {%- set ns.multi_step_tool = false %}
        {%- set ns.last_query_index = index %}
    {%- endif %}
{%- endfor %}
{%- for message in messages %}
    {%- if message.content is string %}
        {%- set content = message.content %}
    {%- else %}
        {%- set content = '' %}
    {%- endif %}
    {%- if (message.role == "user") or (message.role == "system" and not loop.first) %}
        {{- '<|im_start|>' + message.role + '\n' + content + '<|im_end|>' + '\n' }}
    {%- elif message.role == "assistant" %}
        {%- set reasoning_content = '' %}
        {%- if message.reasoning_content is string %}
            {%- set reasoning_content = message.reasoning_content %}
        {%- else %}
            {%- if '</think>' in content %}
                {%- set reasoning_content = content.split('</think>')[0].rstrip('\n').split('<think>')[-1].lstrip('\n') %}
                {%- set content = content.split('</think>')[-1].lstrip('\n') %}
            {%- endif %}
        {%- endif %}
        {%- if loop.index0 > ns.last_query_index %}
            {%- if loop.last or (not loop.last and reasoning_content) %}
                {{- '<|im_start|>' + message.role + '\n<think>\n' + reasoning_content.strip('\n') + '\n</think>\n\n' + content.lstrip('\n') }}
            {%- else %}
                {{- '<|im_start|>' + message.role + '\n' + content }}
            {%- endif %}
        {%- else %}
            {{- '<|im_start|>' + message.role + '\n' + content }}
        {%- endif %}
        {%- if message.tool_calls %}
            {%- for tool_call in message.tool_calls %}
                {%- if (loop.first and content) or (not loop.first) %}
                    {{- '\n' }}
                {%- endif %}
                {%- if tool_call.function %}
                    {%- set tool_call = tool_call.function %}
                {%- endif %}
                {{- '<tool_call>\n{"name": "' }}
                {{- tool_call.name }}
                {{- '", "arguments": ' }}
                {%- if tool_call.arguments is string %}
                    {{- tool_call.arguments }}
                {%- else %}
                    {{- tool_call.arguments | tojson }}
                {%- endif %}
                {{- '}\n</tool_call>' }}
            {%- endfor %}
        {%- endif %}
        {{- '<|im_end|>\n' }}
    {%- elif message.role == "tool" %}
        {%- if loop.first or (messages[loop.index0 - 1].role != "tool") %}
            {{- '<|im_start|>user' }}
        {%- endif %}
        {{- '\n<tool_response>\n' }}
        {{- content }}
        {{- '\n</tool_response>' }}
        {%- if loop.last or (messages[loop.index0 + 1].role != "tool") %}
            {{- '<|im_end|>\n' }}
        {%- endif %}
    {%- endif %}
{%- endfor %}
{%- if add_generation_prompt %}
    {{- '<|im_start|>assistant\n<think>\n' }}
{%- endif %}
```

## Error Handling
Always check for errors in RedisDB operations.
```go
if err := keyConfig.Set("config1", config, time.Hour*24); err != nil {
    fmt.Println("Error setting config:", err)
}
```

## Further Reading
- Full API documentation for each data type is available in the package.
- Check the package sources for advanced usage.
