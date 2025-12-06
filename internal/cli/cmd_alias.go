package cli

import (
	"fmt"
	"sort"

	"github.com/SuzumiyaAoba/via/internal/config"
	"github.com/spf13/cobra"
)

var configAliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage aliases",
	Long:  `Manage command aliases.`,
}

var configAliasAddCmd = &cobra.Command{
	Use:   "add <name> <command>",
	Short: "Add a new alias",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigAliasAdd(cmd, args[0], args[1])
	},
}

var configAliasRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an alias",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigAliasRemove(cmd, args[0])
	},
}

var configAliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all aliases",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigAliasList(cmd)
	},
}

func init() {
	configAliasCmd.AddCommand(configAliasListCmd)
	configAliasCmd.AddCommand(configAliasAddCmd)
	configAliasCmd.AddCommand(configAliasRemoveCmd)
}

func runConfigAliasAdd(cmd *cobra.Command, name, command string) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		// If config doesn't exist, create a new one
		cfg = &config.Config{Version: "1"}
	}

	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
	}

	if _, exists := cfg.Aliases[name]; exists {
		return fmt.Errorf("alias '%s' already exists", name)
	}

	cfg.Aliases[name] = command

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Alias '%s' added successfully\n", name)
	return nil
}

func runConfigAliasRemove(cmd *cobra.Command, name string) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if cfg.Aliases == nil {
		return fmt.Errorf("alias '%s' not found", name)
	}

	if _, exists := cfg.Aliases[name]; !exists {
		return fmt.Errorf("alias '%s' not found", name)
	}

	delete(cfg.Aliases, name)

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Alias '%s' removed successfully\n", name)
	return nil
}

func runConfigAliasList(cmd *cobra.Command) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if len(cfg.Aliases) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No aliases defined")
		return nil
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(cfg.Aliases))
	for k := range cfg.Aliases {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Fprintln(cmd.OutOrStdout(), "Aliases:")
	for _, k := range keys {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n", k, cfg.Aliases[k])
	}

	return nil
}
