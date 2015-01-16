package parse

import (
	"reflect"
	"strings"
	"unicode"
)

// Returns the provided string with the first letter upper-cased
func firstToUpper(s string) string {
	if len(s) < 1 {
		return s
	}
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
}

// Returns the provided string with the first letter lower-cased
func firstToLower(s string) string {
	if len(s) < 1 {
		return s
	}
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}

// parses struct tags in the format:
// parse:"name,option"
//
// and returns each component
func parseTag(tag string) (name, options string) {
	parts := strings.Split(tag, ",")
	if len(parts) > 1 {
		return parts[0], parts[1]
	} else {
		return parts[0], ""
	}
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func canBeNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return true
	default:
		return false
	}
}
