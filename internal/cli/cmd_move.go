package cli

import (
	"fmt"
	"strconv"

	"github.com/SuzumiyaAoba/via/internal/config"
	"github.com/spf13/cobra"
)

var configMoveCmd = &cobra.Command{
	Use:   "move <from_index> <to_index>",
	Short: "Move a rule to a new position",
	Long:  `Move a rule from one position to another. Indices are 1-based.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		from, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid from_index: %s", args[0])
		}
		to, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid to_index: %s", args[1])
		}
		return runConfigMove(cmd, from, to)
	},
}

func runConfigMove(cmd *cobra.Command, from, to int) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if from < 1 || from > len(cfg.Rules) {
		return fmt.Errorf("from_index out of range: %d", from)
	}
	if to < 1 || to > len(cfg.Rules) {
		return fmt.Errorf("to_index out of range: %d", to)
	}

	if from == to {
		fmt.Fprintln(cmd.OutOrStdout(), "Source and destination are the same")
		return nil
	}

	// Adjust to 0-based index
	fromIdx := from - 1
	toIdx := to - 1

	// Move the element
	rule := cfg.Rules[fromIdx]
	
	// Remove from source
	cfg.Rules = append(cfg.Rules[:fromIdx], cfg.Rules[fromIdx+1:]...)
	
	// Insert at destination
	// If toIdx is now greater than length (because we removed one), it means append
	if toIdx >= len(cfg.Rules) {
		cfg.Rules = append(cfg.Rules, rule)
	} else {
		cfg.Rules = append(cfg.Rules[:toIdx], append([]config.Rule{rule}, cfg.Rules[toIdx:]...)...)
	}

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Rule moved from %d to %d\n", from, to)
	return nil
}
