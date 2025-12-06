package cli

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/via/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Explain mode", func() {
	var (
		tmpDir string
		cfg    *config.Config
		outBuf bytes.Buffer
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		outBuf.Reset()
		rootCmd.SetOut(&outBuf)

		cfg = &config.Config{
			Rules: []config.Rule{
				{
					Name:       "Text Editor",
					Extensions: []string{"txt"},
					Command:    "vim {{.File}}",
				},
				{
					Name:       "Markdown Editor",
					Extensions: []string{"md"},
					Regex:      `.*\.md$`,
					Command:    "vim {{.File}}",
				},
				{
					Name:    "Web Browser",
					Scheme:  "https",
					Command: "open {{.File}}",
				},
			},
		}
	})

	Describe("handleExplain", func() {
		It("should explain matching for existing file", func() {
			testFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = handleExplain(rootCmd, cfg, testFile)
			Expect(err).NotTo(HaveOccurred())

			output := outBuf.String()
			Expect(output).To(ContainSubstring("EXPLAIN MODE"))
			Expect(output).To(ContainSubstring("FILE INFORMATION"))
			Expect(output).To(ContainSubstring("RULE EVALUATION"))
			Expect(output).To(ContainSubstring("Text Editor"))
		})

		It("should explain matching for URL", func() {
			err := handleExplain(rootCmd, cfg, "https://example.com")
			Expect(err).NotTo(HaveOccurred())

			output := outBuf.String()
			Expect(output).To(ContainSubstring("EXPLAIN MODE"))
			Expect(output).To(ContainSubstring("URL"))
			Expect(output).To(ContainSubstring("https"))
		})

		It("should explain when no rules match", func() {
			testFile := filepath.Join(tmpDir, "test.pdf")
			err := os.WriteFile(testFile, []byte("PDF content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = handleExplain(rootCmd, cfg, testFile)
			Expect(err).NotTo(HaveOccurred())

			output := outBuf.String()
			Expect(output).To(ContainSubstring("No rules matched"))
		})

		It("should show default command when configured", func() {
			cfg.DefaultCommand = "xdg-open {{.File}}"
			testFile := filepath.Join(tmpDir, "test.pdf")
			err := os.WriteFile(testFile, []byte("PDF content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = handleExplain(rootCmd, cfg, testFile)
			Expect(err).NotTo(HaveOccurred())

			output := outBuf.String()
			Expect(output).To(ContainSubstring("default command"))
		})

		It("should explain for non-existing file", func() {
			err := handleExplain(rootCmd, cfg, "nonexistent.txt")
			Expect(err).NotTo(HaveOccurred())

			output := outBuf.String()
			Expect(output).To(ContainSubstring("Does not exist"))
		})

		It("should show regex matching", func() {
			testFile := filepath.Join(tmpDir, "test.md")
			err := os.WriteFile(testFile, []byte("# Markdown"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = handleExplain(rootCmd, cfg, testFile)
			Expect(err).NotTo(HaveOccurred())

			output := outBuf.String()
			Expect(output).To(ContainSubstring("Markdown Editor"))
			Expect(output).To(ContainSubstring("Ext:"))
		})

		It("should show MIME matching", func() {
			// Add MIME rule
			cfg.Rules = append(cfg.Rules, config.Rule{
				Name: "PDF Viewer",
				Mime: "application/pdf",
				Command: "open {{.File}}",
			})
			
			// Create fake PDF
			testFile := filepath.Join(tmpDir, "test.pdf")
			err := os.WriteFile(testFile, []byte("%PDF-1.4"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = handleExplain(rootCmd, cfg, testFile)
			Expect(err).NotTo(HaveOccurred())

			output := outBuf.String()
			Expect(output).To(ContainSubstring("PDF Viewer"))
			Expect(output).To(ContainSubstring("MIME"))
		})

		It("should show fallthrough", func() {
			// Add fallthrough rule
			cfg.Rules = append([]config.Rule{
				{
					Name: "Logger",
					Regex: ".*",
					Command: "log {{.File}}",
					Fallthrough: true,
				},
			}, cfg.Rules...)

			testFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = handleExplain(rootCmd, cfg, testFile)
			Expect(err).NotTo(HaveOccurred())

			output := outBuf.String()
			Expect(output).To(ContainSubstring("Logger"))
			Expect(output).To(ContainSubstring("Text Editor"))
			Expect(output).To(ContainSubstring("→"))
		})
	})
})
