package matcher

import (
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/gabriel-vasile/mimetype"
)

func Match(rules []config.Rule, filename string) (*config.Rule, error) {

	// Check Scheme
	u, err := url.Parse(filename)
	isURL := err == nil && u.Scheme != ""

	for _, rule := range rules {
		// Check OS
		if len(rule.OS) > 0 {
			matchedOS := false
			for _, osName := range rule.OS {
				if strings.ToLower(osName) == runtime.GOOS {
					matchedOS = true
					break
				}
			}
			if !matchedOS {
				continue
			}
		}

		// Check Scheme
		if rule.Scheme != "" {
			if isURL && strings.ToLower(u.Scheme) == strings.ToLower(rule.Scheme) {
				return &rule, nil
			}
		}

		// Check extensions (only if not a URL, or if URL has extension?)
		// URLs can have extensions (e.g. image.png).
		// But if it's a URL, we might want to prioritize Scheme or Regex.
		// Let's check extensions for URLs too.
		if len(rule.Extensions) > 0 {
			// For URL, get path extension
			var pathExt string
			if isURL {
				pathExt = filepath.Ext(u.Path)
			} else {
				pathExt = filepath.Ext(filename)
			}
			pathExt = strings.ToLower(strings.TrimPrefix(pathExt, "."))

			for _, ruleExt := range rule.Extensions {
				if strings.ToLower(ruleExt) == pathExt {
					return &rule, nil
				}
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

		// Check MIME type (Only for files, unless we fetch URL?)
		// Skip MIME check for URLs for now
		if rule.Mime != "" && !isURL {
			mtype, err := mimetype.DetectFile(filename)
			if err != nil {
				// If file cannot be read, ignore MIME match? Or return error?
				// For now, ignore and continue to next rule.
				continue
			}
			// Use regex for MIME match? Or exact match?
			// Let's use regex for flexibility (e.g. "image/.*").
			matched, err := regexp.MatchString(rule.Mime, mtype.String())
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
