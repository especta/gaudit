// Copyright IBM Corp. 2019, 2020
// SPDX-License-Identifier: MIT

package config

import (
	"strings"
)

// ParseListEnv converts comma/semicolon/newline separated env values into a trimmed list.
func ParseListEnv(value string) []string {
	normalized := strings.ReplaceAll(value, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.ReplaceAll(normalized, ";", ",")
	normalized = strings.ReplaceAll(normalized, "\n", ",")

	items := strings.Split(normalized, ",")
	result := []string{}
	for _, item := range items {
		v := strings.TrimSpace(item)
		if v != "" {
			result = append(result, v)
		}
	}

	return result
}
