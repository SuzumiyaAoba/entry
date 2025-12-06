package cli

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/via/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Export/Import commands", func() {
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

	Describe("runConfigExport", func() {
		It("should export config to valid path", func() {
			destPath := filepath.Join(tmpDir, "exported.yml")
			err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "export", destPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Configuration exported"))
			Expect(fileExists(destPath)).To(BeTrue())
		})

		It("should fail if destination directory does not exist", func() {
			destPath := filepath.Join(tmpDir, "nonexistent", "exported.yml")
			err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "export", destPath})
			Expect(err).To(HaveOccurred())
		})

		It("should fail if config file read fails", func() {
			// Make config unreadable
			err := os.Chmod(cfgFile, 0000)
			Expect(err).NotTo(HaveOccurred())
			defer os.Chmod(cfgFile, 0644)

			destPath := filepath.Join(tmpDir, "exported.yml")
			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "export", destPath})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to read config file"))
		})
	})

	Describe("runConfigImport", func() {
		It("should import valid config", func() {
			// Create a valid config to import
			srcPath := filepath.Join(tmpDir, "import_source.yml")
			err := os.WriteFile(srcPath, []byte("version: '1'\nrules: []"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "import", srcPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Configuration imported"))
		})

		It("should fail if source file does not exist", func() {
			srcPath := filepath.Join(tmpDir, "nonexistent.yml")
			err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "import", srcPath})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid config file"))
		})

		It("should fail if config content is invalid", func() {
			srcPath := filepath.Join(tmpDir, "invalid.yml")
			err := os.WriteFile(srcPath, []byte("invalid yaml content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "import", srcPath})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid config file"))
		})

		It("should fail if imported config fails validation", func() {
			srcPath := filepath.Join(tmpDir, "invalid_rules.yml")
			// Valid YAML but invalid rule (missing command)
			content := `
version: "1"
rules:
  - name: Invalid Rule
`
			err := os.WriteFile(srcPath, []byte(content), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "import", srcPath})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid configuration content"))
		})
	})
})
