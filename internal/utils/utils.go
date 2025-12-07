package utils

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// SplitAndTrim splits a comma-separated string and trims whitespace
func SplitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	return lo.FilterMap(strings.Split(s, ","), func(part string, _ int) (string, bool) {
		trimmed := strings.TrimSpace(part)
		return trimmed, trimmed != ""
	})
}

// ParseEnvList parses a list of "KEY=VALUE" strings into a map
func ParseEnvList(list []string) map[string]string {
	if len(list) == 0 {
		return nil
	}
	env := make(map[string]string)
	for _, item := range list {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
}

// ParseEnvString parses a comma-separated string of "KEY=VALUE" into a map
func ParseEnvString(s string) map[string]string {
	if s == "" {
		return nil
	}
	// Split by comma, respecting quotes would be better but simple comma split for now consistent with OS list
	list := SplitAndTrim(s)
	return ParseEnvList(list)
}

// FormatEnvMap formats a map of environment variables into a comma-separated string
func FormatEnvMap(env map[string]string) string {
	if len(env) == 0 {
		return ""
	}
	var parts []string
	for k, v := range env {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
}
