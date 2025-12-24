package redisdb

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// GoTypeToTypeScriptInterface 将 Go 结构体类型转换为 TypeScript 接口字符串。
// 它也会处理嵌套的结构体。
func GoTypeToTypeScriptInterface(goType interface{}) (string, error) {
	t := reflect.TypeOf(goType)
	if t == nil {
		return "", fmt.Errorf("input type is nil")
	}

	// 如果是指针，获取其元素类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 主要目标是转换结构体为接口
	if t.Kind() != reflect.Struct {
		// 如果不是结构体，但想获得其 TypeScript 类型表示，可以调用下面的辅助函数
		tsType, definedInterfaces := generateTypeScriptType(t, make(map[string]string))
		if len(definedInterfaces) > 0 {
			var sb strings.Builder
			for _, def := range definedInterfaces {
				sb.WriteString(def)
				sb.WriteString("\n\n")
			}
			// 如果顶层类型本身不是interface，但其内部有interface定义
			// 并且我们希望顶层类型也有一个别名，可以这样：
			// if !strings.Contains(tsType, "interface ") {
			//  sb.WriteString(fmt.Sprintf("type %s = %s;", t.Name()+"Type", tsType))
			// }
			return strings.TrimSpace(sb.String()), nil
		}
		return fmt.Sprintf("type %s = %s;", t.Name()+"Type", tsType), nil // 对于非结构体，返回一个类型别名
	}

	interfaces := make(map[string]string)
	generateTypeScriptInterfaces(t, interfaces)

	var result strings.Builder
	// 按名称排序以获得一致的输出（可选）
	// var keys []string
	// for k := range interfaces {
	// 	keys = append(keys, k)
	// }
	// sort.Strings(keys)
	// for _, k := range keys {
	// 	result.WriteString(interfaces[k])
	// 	result.WriteString("\n\n")
	// }
	// 通常我们希望先输出依赖的接口，再输出主接口，但这里的实现是基于首次遇到进行定义
	// 更好的方式是先收集所有类型，然后按依赖顺序或者字母顺序输出
	// 为简单起见，这里直接按发现顺序（map迭代顺序不保证）输出
	mainInterfaceName := formatTypeNameForTypeScript(t.Name())
	if mainDef, ok := interfaces[mainInterfaceName]; ok {
		result.WriteString(mainDef) // 先写主接口
		result.WriteString("\n\n")
	}
	for name, def := range interfaces {
		if name != mainInterfaceName { // 避免重复写入主接口
			result.WriteString(def)
			result.WriteString("\n\n")
		}
	}

	return strings.TrimSpace(result.String()), nil
}

func formatTypeNameForTypeScript(name string) string {
	if name == "" {
		return "AnonymousInterface" // 或者生成一个唯一的名称
	}
	return name // 可以根据需要添加 "I" 前缀等，如 "I" + name
}

// generateTypeScriptType 返回类型的 TypeScript 表示和一个包含所有新定义的接口的 map
func generateTypeScriptType(t reflect.Type, definedInterfaces map[string]string) (string, map[string]string) {
	// 处理 time.Time
	if t == reflect.TypeOf(time.Time{}) {
		return "Date", definedInterfaces // 通常时间会序列化为 ISO 字符串，或者在 JS 中用 Date 对象
	}
	if t == reflect.TypeOf((*time.Time)(nil)).Elem() { // *time.Time
		return "Date | null", definedInterfaces
	}

	switch t.Kind() {
	case reflect.String:
		return "string", definedInterfaces
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number", definedInterfaces
	case reflect.Bool:
		return "boolean", definedInterfaces
	case reflect.Slice, reflect.Array:
		elemType, newInterfaces := generateTypeScriptType(t.Elem(), definedInterfaces)
		definedInterfaces = newInterfaces // 更新已定义的接口
		return fmt.Sprintf("%s[]", elemType), definedInterfaces
	case reflect.Map:
		keyType, newInterfacesKey := generateTypeScriptType(t.Key(), definedInterfaces)
		definedInterfaces = newInterfacesKey
		elemType, newInterfacesElem := generateTypeScriptType(t.Elem(), definedInterfaces)
		definedInterfaces = newInterfacesElem
		// TypeScript 的 Record 类型更适合表示 Go 的 map
		return fmt.Sprintf("Record<%s, %s>", keyType, elemType), definedInterfaces
	case reflect.Struct:
		typeName := formatTypeNameForTypeScript(t.Name())
		if typeName == "AnonymousInterface" { // 对于匿名结构体，尝试生成唯一的名称
			typeName = fmt.Sprintf("AnonymousInterface%d", len(definedInterfaces)+1)
		}
		if _, exists := definedInterfaces[typeName]; !exists {
			// 预先放置一个标记，表示正在生成，以处理循环引用
			definedInterfaces[typeName] = fmt.Sprintf("interface %s { /* cyclic */ }", typeName)
			generateTypeScriptInterfaces(t, definedInterfaces) // 这会填充真正的接口定义
		}
		return typeName, definedInterfaces
	case reflect.Ptr:
		elemType, newInterfaces := generateTypeScriptType(t.Elem(), definedInterfaces)
		definedInterfaces = newInterfaces
		return fmt.Sprintf("%s | null", elemType), definedInterfaces // 指针在 TypeScript 中通常表示为 T | null 或 T | undefined
	case reflect.Interface:
		// interface{} 在 Go 中是任意类型，对应 TypeScript 的 any 或 unknown
		// unknown 更安全，但 any 更常见于直接转换
		if t.NumMethod() == 0 { // 空接口 interface{}
			return "any", definedInterfaces
		}
		// 对于有方法的接口，转换会更复杂，这里简化处理
		return "any /* interface with methods */", definedInterfaces
	default:
		return "any", definedInterfaces
	}
}

func generateTypeScriptInterfaces(t reflect.Type, interfaces map[string]string) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return // 只为结构体生成接口
	}

	typeName := formatTypeNameForTypeScript(t.Name())
	if typeName == "AnonymousInterface" { // 对于匿名结构体，尝试生成唯一的名称
		typeName = fmt.Sprintf("AnonymousInterface%d", len(interfaces)+1)
	}

	// 检查是否已经生成或正在生成，以避免重复和循环（generateTypeScriptType 中已有初步处理）
	if def, exists := interfaces[typeName]; exists && !strings.Contains(def, "/* cyclic */") {
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("interface %s {\n", typeName))

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// 忽略非导出字段
		if field.PkgPath != "" {
			continue
		}

		fieldName := field.Name // 可以根据 json tag 获取实际的字段名
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			}
			if parts[0] == "-" { // 字段被忽略
				continue
			}
		}

		fieldType, newInterfaces := generateTypeScriptType(field.Type, interfaces)
		interfaces = newInterfaces // 更新已定义的接口

		// 处理可选字段: 如果是 Go 指针，或者 json tag 中有 omitempty
		isOptional := field.Type.Kind() == reflect.Ptr
		if jsonTag != "" {
			if strings.Contains(jsonTag, "omitempty") {
				isOptional = true
			}
		}

		if isOptional {
			// 对于已经是 T | null 的类型，不再加 ?
			if strings.HasSuffix(fieldType, "| null") {
				sb.WriteString(fmt.Sprintf("  %s: %s;\n", fieldName, fieldType))
			} else {
				sb.WriteString(fmt.Sprintf("  %s?: %s;\n", fieldName, fieldType))
			}
		} else {
			sb.WriteString(fmt.Sprintf("  %s: %s;\n", fieldName, fieldType))
		}
	}
	sb.WriteString("}")
	interfaces[typeName] = sb.String()

	// 递归处理字段类型中的结构体，确保它们也被定义
	// （generateTypeScriptType 内部已经递归调用了 generateTypeScriptInterfaces）
}
