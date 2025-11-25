package matcher

import (
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/dop251/goja"
	"github.com/gabriel-vasile/mimetype"
	"github.com/samber/lo"
)

func Match(rules []config.Rule, filename string) ([]*config.Rule, error) {
	var matches []*config.Rule
	
	for i := range rules {
		rule := &rules[i]
		matched, err := matchRule(rule, filename)
		if err != nil {
			return nil, err
		}

		if matched {
			matches = append(matches, rule)
			if !rule.Fallthrough {
				break
			}
		}
	}

	return matches, nil
}


func MatchAll(rules []config.Rule, filename string) ([]*config.Rule, error) {
	var matches []*config.Rule

	for i := range rules {
		rule := &rules[i]
		matched, err := matchRule(rule, filename)
		if err != nil {
			return nil, err
		}

		if matched {
			matches = append(matches, rule)
		}
	}

	return matches, nil
}

func matchRule(rule *config.Rule, filename string) (bool, error) {
	// Parse URL once if needed? 
	// Actually, we can parse it inside here. It's cheap enough.
	u, err := url.Parse(filename)
	isURL := err == nil && u.Scheme != ""

	// Check OS
	if len(rule.OS) > 0 {
		if !lo.ContainsBy(rule.OS, func(osName string) bool {
			return strings.EqualFold(osName, runtime.GOOS)
		}) {
			return false, nil
		}
	}

	// Check Scheme
	if rule.Scheme != "" {
		if isURL && strings.EqualFold(u.Scheme, rule.Scheme) {
			return true, nil
		}
		// If scheme is specified but doesn't match, this rule is not a match
		return false, nil
	}

	// Check extensions
	if len(rule.Extensions) > 0 {
		var pathExt string
		if isURL {
			pathExt = filepath.Ext(u.Path)
		} else {
			pathExt = filepath.Ext(filename)
		}
		pathExt = strings.ToLower(strings.TrimPrefix(pathExt, "."))

		if lo.ContainsBy(rule.Extensions, func(ruleExt string) bool {
			return strings.EqualFold(ruleExt, pathExt)
		}) {
			return true, nil
		}
		// If extensions are specified but none matched, we continue to check other conditions (Regex, MIME, etc.)
		// This allows a rule to match EITHER by extension OR by regex/mime.
	}

	// Check regex
	if rule.Regex != "" {
		regexMatched, err := regexp.MatchString(rule.Regex, filename)
		if err != nil {
			return false, err
		}
		if regexMatched {
			return true, nil
		}
	}

	// Check MIME type
	if rule.Mime != "" && !isURL {
		mtype, err := mimetype.DetectFile(filename)
		if err == nil {
			mimeMatched, err := regexp.MatchString(rule.Mime, mtype.String())
			if err == nil && mimeMatched {
				return true, nil
			}
		}
	}

	// Check Script
	if rule.Script != "" {
		return matchScript(rule.Script, filename)
	}

	return false, nil
}

func matchScript(script string, filename string) (bool, error) {
	vm := goja.New()
	vm.Set("file", filename)
	
	val, err := vm.RunString(script)
	if err != nil {
		return false, err
	}
	
	return val.ToBoolean(), nil
}
