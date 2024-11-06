package database

import (
	"strings"
)

var (
	keySeparator = ":"
)

// ConstructKey construct key name from string slice
func ConstructKey(keys ...string) string {
	return strings.Join(keys, keySeparator)
}
