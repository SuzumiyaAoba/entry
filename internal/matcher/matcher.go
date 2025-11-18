package matcher

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/SuzumiyaAoba/entry/internal/config"
)

func Match(rules []config.Rule, filename string) (*config.Rule, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	// Remove dot from extension if present for comparison, or keep it?
	// Usually config might have ".txt" or "txt". Let's handle both or assume one.
	// Let's assume config has "txt" (no dot) for simplicity, or handle both.
	// Better: trim dot from file ext.
	ext = strings.TrimPrefix(ext, ".")

	for _, rule := range rules {
		// Check extensions
		for _, ruleExt := range rule.Extensions {
			if strings.ToLower(ruleExt) == ext {
				return &rule, nil
			}
		}

		// Check regex
		if rule.Regex != "" {
			matched, err := regexp.MatchString(rule.Regex, filename)
			if err != nil {
				return nil, err
			}
			if matched {
				return &rule, nil
			}
		}
	}

	return nil, nil
}
