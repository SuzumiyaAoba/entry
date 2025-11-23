package cli

import (
	"bytes"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Move command", func() {
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
				{Name: "Rule 2", Command: "cmd2"},
				{Name: "Rule 3", Command: "cmd3"},
			},
		}
		err := config.SaveConfig(cfgFile, cfg)
		Expect(err).NotTo(HaveOccurred())

		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&outBuf)
	})

	It("should move rule up", func() {
		// Move Rule 2 (index 2) to index 1
		err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "move", "2", "1"})
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("Rule moved"))

		// Verify order: Rule 2, Rule 1, Rule 3
		cfg, err := config.LoadConfig(cfgFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Rules[0].Name).To(Equal("Rule 2"))
		Expect(cfg.Rules[1].Name).To(Equal("Rule 1"))
	})

	It("should move rule down", func() {
		// Move Rule 2 (index 2) to index 3
		err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "move", "2", "3"})
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("Rule moved"))

		// Verify order: Rule 1, Rule 3, Rule 2
		cfg, err := config.LoadConfig(cfgFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Rules[1].Name).To(Equal("Rule 3"))
		Expect(cfg.Rules[2].Name).To(Equal("Rule 2"))
	})

	It("should move rule to specific index", func() {
		// Move Rule 3 (index 3) to index 1
		err := rootCmd.RunE(rootCmd, []string{"--config", cfgFile, ":config", "move", "3", "1"})
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("Rule moved"))

		// Verify order: Rule 3, Rule 1, Rule 2
		cfg, err := config.LoadConfig(cfgFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Rules[0].Name).To(Equal("Rule 3"))
		Expect(cfg.Rules[1].Name).To(Equal("Rule 1"))
	})
})
