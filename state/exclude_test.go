// Copyright IBM Corp. 2019, 2020
// SPDX-License-Identifier: MIT

package state

import "testing"

func TestIsRepoExcluded(t *testing.T) {
	excluded := []string{"org/repo-a", "repo-b"}

	if !isRepoExcluded("org/repo-a", "repo-a", excluded) {
		t.Fatal("expected repo-a to be excluded by full name")
	}
	if !isRepoExcluded("org/repo-b", "repo-b", excluded) {
		t.Fatal("expected repo-b to be excluded by short name")
	}
	if !isRepoExcluded("ORG/REPO-A", "REPO-A", excluded) {
		t.Fatal("expected case-insensitive exclusion to match")
	}
	if isRepoExcluded("org/repo-c", "repo-c", excluded) {
		t.Fatal("did not expect repo-c to be excluded")
	}
}
