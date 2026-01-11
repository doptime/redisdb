package utils

import "reflect"

// CreateNonNilInstance creates a non-nil instance of type T.
// If T is a pointer (even deeply nested like **Struct), it recursively
// initializes the underlying type and re-wraps the pointers.
func CreateNonNilInstance[T any]() T {
	var zero T
	typ := reflect.TypeOf(&zero).Elem()

	// 1. If T is not a pointer, return the zero value (e.g., empty struct).
	if typ.Kind() != reflect.Ptr {
		return zero
	}

	// 2. Unwrap pointers to find the underlying base type.
	baseType := typ
	for baseType.Kind() == reflect.Ptr {
		baseType = baseType.Elem()
	}

	// 3. Instantiate the base type.
	// reflect.New returns a pointer to the type (*BaseType).
	val := reflect.New(baseType)

	// 4. Re-wrap the pointer layers until the type matches T.
	for val.Type() != typ {
		ptr := reflect.New(val.Type())
		ptr.Elem().Set(val)
		val = ptr
	}

	// 5. Cast back to generic type T and return.
	return val.Interface().(T)
}
