package parse

import "unicode"

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
