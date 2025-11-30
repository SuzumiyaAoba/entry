package cli

import (
	"fmt"

	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/history"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func init() {
	historyCmd.AddCommand(historyClearCmd)
}

var historyCmd = &cobra.Command{
	Use:   ":history",
	Short: "View and execute command history",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHistory(cmd)
	},
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear command history",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := history.ClearHistory(); err != nil {
			return fmt.Errorf("failed to clear history: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "History cleared")
		return nil
	},
}

// HistoryOption represents an option in the history selection menu
type HistoryOption struct {
	Label string
	Entry history.HistoryEntry
}

// showHistorySelector displays an interactive selector for history entries
func showHistorySelector(entries []history.HistoryEntry) (history.HistoryEntry, error) {
	var options []huh.Option[HistoryOption]
	for _, entry := range entries {
		label := fmt.Sprintf("%s  %s (%s)", 
			entry.Timestamp.Format("2006-01-02 15:04:05"), 
			entry.Command, 
			entry.RuleName)
		options = append(options, huh.NewOption(label, HistoryOption{Label: label, Entry: entry}))
	}

	var selected HistoryOption
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[HistoryOption]().
				Title("Select a command to re-run").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return history.HistoryEntry{}, err
	}

	return selected.Entry, nil
}

// HistorySelectorFunc is the function signature for selecting a history entry
type HistorySelectorFunc func(entries []history.HistoryEntry) (history.HistoryEntry, error)

// CurrentHistorySelector is the current selector function, can be swapped for testing
var CurrentHistorySelector HistorySelectorFunc = showHistorySelector

func runHistory(cmd *cobra.Command) error {
	entries, err := history.LoadHistory()
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No history available")
		return nil
	}

	selectedEntry, err := CurrentHistorySelector(entries)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Re-running: %s\n", selectedEntry.Command)
	
	exec := executor.NewExecutor(cmd.OutOrStdout(), false)
	return exec.ExecuteCommand("et", []string{selectedEntry.Command})
}
