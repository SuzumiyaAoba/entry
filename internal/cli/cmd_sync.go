package cli

import (
	"fmt"
	"os"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/sync"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var configSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync configuration with Gist",
}

var configSyncInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize sync configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSyncInit(cmd)
	},
}

var configSyncPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push configuration to Gist",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSyncPush(cmd)
	},
}

var configSyncPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull configuration from Gist",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSyncPull(cmd)
	},
}

func init() {
	configSyncCmd.AddCommand(configSyncInitCmd)
	configSyncCmd.AddCommand(configSyncPushCmd)
	configSyncCmd.AddCommand(configSyncPullCmd)
}

// SyncInitInput holds the input gathered from the user
type SyncInitInput struct {
	CreateNew  bool
	GistID     string
	Token      string
	StoreToken bool
}

// getSyncInitInput gathers input from the user for sync initialization
var getSyncInitInput = func() (SyncInitInput, error) {
	var input SyncInitInput

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Create new Gist?").
				Description("If no, you must provide an existing Gist ID").
				Value(&input.CreateNew),
		),
	)

	if err := form.Run(); err != nil {
		return input, err
	}

	if !input.CreateNew {
		form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Gist ID").
					Value(&input.GistID),
			),
		)
		if err := form.Run(); err != nil {
			return input, err
		}
	}

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("GitHub Token").
				Description("Required for private Gists or creating new ones").
				EchoMode(huh.EchoModePassword).
				Value(&input.Token),
		),
	)

	if err := form.Run(); err != nil {
		return input, err
	}

	confirm := huh.NewConfirm().
		Title("Store token in config?").
		Description("WARNING: This is insecure as the config file is plain text.").
		Value(&input.StoreToken)
	
	if err := confirm.Run(); err != nil {
		return input, err
	}

	return input, nil
}

// SyncClient interface for mocking
type SyncClient interface {
	CreateGist(cfg *config.Config, public bool) (string, error)
	UpdateGist(gistID string, cfg *config.Config) error
	GetGist(gistID string) (*config.Config, error)
}

// newSyncClient creates a new sync client
var newSyncClient = func(token string) SyncClient {
	return sync.NewClient(token)
}

func runConfigSyncInit(cmd *cobra.Command) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	input, err := getSyncInitInput()
	if err != nil {
		return err
	}

	client := newSyncClient(input.Token)
	gistID := input.GistID

	if input.CreateNew {
		id, err := client.CreateGist(cfg, false) // Default to private
		if err != nil {
			return err
		}
		gistID = id
		fmt.Fprintf(cmd.OutOrStdout(), "Created new Gist: %s\n", gistID)
	}

	if cfg.Sync == nil {
		cfg.Sync = &config.SyncConfig{}
	}
	cfg.Sync.GistID = gistID

	if input.StoreToken {
		cfg.Sync.Token = input.Token
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Token not stored. You will need to provide it via ENTRY_GITHUB_TOKEN env var.")
	}

	if err := config.SaveConfig(cfgFile, cfg); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Sync initialized successfully")
	return nil
}

func runConfigSyncPush(cmd *cobra.Command) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if cfg.Sync == nil || cfg.Sync.GistID == "" {
		return fmt.Errorf("sync not initialized. Run 'et :config sync init' first")
	}

	token := cfg.Sync.Token
	if token == "" {
		token = os.Getenv("ENTRY_GITHUB_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("token not found. Set ENTRY_GITHUB_TOKEN or store it in config")
	}

	client := sync.NewClient(token)
	if err := client.UpdateGist(cfg.Sync.GistID, cfg); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Configuration pushed to Gist")
	return nil
}

func runConfigSyncPull(cmd *cobra.Command) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if cfg.Sync == nil || cfg.Sync.GistID == "" {
		return fmt.Errorf("sync not initialized. Run 'et :config sync init' first")
	}

	token := cfg.Sync.Token
	if token == "" {
		token = os.Getenv("ENTRY_GITHUB_TOKEN")
	}
	// Token might not be needed for public gists, but usually good to have.
	
	client := sync.NewClient(token)
	newCfg, err := client.GetGist(cfg.Sync.GistID)
	if err != nil {
		return err
	}

	// Preserve local sync settings
	newCfg.Sync = cfg.Sync

	if err := config.SaveConfig(cfgFile, newCfg); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Configuration pulled from Gist")
	return nil
}
