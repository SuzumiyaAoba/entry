package main

import (
	"fmt"
	"os"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/matcher"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: entry <file>")
		os.Exit(1)
	}

	file := os.Args[1]

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	rule, err := matcher.Match(cfg.Rules, file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error matching rule: %v\n", err)
		os.Exit(1)
	}

	if rule == nil {
		fmt.Printf("No matching rule found for %s\n", file)
		os.Exit(0)
	}

	if err := executor.Execute(rule.Command, file); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
