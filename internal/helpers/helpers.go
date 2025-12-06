package helpers

import (
	"log/slog"
	"strconv"
	"strings"
)

func Int64ToInterfaceSlice(slice []int64) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}

func StringToInt64Slice(slice []string) ([]int64, error) {
	result := make([]int64, len(slice))
	for i, v := range slice {
		num, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		result[i] = num
	}
	return result, nil
}

// Helper comparison functions
func compareStrings(a, b string, order string) bool {
	if order == "desc" {
		return strings.ToLower(a) > strings.ToLower(b)
	}
	return strings.ToLower(a) < strings.ToLower(b)
}

func compareInts(a, b int, order string) bool {
	if order == "desc" {
		return a > b
	}
	return a < b
}

func compareQuantities(a, b string, order string) bool {
	quantityA, errA := parseQuantity(a)
	quantityB, errB := parseQuantity(b)

	// Handle parsing errors - put unparseable quantities at the end
	if errA != nil && errB != nil {
		return compareStrings(a, b, order)
	}
	if errA != nil {
		return order == "desc" // Put unparseable at the end for asc, at start for desc
	}
	if errB != nil {
		return order != "desc" // Put unparseable at the end for asc, at start for desc
	}

	if order == "desc" {
		return quantityA > quantityB
	}
	return quantityA < quantityB
}

// Utility function to check if slice contains value
func contains(slice []int64, value int64) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// ParseSortString parses sort string like "name" or "-name" into SortConfig
func ParseSortString(sortStr string) SortConfig {
	if sortStr == "" {
		return SortConfig{Field: "name", Order: "asc"}
	}

	config := SortConfig{Order: "asc"}

	if strings.HasPrefix(sortStr, "-") {
		config.Field = sortStr[1:]
		config.Order = "desc"
	} else {
		config.Field = sortStr
	}

	return config
}

// ParseQuantityRange parses min and max quantity strings into float pointers
func ParseQuantityRange(minStr, maxStr string) (*float64, *float64, error) {
	var min, max *float64

	if minStr != "" {
		minVal, err := parseQuantity(minStr)
		if err != nil {
			return nil, nil, err
		}
		min = &minVal
	}

	if maxStr != "" {
		maxVal, err := parseQuantity(maxStr)
		if err != nil {
			return nil, nil, err
		}
		max = &maxVal
	}

	return min, max, nil
}

func ParseLogLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "ERROR":
		return slog.LevelError
	case "WARN", "WARNING":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
