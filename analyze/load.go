// Copyright IBM Corp. 2019, 2020
// SPDX-License-Identifier: MIT

package analyze

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/gaudit/config"
	"gopkg.in/yaml.v2"
)

func Load(options config.Options) (rules []Rule, err error) {

	if _, err = os.Stat(options.Rules); err != nil {
		if os.IsNotExist(err) {
			return rules, nil
		}
	}

	b, err := ioutil.ReadFile(options.Rules)
	if err != nil {
		return rules, err
	}

	err = yaml.Unmarshal(b, &rules)
	if err != nil {
		normalized := normalizeGlobResourceValues(b)
		if string(normalized) != string(b) {
			retryErr := yaml.Unmarshal(normalized, &rules)
			if retryErr == nil {
				return rules, nil
			}
		}
		return rules, err
	}

	return rules, nil

}

func normalizeGlobResourceValues(input []byte) []byte {
	lines := strings.Split(string(input), "\n")
	changed := false

	for i, line := range lines {
		idx := strings.Index(line, "resource:")
		if idx < 0 {
			continue
		}

		prefix := line[:idx]
		valuePart := strings.TrimSpace(line[idx+len("resource:"):])
		if valuePart == "" {
			continue
		}
		if strings.HasPrefix(valuePart, "\"") || strings.HasPrefix(valuePart, "'") {
			continue
		}

		// YAML interprets unquoted leading *, ?, [ as aliases/tags/flow syntax.
		// Quote these resource values so glob patterns like **/*.yml can be read.
		if strings.HasPrefix(valuePart, "*") || strings.HasPrefix(valuePart, "?") || strings.HasPrefix(valuePart, "[") {
			escaped := strings.ReplaceAll(valuePart, "'", "''")
			lines[i] = prefix + "resource: '" + escaped + "'"
			changed = true
		}
	}

	if !changed {
		return input
	}

	return []byte(strings.Join(lines, "\n"))
}
