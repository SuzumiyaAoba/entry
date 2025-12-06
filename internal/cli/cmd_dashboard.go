package cli

import (
	"fmt"

	"github.com/SuzumiyaAoba/via/internal/config"
	"github.com/SuzumiyaAoba/via/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   ":dashboard",
	Short: "Open TUI dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDashboard(cmd)
	},
}

// runTeaProgram runs the tea program, can be swapped for testing
var runTeaProgram = func(model tea.Model, opts ...tea.ProgramOption) (tea.Model, error) {
	fmt.Println("Real runTeaProgram called")
	p := tea.NewProgram(model, opts...)
	return p.Run()
}

func runDashboard(cmd *cobra.Command) error {
	configPath, err := config.GetConfigPath(cfgFile)
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		// Only ignore if file does not exist, but for now let's return error
		// because swallowing parsing errors is bad.
		// If the user wants to start fresh, they should ensure no broken config exists.
		return fmt.Errorf("failed to load config: %w", err)
	}

	model, err := tui.NewModel(cfg, configPath)
	if err != nil {
		return err
	}

	if _, err := runTeaProgram(model, tea.WithAltScreen()); err != nil {
		return fmt.Errorf("error running dashboard: %w", err)
	}

	return nil
}
