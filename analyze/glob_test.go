// Copyright IBM Corp. 2019, 2020
// SPDX-License-Identifier: MIT

package analyze

import (
	"reflect"
	"testing"
)

func TestFilterResourcesByPattern(t *testing.T) {
	resources := []string{
		"README.md",
		"docs/guide.md",
		"docs/intro/getting-started.md",
		"src/main.go",
		"src/app/config.yml",
	}

	cases := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "single level wildcard",
			pattern: "docs/*.md",
			want:    []string{"docs/guide.md"},
		},
		{
			name:    "recursive wildcard",
			pattern: "**/*.md",
			want:    []string{"README.md", "docs/guide.md", "docs/intro/getting-started.md"},
		},
		{
			name:    "exact path still works",
			pattern: "src/main.go",
			want:    []string{"src/main.go"},
		},
		{
			name:    "recursive with filename",
			pattern: "src/**/config.yml",
			want:    []string{"src/app/config.yml"},
		},
	}

	for _, tc := range cases {
		got, err := filterResourcesByPattern(tc.pattern, resources)
		if err != nil {
			t.Fatalf("%s: unexpected error: %s", tc.name, err.Error())
		}
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, got)
		}
	}
}

func TestFilterResourcesByPattern_InvalidPattern(t *testing.T) {
	_, err := filterResourcesByPattern("docs/[*.md", []string{"docs/guide.md"})
	if err == nil {
		t.Fatal("expected invalid glob pattern error")
	}
}
