package utils

import "reflect"

func CreateNonNilInstance[T any]() T {
	var zero T
	typ := reflect.TypeOf(&zero).Elem() // 获取 T 的反射类型

	// 1. 如果 T 本身不是指针，直接返回零值即可（结构体零值即为可用实例）
	// 或者如果你希望结构体字段也有默认值，这里需要额外逻辑，但通常零值够用。
	if typ.Kind() != reflect.Ptr {
		return zero
	}

	// 2. 剥离指针，寻找最底层的 Base Type
	// 例如：如果是 ***int，我们需要找到 int
	baseType := typ
	for baseType.Kind() == reflect.Ptr {
		baseType = baseType.Elem()
	}

	// 3. 实例化底层类型
	// reflect.New(baseType) 返回的是 *BaseType (即指向底层的指针)
	// 例如：baseType 是 int，这里得到 *int
	val := reflect.New(baseType)

	// 4. 重新包装指针层级
	// 我们现在的 val 是 *BaseType。
	// 如果目标 T 是 **BaseType，我们需要创建一个指向 val 的指针。
	// 我们需要一直包装，直到 val 的类型等于 T。
	for val.Type() != typ {
		// 创建一个指向当前 val 的新指针
		ptr := reflect.New(val.Type())
		// 将新指针的值设置为当前 val
		ptr.Elem().Set(val)
		// 更新 val 为这个新指针
		val = ptr
	}

	// 5. 转回具体的泛型类型 T
	return val.Interface().(T)
}
