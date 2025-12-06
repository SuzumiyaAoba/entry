package cli

import (
	"bytes"
	"path/filepath"

	"github.com/SuzumiyaAoba/via/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Move command", func() {
	var (
		tmpDir     string
		configFile string
		outBuf     bytes.Buffer
	)

	BeforeEach(func() {
		resetGlobals()
		tmpDir = GinkgoT().TempDir()
		configFile = filepath.Join(tmpDir, "config.yml")
		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&outBuf)
		cfgFile = configFile
	})

	AfterEach(func() {
		cfgFile = ""
	})

	Describe("runConfigMove", func() {
		BeforeEach(func() {
			cfg := config.Config{
				Version: "1",
				Rules: []config.Rule{
					{Name: "Rule 1", Command: "cmd1"},
					{Name: "Rule 2", Command: "cmd2"},
					{Name: "Rule 3", Command: "cmd3"},
				},
			}
			err := config.SaveConfig(configFile, &cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should move rule forward", func() {
			err := runConfigMove(rootCmd, 1, 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Rule moved from 1 to 3"))

			cfg, err := config.LoadConfig(configFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Rules[0].Name).To(Equal("Rule 2"))
			Expect(cfg.Rules[1].Name).To(Equal("Rule 3"))
			Expect(cfg.Rules[2].Name).To(Equal("Rule 1"))
		})

		It("should move rule backward", func() {
			err := runConfigMove(rootCmd, 3, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Rule moved from 3 to 1"))

			cfg, err := config.LoadConfig(configFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Rules[0].Name).To(Equal("Rule 3"))
			Expect(cfg.Rules[1].Name).To(Equal("Rule 1"))
			Expect(cfg.Rules[2].Name).To(Equal("Rule 2"))
		})

		It("should do nothing if from == to", func() {
			err := runConfigMove(rootCmd, 2, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Source and destination are the same"))
		})

		It("should fail if from index out of range", func() {
			err := runConfigMove(rootCmd, 0, 1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("from_index out of range"))

			err = runConfigMove(rootCmd, 4, 1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("from_index out of range"))
		})

		It("should fail if to index out of range", func() {
			err := runConfigMove(rootCmd, 1, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("to_index out of range"))

			err = runConfigMove(rootCmd, 1, 4) // 4 is valid (append) if length is 3? 
			// Logic: toIdx > len(cfg.Rules) check.
			// Current logic: if to < 1 || to > len(cfg.Rules) -> error.
			// So appending to end+1 is not allowed by validation, but logic supports it?
			// Let's check code: `if to < 1 || to > len(cfg.Rules)`
			// So max index is len.
			
			err = runConfigMove(rootCmd, 1, 5)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("to_index out of range"))
		})
	})
})
