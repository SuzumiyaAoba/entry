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
		cfgFile string
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
		It("should fail if sync not initialized", func() {
			err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "sync", "pull"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("sync not initialized"))
		})
	})
})
