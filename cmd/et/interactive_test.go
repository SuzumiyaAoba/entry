package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/executor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Interactive mode", func() {
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
					Name:       "Text Editor",
					Extensions: []string{"txt"},
					Command:    "vim {{.File}}",
				},
				{
					Name:       "Text Viewer",
					Extensions: []string{"txt"},
					Command:    "cat {{.File}}",
				},
			},
		}
	})

	Describe("executeSelectedOption", func() {
		It("should execute rule option", func() {
			option := Option{
				Label: "Text Editor",
				Rule: &config.Rule{
					Command: "vim {{.File}}",
				},
				IsSystem: false,
			}

			err := executeSelectedOption(cfg, exec, option, "test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("vim test.txt"))
		})

		It("should execute system default option", func() {
			testFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			option := Option{
				Label:    "System Default",
				IsSystem: true,
			}

			err = executeSelectedOption(cfg, exec, option, testFile)
			Expect(err).NotTo(HaveOccurred())
			// System opener will be called
			Expect(outBuf.String()).To(ContainSubstring(testFile))
		})

		It("should use default command when configured", func() {
			cfg.DefaultCommand = "nano {{.File}}"

			option := Option{
				Label:    "System Default",
				IsSystem: true,
			}

			err := executeSelectedOption(cfg, exec, option, "test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("nano test.txt"))
		})
	})
})
