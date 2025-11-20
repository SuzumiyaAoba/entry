package cli

import (
	"fmt"
	"os"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage the entry configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigList()
	},
}

var configOpenCmd = &cobra.Command{
	Use:   "open",
	Short: "Open configuration file in editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigOpen()
	},
}

var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigAdd(cmd, args)
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configOpenCmd)
	configCmd.AddCommand(configAddCmd)

	configAddCmd.Flags().String("ext", "", "Extension to match (comma separated)")
	configAddCmd.Flags().String("cmd", "", "Command to execute")
	configAddCmd.MarkFlagRequired("cmd")

	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configRemoveCmd)
	configCmd.AddCommand(configSetDefaultCmd)
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigInit()
	},
}

var configCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check configuration validity",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigCheck()
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <index>",
	Short: "Remove a rule by index (1-based)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigRemove(args[0])
	},
}

var configSetDefaultCmd = &cobra.Command{
	Use:   "set-default <command>",
	Short: "Set default command",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSetDefault(args[0])
	},
}

func handleConfigCommand(args []string) error {
	// We need to manually dispatch or use configCmd with custom args
	// Since configCmd is attached to rootCmd, executing it might be tricky if we don't want full root execution.
	// But we can just use it for Help and dispatch manually, OR try to use it for execution.
	
	// Let's try to use SetArgs and Execute? 
	// But configCmd.Execute() calls the root command's Execute usually? No, it calls the command's Execute.
	// However, configCmd is a child.
	
	// Simplest way: Use manual dispatch but use Cobra for 'add' flag parsing if possible, 
	// or just keep manual dispatch and use configCmd ONLY for help registration in root.
	
	// User wants "help" to show config.
	// So configCmd MUST be added to rootCmd.
	
	// For execution:
	if len(args) == 0 {
		return configCmd.Help()
	}

	// Check for help flag in args to show config help
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return configCmd.Help()
		}
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "list":
		return runConfigList()
	case "open":
		return runConfigOpen()
	case "add":
		// Use Cobra's flag parsing for add?
		// We can use configAddCmd.ParseFlags(subargs)
		if err := configAddCmd.ParseFlags(subargs); err != nil {
			return err
		}
		// And then call RunE
		return runConfigAdd(configAddCmd, configAddCmd.Flags().Args())
	case "init":
		return runConfigInit()
	case "check":
		return runConfigCheck()
	case "remove":
		if len(subargs) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(subargs))
		}
		return runConfigRemove(subargs[0])
	case "set-default":
		if len(subargs) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(subargs))
		}
		return runConfigSetDefault(subargs[0])
	default:
		return fmt.Errorf("unknown config subcommand: %s", subcmd)
	}
}

func runConfigList() error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func runConfigOpen() error {
	configPath, err := config.GetConfigPath(cfgFile)
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := &config.Config{Version: "1"}
		if err := config.SaveConfig(cfgFile, cfg); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
		fmt.Printf("Created default config at %s\n", configPath)
	}

	exec := executor.NewExecutor(os.Stdout, false)
	return exec.OpenSystem(configPath)
}

func runConfigAdd(cmd *cobra.Command, args []string) error {
	ext, _ := cmd.Flags().GetString("ext")
	command, _ := cmd.Flags().GetString("cmd")

	if command == "" {
		return fmt.Errorf("--cmd is required")
	}

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		cfg = &config.Config{Version: "1"}
	}

	rule := config.Rule{
		Command: command,
	}

	if ext != "" {
		rule.Extensions = []string{ext}
	}

	cfg.Rules = append(cfg.Rules, rule)

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return err
	}

	fmt.Println("Rule added successfully")
	return nil
}

func runConfigInit() error {
	configPath, err := config.GetConfigPath(cfgFile)
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	cfg := &config.Config{
		Version: "1",
		DefaultCommand: "vim {{.File}}",
		Rules: []config.Rule{
			{
				Name: "Example Rule",
				Extensions: []string{"txt", "md"},
				Command: "cat {{.File}}",
				Terminal: true,
			},
		},
	}

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return fmt.Errorf("failed to create default config: %w", err)
	}

	fmt.Printf("Created default config at %s\n", configPath)
	return nil
}

func runConfigCheck() error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if err := config.ValidateConfig(cfg); err != nil {
		return err
	}

	fmt.Println("Configuration is valid")
	return nil
}

func runConfigRemove(indexStr string) error {
	var index int
	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		return fmt.Errorf("invalid index: %s", indexStr)
	}

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if index < 1 || index > len(cfg.Rules) {
		return fmt.Errorf("index out of range: %d", index)
	}

	// Remove rule (1-based index)
	cfg.Rules = append(cfg.Rules[:index-1], cfg.Rules[index:]...)

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return err
	}

	fmt.Println("Rule removed successfully")
	return nil
}

func runConfigSetDefault(command string) error {
	// Check if config exists first
	configPath, err := config.GetConfigPath(cfgFile)
	if err != nil {
		return err
	}

	var cfg *config.Config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg = &config.Config{Version: "1"}
	} else {
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			return err
		}
	}

	cfg.DefaultCommand = command
	// Clear alias if present to avoid confusion
	cfg.Default = ""

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return err
	}

	fmt.Println("Default command updated successfully")
	return nil
}
