package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/spf13/cobra"
)


var (
	cfgFile string
	dryRun  bool
)

func init() {
	rootCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/entry/config.yml)")
	rootCmd.Flags().Bool("dry-run", false, "print command instead of executing")
	rootCmd.Flags().BoolP("select", "s", false, "Interactive selection")
	rootCmd.Flags().Bool("explain", false, "Show detailed matching information")

	rootCmd.AddCommand(configCmd)
}

var rootCmd = &cobra.Command{
	Use:   "et <file>",
	Short: "Entry is a CLI file association tool",
	Long:  `Entry allows you to execute specific commands based on file extensions or regex patterns matched against a provided file argument.`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Manual flag parsing
		var commandArgs []string
		for i := 0; i < len(args); i++ {
			arg := args[i]
			if arg == "--version" || arg == "-v" {
				fmt.Println("et v0.1.0")
				return nil
			}
			if arg == "--help" || arg == "-h" {
				return cmd.Help()
			}
			if arg == "--dry-run" {
				dryRun = true
				continue
			} else if arg == "--config" {
				if i+1 < len(args) {
					cfgFile = args[i+1]
					i++ // Skip next arg
					continue
				} else {
					return fmt.Errorf("flag needs an argument: --config")
				}
			} else if strings.HasPrefix(arg, "--config=") {
				cfgFile = strings.TrimPrefix(arg, "--config=")
				continue
			} else if arg == "--select" || arg == "-s" || arg == "--explain" {
				// These flags are handled later, skip them here
				continue
			} else if arg == "--" { // Stop parsing at first non-flag or "--"
				commandArgs = args[i+1:]
				break
			} else if !strings.HasPrefix(arg, "-") {
				commandArgs = args[i:]
				break
			}
			// Unknown flag - treat as part of command args
			commandArgs = args[i:]
			break
		}

		if len(commandArgs) == 0 {
			return fmt.Errorf("requires at least 1 argument")
		}

		// Check for built-in commands (manual dispatch)
		if commandArgs[0] == "config" {
			return handleConfigCommand(commandArgs[1:])
		}

		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		// Initialize Executor
		exec := executor.NewExecutor(cmd.OutOrStdout(), dryRun)

		// Check for mode flags
		interactive := false
		explain := false
		for _, arg := range args {
			if arg == "--select" || arg == "-s" {
				interactive = true
			}
			if arg == "--explain" {
				explain = true
			}
		}

		// Explain mode: show detailed matching information
		if explain && len(commandArgs) == 1 {
			return handleExplain(cmd, cfg, commandArgs[0])
		}

		// Handle file execution with single argument
		if len(commandArgs) == 1 {
			filename := commandArgs[0]
			
			// Interactive mode
			if interactive {
				return handleInteractive(cfg, exec, filename)
			}

			// Normal file execution - if it fails, try as command
			if err := handleFileExecution(cfg, exec, filename); err == nil {
				return nil
			}
			// If file execution failed, fall through to command execution
		}

		// Handle command execution with multiple arguments
		return handleCommandExecution(cfg, exec, commandArgs)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
