package cli

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
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
		
		cfg := &config.Config{
			Version: "1",
			Rules: []config.Rule{
				{Name: "Rule 1", Command: "cmd1"},
			},
		}
		err := config.SaveConfig(cfgFile, cfg)
		Expect(err).NotTo(HaveOccurred())

		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&outBuf)
	})

	It("should export config to file", func() {
		exportFile := filepath.Join(tmpDir, "export.yml")
		err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "export", exportFile})
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("Configuration exported"))

		// Verify file content
		content, err := os.ReadFile(exportFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("Rule 1"))
	})

	It("should import config from file", func() {
		// Create import file
		importFile := filepath.Join(tmpDir, "import.yml")
		cfg := &config.Config{
			Version: "1",
			Rules: []config.Rule{
				{Name: "Imported Rule", Command: "cmd2"},
			},
		}
		err := config.SaveConfig(importFile, cfg)
		Expect(err).NotTo(HaveOccurred())

		err = rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "import", importFile})
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("Configuration imported"))

		// Verify config updated
		loadedCfg, err := config.LoadConfig(cfgFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(loadedCfg.Rules[0].Name).To(Equal("Imported Rule"))
	})
})
