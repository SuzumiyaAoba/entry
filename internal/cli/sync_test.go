package cli

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sync command", func() {
	var (
		tmpDir  string
		outBuf  bytes.Buffer
	)

	BeforeEach(func() {
		resetGlobals()
		tmpDir = GinkgoT().TempDir()
		cfgFile = filepath.Join(tmpDir, "config.yml")
		
		cfg := &config.Config{Version: "1"}
		err := config.SaveConfig(cfgFile, cfg)
		Expect(err).NotTo(HaveOccurred())

		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&outBuf)
	})

	Describe("runConfigSyncPush", func() {
		It("should fail if sync not initialized", func() {
			err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "sync", "push"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("sync not initialized"))
		})

		It("should fail if token is missing", func() {
			// Initialize sync without token
			cfg := &config.Config{
				Version: "1",
				Sync: &config.SyncConfig{
					GistID: "123",
				},
			}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())
			
			// Ensure env var is unset
			os.Unsetenv("ENTRY_GITHUB_TOKEN")

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "sync", "push"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("token not found"))
		})
	})

	Describe("runConfigSyncPull", func() {
		It("should fail pull if sync not initialized", func() {
			// Create config without sync
			cfg := &config.Config{Version: "1"}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "sync", "pull"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("sync not initialized"))
		})

		It("should fail push if token missing", func() {
			cfg := &config.Config{
				Version: "1",
				Sync: &config.SyncConfig{
					GistID: "gist123",
				},
			}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			// Ensure env var is unset
			os.Unsetenv("ENTRY_GITHUB_TOKEN")

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "sync", "push"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("token not found"))
		})

		It("should fail pull if gist fetch fails", func() {
			cfg := &config.Config{
				Version: "1",
				Sync: &config.SyncConfig{
					GistID: "nonexistent",
					Token:  "token",
				},
			}
			err := config.SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "sync", "pull"})
			Expect(err).To(HaveOccurred())
		})
	})
	Describe("runConfigSyncInit", func() {
		var (
			originalInputProvider func() (SyncInitInput, error)
			originalClientFactory func(token string) SyncClient
		)

		BeforeEach(func() {
			originalInputProvider = getSyncInitInput
			originalClientFactory = newSyncClient
		})

		AfterEach(func() {
			getSyncInitInput = originalInputProvider
			newSyncClient = originalClientFactory
		})

		It("should fail if config load fails", func() {
			// Corrupt config file
			err := os.WriteFile(cfgFile, []byte("["), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = runConfigSyncInit(rootCmd)
			Expect(err).To(HaveOccurred())
		})

		It("should initialize with existing Gist", func() {
			// Mock input
			getSyncInitInput = func() (SyncInitInput, error) {
				return SyncInitInput{
					CreateNew:  false,
					GistID:     "existing-gist-id",
					Token:      "test-token",
					StoreToken: true,
				}, nil
			}

			// Mock client
			newSyncClient = func(token string) SyncClient {
				return &mockSyncClient{}
			}

			err := runConfigSyncInit(rootCmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Sync initialized successfully"))

			// Verify config
			cfg, err := config.LoadConfig(cfgFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Sync.GistID).To(Equal("existing-gist-id"))
			Expect(cfg.Sync.Token).To(Equal("test-token"))
		})

		It("should initialize by creating new Gist", func() {
			// Mock input
			getSyncInitInput = func() (SyncInitInput, error) {
				return SyncInitInput{
					CreateNew:  true,
					Token:      "test-token",
					StoreToken: false,
				}, nil
			}

			// Mock client
			newSyncClient = func(token string) SyncClient {
				return &mockSyncClient{
					createGistFunc: func(cfg *config.Config, public bool) (string, error) {
						return "new-gist-id", nil
					},
				}
			}

			err := runConfigSyncInit(rootCmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Created new Gist: new-gist-id"))
			Expect(outBuf.String()).To(ContainSubstring("Token not stored"))

			// Verify config
			cfg, err := config.LoadConfig(cfgFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Sync.GistID).To(Equal("new-gist-id"))
			Expect(cfg.Sync.Token).To(BeEmpty())
		})
	})
})

type mockSyncClient struct {
	createGistFunc func(cfg *config.Config, public bool) (string, error)
	updateGistFunc func(gistID string, cfg *config.Config) error
	getGistFunc    func(gistID string) (*config.Config, error)
}

func (m *mockSyncClient) CreateGist(cfg *config.Config, public bool) (string, error) {
	if m.createGistFunc != nil {
		return m.createGistFunc(cfg, public)
	}
	return "mock-gist-id", nil
}

func (m *mockSyncClient) UpdateGist(gistID string, cfg *config.Config) error {
	if m.updateGistFunc != nil {
		return m.updateGistFunc(gistID, cfg)
	}
	return nil
}

func (m *mockSyncClient) GetGist(gistID string) (*config.Config, error) {
	if m.getGistFunc != nil {
		return m.getGistFunc(gistID)
	}
	return &config.Config{}, nil
}
