package cli

import (
	"bytes"
	"path/filepath"

	"github.com/SuzumiyaAoba/via/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Match command", func() {
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

	Describe("runMatch", func() {
		BeforeEach(func() {
			cfg := config.Config{
				Version: "1",
				Rules: []config.Rule{
					{Name: "Text Rule", Extensions: []string{"txt"}, Command: "cat"},
					{Command: "echo", Extensions: []string{"log"}},
				},
			}
			err := config.SaveConfig(configFile, &cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should print matched rule name", func() {
			err := runMatch(rootCmd, "file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("Text Rule"))
		})

		It("should print matched rule command if name missing", func() {
			err := runMatch(rootCmd, "file.log")
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("echo"))
		})

		It("should return error if no match found", func() {
			err := runMatch(rootCmd, "file.pdf")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no match found"))
		})

		It("should return error if config load fails", func() {
			cfgFile = filepath.Join(tmpDir, "nonexistent.yml")
			err := runMatch(rootCmd, "file.txt")
			Expect(err).To(HaveOccurred())
		})
	})
})
