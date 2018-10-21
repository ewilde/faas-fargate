package system

import "os"

// GetEnv returns the environment variable or the fallback value
func GetEnv(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}
