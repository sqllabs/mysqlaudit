// Copyright (C) 2025 JustCoding247. All rights reserved.
package session_test

import (
	"os"
	"strconv"
	"strings"
)

func getEnvWithDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getTestHost() string {
	return getEnvWithDefault("TEST_MYSQL_HOST", "127.0.0.1")
}

func getTestPort() uint {
	val := getEnvWithDefault("TEST_MYSQL_PORT", "3306")
	if p, err := strconv.Atoi(val); err == nil && p > 0 {
		return uint(p)
	}
	return 3306
}

func getTestUser() string {
	return getEnvWithDefault("TEST_MYSQL_USER", "test")
}

func getTestPassword() string {
	return getEnvWithDefault("TEST_MYSQL_PASSWORD", "test")
}

func getTestDBName() string {
	return getEnvWithDefault("TEST_MYSQL_DB", "test_inc")
}

func normalizeSchemaComponent(input string) string {
	replacer := strings.NewReplacer(".", "_", ":", "_")
	return replacer.Replace(input)
}
