package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/via/internal/config"
	"github.com/SuzumiyaAoba/via/internal/executor"
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

var _ = Describe("Interactive helpers", func() {
	var (
		tmpDir string
		cfg    *config.Config
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
	})

	Describe("buildOptionLabel", func() {
		It("should use name when available", func() {
			rule := &config.Rule{
				Name:    "My Editor",
				Command: "vim {{.File}}",
			}
			label := buildOptionLabel(rule)
			Expect(label).To(Equal("My Editor"))
		})

		It("should use command when name is empty", func() {
			rule := &config.Rule{
				Command: "cat {{.File}}",
			}
			label := buildOptionLabel(rule)
			Expect(label).To(Equal("Command: cat {{.File}}"))
		})
	})

	Describe("buildInteractiveOptions", func() {
		BeforeEach(func() {
			cfg = &config.Config{
				Rules: []config.Rule{
					{
						Name:       "Editor",
						Extensions: []string{"txt"},
						Command:    "vim {{.File}}",
					},
					{
						Extensions: []string{"txt"},
						Command:    "cat {{.File}}",
					},
				},
			}
		})

		It("should build options from matching rules", func() {
			testFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			options, err := buildInteractiveOptions(cfg, testFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(options).To(HaveLen(3)) // 2 rules + system default
			Expect(options[0].Label).To(Equal("Editor"))
			Expect(options[1].Label).To(ContainSubstring("Command:"))
			Expect(options[2].IsSystem).To(BeTrue())
		})

		It("should include system default for URLs", func() {
			options, err := buildInteractiveOptions(cfg, "https://example.com")
			Expect(err).NotTo(HaveOccurred())
			// URL won't match .txt extension rules
			// But system default should be included
			found := false
			for _, opt := range options {
				if opt.IsSystem {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should not include system default for non-existing file", func() {
			options, err := buildInteractiveOptions(cfg, "nonexistent.xyz")
			Expect(err).NotTo(HaveOccurred())
			found := false
			for _, opt := range options {
				if opt.IsSystem {
					found = true
					break
				}
			}
			Expect(found).To(BeFalse())
		})
	})

	Describe("handleInteractive", func() {
		var (
			testFile string
			cfg      *config.Config
			exec     *executor.Executor
			outBuf   bytes.Buffer
		)

		BeforeEach(func() {
			testFile = filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			outBuf.Reset()
			exec = executor.NewExecutor(&outBuf, true)
			cfg = &config.Config{
				Rules: []config.Rule{
					{
						Name:       "Text Editor",
						Extensions: []string{"txt"},
						Command:    "vim {{.File}}",
					},
				},
			}
		})

		AfterEach(func() {
			// Restore default selector
			CurrentSelector = showOptionSelector
		})

		It("should execute selected option", func() {
			// Mock selector
			CurrentSelector = func(options []Option, filename string) (Option, error) {
				// Find the "Text Editor" option
				for _, opt := range options {
					if opt.Label == "Text Editor" {
						return opt, nil
					}
				}
				return Option{}, fmt.Errorf("option not found")
			}

			err := handleInteractive(cfg, exec, testFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("vim " + testFile))
		})

		It("should return error if no matching rules", func() {
			// Do NOT create the file, so System Default is not added
			noMatchFile := filepath.Join(tmpDir, "nomatch.xyz")
			
			err := handleInteractive(cfg, exec, noMatchFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no matching rules found"))
		})

		It("should return error if selector fails", func() {
			CurrentSelector = func(options []Option, filename string) (Option, error) {
				return Option{}, fmt.Errorf("selection aborted")
			}

			err := handleInteractive(cfg, exec, testFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("selection aborted"))
		})
	})
})
