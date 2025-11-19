package main

import (
	"fmt"
	"net/url"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/matcher"
	"github.com/charmbracelet/huh"
)

type Option struct {
	Label    string
	Rule     *config.Rule
	IsSystem bool
}

func handleInteractive(cfg *config.Config, exec *executor.Executor, filename string) error {
	matches, err := matcher.MatchAll(cfg.Rules, filename)
	if err != nil {
		return fmt.Errorf("error matching rules: %w", err)
	}

	var options []Option

	for _, m := range matches {
		label := m.Name
		if label == "" {
			label = fmt.Sprintf("Command: %s", m.Command)
		}
		options = append(options, Option{Label: label, Rule: m})
	}

	// Add System Default if applicable
	// Check if it is a URL or File to decide if system default is valid option
	isURL := false
	if u, err := url.Parse(filename); err == nil && u.Scheme != "" {
		isURL = true
	}
	if isURL || fileExists(filename) {
		options = append(options, Option{Label: "System Default", IsSystem: true})
	}

	if len(options) == 0 {
		return fmt.Errorf("no matching rules found for %s", filename)
	}

	var selected Option
	
	// Use huh for selection
	var huhOptions []huh.Option[Option]
	for _, opt := range options {
		huhOptions = append(huhOptions, huh.NewOption(opt.Label, opt))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[Option]().
				Title("Select action for " + filename).
				Options(huhOptions...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if selected.IsSystem {
		if cfg.DefaultCommand != "" {
			return exec.Execute(cfg.DefaultCommand, filename, executor.ExecutionOptions{})
		}
		return exec.OpenSystem(filename)
	} else {
		opts := executor.ExecutionOptions{
			Background: selected.Rule.Background,
			Terminal:   selected.Rule.Terminal,
		}
		return exec.Execute(selected.Rule.Command, filename, opts)
	}
}
