// Copyright IBM Corp. 2019, 2020
// SPDX-License-Identifier: MIT

package analyze

import (
	"path"
	"sort"
	"strings"
)

func hasGlobPattern(resource string) bool {
	return strings.ContainsAny(resource, "*?[")
}

func normalizeResourcePath(resource string) string {
	normalized := strings.TrimSpace(resource)
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimPrefix(normalized, "/")
	if normalized == "" {
		return ""
	}

	cleaned := path.Clean(normalized)
	if cleaned == "." {
		return ""
	}

	return cleaned
}

func filterResourcesByPattern(pattern string, resources []string) ([]string, error) {
	matched := []string{}
	normalizedPattern := normalizeResourcePath(pattern)

	for _, resource := range resources {
		normalizedResource := normalizeResourcePath(resource)
		ok, err := matchResourcePattern(normalizedPattern, normalizedResource)
		if err != nil {
			return matched, err
		}
		if ok {
			matched = append(matched, normalizedResource)
		}
	}

	sort.Strings(matched)
	return matched, nil
}

func matchResourcePattern(pattern string, resource string) (bool, error) {
	patternSegments := splitPathSegments(pattern)
	resourceSegments := splitPathSegments(resource)
	return matchResourceSegments(patternSegments, resourceSegments)
}

func splitPathSegments(v string) []string {
	if v == "" {
		return []string{}
	}
	return strings.Split(v, "/")
}

func matchResourceSegments(pattern []string, resource []string) (bool, error) {
	if len(pattern) == 0 {
		return len(resource) == 0, nil
	}

	if pattern[0] == "**" {
		// collapse repeated ** for faster matching
		for len(pattern) > 1 && pattern[1] == "**" {
			pattern = pattern[1:]
		}

		if ok, err := matchResourceSegments(pattern[1:], resource); ok || err != nil {
			return ok, err
		}

		if len(resource) == 0 {
			return false, nil
		}
		return matchResourceSegments(pattern, resource[1:])
	}

	if len(resource) == 0 {
		return false, nil
	}

	ok, err := path.Match(pattern[0], resource[0])
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	return matchResourceSegments(pattern[1:], resource[1:])
}
