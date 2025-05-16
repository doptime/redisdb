package redisdb

import (
	"fmt"
	"reflect"
	"strings"
)

// GoTypeToJSDocType 将 Go 类型转换为 JSDoc 类型字符串。
// 对于结构体，它会生成一个 @typedef 的 JSDoc 注释。
func GoTypeToJSDocType(goType interface{}) (string, error) {
	t := reflect.TypeOf(goType)
	if t == nil {
		return "", fmt.Errorf("input type is nil")
	}
	return generateJSDocType(t, make(map[string]string)), nil
}

func generateJSDocType(t reflect.Type, definedTypes map[string]string) string {
	typeName := t.Name()
	if typeName != "" {
		if jsDocType, ok := definedTypes[typeName]; ok {
			return jsDocType // 如果已经定义过，直接返回
		}
	}

	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		elemType := t.Elem()
		return fmt.Sprintf("Array<%s>", generateJSDocType(elemType, definedTypes))
	case reflect.Map:
		keyType := generateJSDocType(t.Key(), definedTypes)
		elemType := generateJSDocType(t.Elem(), definedTypes)
		return fmt.Sprintf("Object<%s, %s>", keyType, elemType)
	case reflect.Struct:
		if typeName == "" { // 匿名结构体
			typeName = fmt.Sprintf("AnonymousStruct%d", len(definedTypes)+1)
		}
		if _, exists := definedTypes[typeName]; exists {
			return typeName // 防止递归定义时重复生成
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("/**\n * @typedef {%s} %s\n", "Object", typeName))
		definedTypes[typeName] = typeName // 先占位，防止无限递归

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			// 忽略非导出字段
			if field.PkgPath != "" {
				continue
			}
			jsDocFieldType := generateJSDocType(field.Type, definedTypes)
			// 简单处理可选字段：如果类型是指针，我们认为是可选的
			// JSDoc 中通常用 `[typeName]` 或 `typeName|undefined` 表示可选
			// 这里我们简单地不在 JSDoc @property 中显式标记可选性，依赖具体使用场景
			// 或者，如果字段是omitempty，也可以认为是可选的
			// tag := field.Tag.Get("json") // 可以根据json tag判断是否omitempty
			// isOptional := strings.Contains(tag, "omitempty") || field.Type.Kind() == reflect.Ptr

			// JSDoc @property {type} name - description
			// 这里我们不生成 description
			sb.WriteString(fmt.Sprintf(" * @property {%s} %s\n", jsDocFieldType, field.Name))
		}
		sb.WriteString(" */\n")
		// 将完整的 typedef 存储起来
		definedTypes[typeName] = sb.String()
		return typeName // 返回类型名称，具体的 typedef 会在外部收集
	case reflect.Ptr:
		return generateJSDocType(t.Elem(), definedTypes) // JSDoc 通常不直接表示指针，而是表示其指向的类型
	case reflect.Interface:
		return "Object" // interface{} 通常映射为 Object 或 any
	default:
		return "any" // 未知类型默认为 any
	}
}

// GenerateAllJSDocTypeDefs 会处理一个顶层类型，并返回所有相关的 @typedef 定义。
func GenerateAllJSDocTypeDefs(goType interface{}) (string, error) {
	t := reflect.TypeOf(goType)
	if t == nil {
		return "", fmt.Errorf("input type is nil")
	}
	definedTypes := make(map[string]string)
	_ = generateJSDocType(t, definedTypes) // 主调用，填充 definedTypes

	var allTypeDefs strings.Builder
	for _, typeDef := range definedTypes {
		// 确保只添加实际的 typedef 块，而不是简单的类型名称
		if strings.HasPrefix(typeDef, "/**") {
			allTypeDefs.WriteString(typeDef)
			allTypeDefs.WriteString("\n")
		}
	}
	// 如果顶层类型不是结构体，但内部引用了结构体，上面的循环会处理
	// 如果顶层类型本身就是结构体，其 typedef 已经包含在 definedTypes 中
	// 如果顶层类型是简单类型或数组/Map等，它不会生成 @typedef，这是预期的
	// 如果需要为顶层类型也生成一个引用，可以这样做：
	topLevelJSDocName := generateJSDocType(t, make(map[string]string)) // 重新获取顶层类型的JSDoc名称
	if !strings.HasPrefix(topLevelJSDocName, "/**") && allTypeDefs.Len() > 0 {
		// 如果顶层类型不是一个直接的 typedef (比如 Array<MyStruct>),
		// 并且已经生成了一些 typedefs (比如 MyStruct 的)
		// 那么我们不需要为顶层类型本身再生成一个 typedef，除非你希望为它起一个别名
	} else if !strings.HasPrefix(topLevelJSDocName, "/**") && allTypeDefs.Len() == 0 && t.Kind() != reflect.Struct {
		// 如果顶层类型是简单类型，且没有其他嵌套结构体，则不需要 typedef
		return fmt.Sprintf("/** @type {%s} */", topLevelJSDocName), nil
	}

	result := allTypeDefs.String()
	if result == "" && t.Kind() == reflect.Struct { // 如果顶层是结构体，但由于某种原因没有生成（理论上不应该）
		typeDef := generateJSDocType(t, make(map[string]string)) // 再次尝试生成
		if strings.HasPrefix(typeDef, "/**") {
			return typeDef, nil
		}
	} else if result == "" && t.Kind() != reflect.Struct {
		// 如果顶层类型不是结构体，也没有嵌套结构体，直接返回其 JSDoc 类型
		return fmt.Sprintf("/** @type {%s} */", generateJSDocType(t, make(map[string]string))), nil
	}

	return strings.TrimSpace(result), nil
}
