// Copyright IBM Corp. 2019, 2020
// SPDX-License-Identifier: MIT

package state

import (
	"strings"
)

func isRepoExcluded(fullName string, name string, excluded []string) bool {
	if len(excluded) == 0 {
		return false
	}

	for _, item := range excluded {
		if strings.EqualFold(item, fullName) || strings.EqualFold(item, name) {
			return true
		}
	}

	return false
}
