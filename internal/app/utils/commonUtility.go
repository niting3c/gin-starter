package utils

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// StringContains checks if a string is present in a slice of strings.
// Returns true if the string is found; otherwise, returns false.
func StringContains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func IntContains(slice []int64, i int64) bool {
	for _, item := range slice {
		if item == i {
			return true
		}
	}
	return false
}

func ConvertToInt64(val interface{}) (int64, error) {
	switch v := val.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case float64:
		// Note: This conversion can lead to precision loss
		return int64(v), nil
	case float32:
		// Note: This conversion can lead to precision loss
		return int64(v), nil
	case string:
		// Attempt to parse the string as an int64
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported type: %T", val)
	}
}

type RowMapperFunc func(pgx.Row) (interface{}, error)

func GetEnvAsInt(name string, defaultVal int) int {
	value, exists := os.LookupEnv(name)
	if !exists {
		return defaultVal
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	return intValue
}

func GetEnvAsString(name string, defaultVal string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		return defaultVal
	}
	return value
}
