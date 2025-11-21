package cli

import (
	"fmt"
	"os"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/logger"
	"github.com/spf13/cobra"
)


var (
	cfgFile     string
	dryRun      bool
	interactive bool
	explain     bool
	verbose     bool
	profile     string
)

func init() {
	rootCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/entry/config.yml)")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "print command instead of executing")
	rootCmd.Flags().BoolVarP(&interactive, "select", "s", false, "Interactive selection")
	rootCmd.Flags().BoolVar(&explain, "explain", false, "Show detailed matching information")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().StringVarP(&profile, "profile", "p", "", "Configuration profile to use")
	rootCmd.RegisterFlagCompletionFunc("profile", CompletionProfiles)
	
	// Allow flags after positional arguments to be passed to the command
	rootCmd.Flags().SetInterspersed(false)

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(completionCmd)
}

var rootCmd = &cobra.Command{
	Use:     "et <file>",
	Short:   "Entry is a CLI file association tool",
	Long:    `Entry allows you to execute specific commands based on file extensions or regex patterns matched against a provided file argument.`,
	Version: Version,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Manually parse flags
		if err := cmd.Flags().Parse(args); err != nil {
			return err
		}

		// Handle help flag manually since parsing is disabled
		if helpVal, _ := cmd.Flags().GetBool("help"); helpVal {
			return cmd.Help()
		}
		
		// Handle version flag manually
		if versionVal, _ := cmd.Flags().GetBool("version"); versionVal {
			fmt.Printf("et version %s\n", Version)
			return nil
		}

		// Check if we have a double dash separator at the beginning (meaning explicit subcommand)
		// ArgsLenAtDash returns -1 if no dash, or the index.
		// If we have `et -- config`, args will be ["config"] after parsing flags?
		// Wait, DisableFlagParsing: true means args contains EVERYTHING including flags.
		// But cmd.Flags().Parse(args) will consume flags.
		// Let's check how Cobra handles this.
		
		// Actually, with DisableFlagParsing: true, args contains everything.
		// We need to parse flags ourselves.
		// After cmd.Flags().Parse(args), we can get remaining args.
		remainingArgs := cmd.Flags().Args()
		
		// Check if the command was invoked with `--` to separate flags from subcommand
		// cmd.Flags().ArgsLenAtDash() tells us where the dash was.
		dashLen := cmd.Flags().ArgsLenAtDash()
		
		// If dashLen is 0, it means `et -- subcommand ...` (assuming no flags before --)
		// Or `et -v -- subcommand ...` -> dashLen will be index of --
		
		// If we have a dash, and it's separating flags from the rest
		if dashLen != -1 {
			// The args after dash are potential subcommands
			// But wait, cmd.Flags().Parse() might have already consumed args before dash as flags.
			// If we use `et -- config`, args is ["--", "config"].
			// Parse() handles `--` as terminator.
			
			// Let's look at how we want to behave:
			// et config -> treat "config" as file
			// et -- config -> treat "config" as subcommand
			
			// If we use DisableFlagParsing=true, we get all args.
			// If we call cmd.Flags().Parse(args), it parses flags.
			// If there was a `--`, Parse stops there.
			
			// We need to know if `--` was present and used to separate subcommand.
			// If ArgsLenAtDash() != -1, it means `--` was present.
			
			if dashLen != -1 {
				// We have a double dash.
				// The arguments AFTER the dash are in remainingArgs?
				// Actually, Parse() puts non-flag args in Args().
				// If we have `et -- config`, Parse() sees `--` and stops. "config" is in Args().
				
				// We want to treat the first arg after `--` as a subcommand if possible.
				if len(remainingArgs) > 0 {
					subCmd, subArgs, err := cmd.Find(remainingArgs)
					if err == nil && subCmd != cmd {
						// Found a subcommand!
						// We need to execute it.
						// We pass the found args to the subcommand
						subCmd.SetArgs(subArgs)
						return subCmd.Execute()
					}
				}
			}
		}

		// If we are here, it means either:
		// 1. No `--` was used.
		// 2. `--` was used but no subcommand found (e.g. `et -- file.txt`).
		
		// In this case, we proceed with normal execution (rule matching).
		// We use remainingArgs as the arguments.
		
		args = remainingArgs
		
		if len(args) < 1 {
			return cmd.Help()
		}

		// Check for profile environment variable
		if profile == "" && os.Getenv("ENTRY_PROFILE") != "" {
			profile = os.Getenv("ENTRY_PROFILE")
		}

		// Resolve config file path with profile
		if cfgFile == "" && profile != "" {
			resolvedPath, err := config.GetConfigPathWithProfile("", profile)
			if err != nil {
				return fmt.Errorf("failed to resolve profile config path: %w", err)
			}
			cfgFile = resolvedPath
			logger.Debug("Using profile '%s' with config: %s", profile, cfgFile)
		}

		// Initialize logger
		if err := initLogger(); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		defer logger.GetGlobal().Close()

		logger.Debug("Starting entry with args: %v", args)
		logger.Debug("Flags - verbose: %v, dryRun: %v, interactive: %v, explain: %v, profile: %s", verbose, dryRun, interactive, explain, profile)

		// args[0] is the file/command
		// args[1:] are arguments to the command (if matched) or part of the command

		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		// Initialize Executor
		exec := executor.NewExecutor(cmd.OutOrStdout(), dryRun)

		// Explain mode: show detailed matching information
		// Only valid if we have exactly one argument (the file)
		if explain {
			if len(args) == 1 {
				return handleExplain(cmd, cfg, args[0])
			}
		}

		// Handle file execution with single argument
		if len(args) == 1 {
			filename := args[0]
			
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
		// Or if file execution failed
		return handleCommandExecution(cfg, exec, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}


