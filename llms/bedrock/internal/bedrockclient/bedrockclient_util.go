package bedrockclient

import (
	"reflect"
)

// isSmithyValidObject validates that an object contains only simple types
// or structs with proper document tags on all public fields
func isSmithyValidObject(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)

	if isSimpleType(t) {
		return true
	}

	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return true
		}
		return isSmithyValidObject(v.Elem().Interface())
	}

	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			if !field.IsExported() {
				continue
			}

			if _, hasTag := field.Tag.Lookup("document"); !hasTag {
				return false
			}

			fieldValue := v.Field(i).Interface()
			if !isSmithyValidObject(fieldValue) {
				return false
			}
		}
		return true
	}

	if t.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			elementValue := v.Index(i).Interface()
			if !isSmithyValidObject(elementValue) {
				return false
			}
		}
		return true
	}

	if t.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			// Check key validity
			if !isSmithyValidObject(key.Interface()) {
				return false
			}
			// Check value validity
			mapValue := v.MapIndex(key).Interface()
			if !isSmithyValidObject(mapValue) {
				return false
			}
		}
		return true
	}

	return false
}

func isSimpleType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	default:
		return false
	}
}
