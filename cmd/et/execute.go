package main

import (
	"fmt"
	"net/url"
	"os"
	os_exec "os/exec"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/matcher"
)

func handleFileExecution(cfg *config.Config, exec *executor.Executor, filename string) error {
	// Non-interactive normal flow
	rules, err := matcher.Match(cfg.Rules, filename)
	if err != nil {
		return fmt.Errorf("error matching rule: %w", err)
	}
	if len(rules) > 0 {
		// Execute all matched rules (fallthrough support)
		for _, rule := range rules {
			opts := executor.ExecutionOptions{
				Background: rule.Background,
				Terminal:   rule.Terminal,
			}
			if err := exec.Execute(rule.Command, filename, opts); err != nil {
				return err
			}
		}
		return nil
	}

	// Check if it is a URL or File
	isURL := false
	if u, err := url.Parse(filename); err == nil && u.Scheme != "" {
		isURL = true
	}
	
	if isURL {
		if cfg.DefaultCommand != "" {
			return exec.Execute(cfg.DefaultCommand, filename, executor.ExecutionOptions{})
		}
		return exec.OpenSystem(filename)
	}

	if _, err := os.Stat(filename); err == nil {
		if cfg.DefaultCommand != "" {
			return exec.Execute(cfg.DefaultCommand, filename, executor.ExecutionOptions{})
		}
		return exec.OpenSystem(filename)
	}

	// File not found - caller should handle this as a command
	return fmt.Errorf("file not found and no matching rule")
}

func handleCommandExecution(cfg *config.Config, exec *executor.Executor, commandArgs []string) error {
	command := commandArgs[0]
	cmdArgs := commandArgs[1:]

	// Check aliases
	if alias, ok := cfg.Aliases[command]; ok {
		command = alias
		return exec.ExecuteCommand(command, cmdArgs)
	}

	// Fallback to command execution
	// Check if command exists in PATH
	if _, err := os_exec.LookPath(command); err != nil {
		// Command not found.
		// If single argument and default command exists, assume it's a new file and use default command.
		if len(commandArgs) == 1 && cfg.DefaultCommand != "" {
			return exec.Execute(cfg.DefaultCommand, commandArgs[0], executor.ExecutionOptions{})
		}
	}

	return exec.ExecuteCommand(command, cmdArgs)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
