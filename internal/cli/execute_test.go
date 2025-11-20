package cli

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Execute handlers", func() {
	var (
		tmpDir string
		cfg    *config.Config
		exec   *executor.Executor
		outBuf bytes.Buffer
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		outBuf.Reset()
		exec = executor.NewExecutor(&outBuf, true) // dry-run mode
		cfg = &config.Config{
			Rules: []config.Rule{
				{
					Extensions: []string{"txt"},
					Command:    "cat {{.File}}",
				},
				{
					Extensions:  []string{"md"},
					Command:     "vim {{.File}}",
					Fallthrough: true,
				},
				{
					Extensions: []string{"md"},
					Command:    "echo {{.File}}",
				},
			},
		}
	})

	Describe("handleFileExecution", func() {
		It("should execute matched rule", func() {
			testFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = handleFileExecution(cfg, exec, testFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("cat"))
		})

		It("should execute fallthrough rules", func() {
			testFile := filepath.Join(tmpDir, "test.md")
			err := os.WriteFile(testFile, []byte("# Markdown"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = handleFileExecution(cfg, exec, testFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("vim"))
			Expect(outBuf.String()).To(ContainSubstring("echo"))
		})

		It("should use default command when no rule matches", func() {
			cfg.DefaultCommand = "xdg-open {{.File}}"
			testFile := filepath.Join(tmpDir, "test.pdf")
			err := os.WriteFile(testFile, []byte("PDF content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = handleFileExecution(cfg, exec, testFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("xdg-open"))
		})

		It("should handle URL with matching rule", func() {
			cfg.Rules = []config.Rule{
				{
					Scheme:  "https",
					Command: "curl {{.File}}",
				},
			}

			err := handleFileExecution(cfg, exec, "https://example.com")
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("curl"))
		})

		It("should return error for non-existing file without default", func() {
			err := handleFileExecution(cfg, exec, "nonexistent.xyz")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("handleCommandExecution", func() {
		It("should execute alias", func() {
			cfg.Aliases = map[string]string{
				"ll": "ls -la",
			}

			err := handleCommandExecution(cfg, exec, []string{"ll", "/tmp"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("ls -la /tmp"))
		})

		It("should execute command without alias", func() {
			err := handleCommandExecution(cfg, exec, []string{"echo", "hello"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("echo hello"))
		})

		It("should use default command for single unknown argument", func() {
			cfg.DefaultCommand = "vim {{.File}}"

			err := handleCommandExecution(cfg, exec, []string{"newfile.txt"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("vim newfile.txt"))
		})

		It("should execute as command with multiple args", func() {
			// Using echo which should be available on all systems
			err := handleCommandExecution(cfg, exec, []string{"echo", "hello", "world"})
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("echo hello world"))
		})
	})
})
