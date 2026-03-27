// Copyright IBM Corp. 2019, 2020
// SPDX-License-Identifier: MIT

package config

import (
	"reflect"
	"testing"
)

func TestParseListEnv(t *testing.T) {
	got := ParseListEnv("org/repo-a, repo-b;repo-c\nrepo-d\r\nrepo-e, ,")
	want := []string{"org/repo-a", "repo-b", "repo-c", "repo-d", "repo-e"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestParseListEnv_Empty(t *testing.T) {
	got := ParseListEnv("   ")
	if len(got) != 0 {
		t.Fatalf("expected empty list, got %v", got)
	}
}
