// Copyright IBM Corp. 2019, 2020
// SPDX-License-Identifier: MIT

package analyze

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/gaudit/config"
)

func TestLoad_UnquotedLeadingGlobPattern(t *testing.T) {
	dir := t.TempDir()
	rulesPath := filepath.Join(dir, "rules.yml")
	content := "-\n  name: Glob Rule\n  action: exists\n  resource: */checksum.txt\n"
	if err := writeFile(rulesPath, content); err != nil {
		t.Fatal(err)
	}

	rules, err := Load(config.Options{Rules: rulesPath})
	if err != nil {
		t.Fatalf("expected rules to load, got error: %s", err.Error())
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Resource != "*/checksum.txt" {
		t.Fatalf("expected resource '*/checksum.txt', got %q", rules[0].Resource)
	}
}

func TestLoad_QuotedGlobPattern(t *testing.T) {
	dir := t.TempDir()
	rulesPath := filepath.Join(dir, "rules.yml")
	content := "-\n  name: Glob Rule\n  action: exists\n  resource: \"**/*.yml\"\n"
	if err := writeFile(rulesPath, content); err != nil {
		t.Fatal(err)
	}

	rules, err := Load(config.Options{Rules: rulesPath})
	if err != nil {
		t.Fatalf("expected rules to load, got error: %s", err.Error())
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Resource != "**/*.yml" {
		t.Fatalf("expected resource '**/*.yml', got %q", rules[0].Resource)
	}
}

func writeFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0600)
}
