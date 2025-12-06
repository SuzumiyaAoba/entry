package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/via/internal/config"
	"github.com/spf13/cobra"
)

var configExportCmd = &cobra.Command{
	Use:   "export <path>",
	Short: "Export configuration to a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigExport(cmd, args[0])
	},
}

var configImportCmd = &cobra.Command{
	Use:   "import <path>",
	Short: "Import configuration from a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigImport(cmd, args[0])
	},
}

func runConfigExport(cmd *cobra.Command, destPath string) error {
	configPath, err := config.GetConfigPath(cfgFile)
	if err != nil {
		return err
	}

	// Ensure absolute path for destination
	absDest, err := filepath.Abs(destPath)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Read source
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Write destination
	if err := os.WriteFile(absDest, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Configuration exported to %s\n", absDest)
	return nil
}

func runConfigImport(cmd *cobra.Command, srcPath string) error {
	configPath, err := config.GetConfigPath(cfgFile)
	if err != nil {
		return err
	}

	// Validate source file by trying to load it
	cfg, err := config.LoadConfig(srcPath)
	if err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	// Validate content
	if err := config.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration content: %w", err)
	}

	// Copy file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Configuration imported from %s\n", srcPath)
	return nil
}
