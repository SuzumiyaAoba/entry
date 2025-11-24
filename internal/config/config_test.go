package config

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {
	var (
		tmpDir  string
		cfgFile string
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		cfgFile = filepath.Join(tmpDir, "config.yml")
	})

	Describe("LoadConfig", func() {
		It("should load valid config", func() {
			err := os.WriteFile(cfgFile, []byte("version: '1'\nrules: []"), 0644)
			Expect(err).NotTo(HaveOccurred())

			cfg, err := LoadConfig(cfgFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Version).To(Equal("1"))
		})

		It("should return error if file does not exist", func() {
			_, err := LoadConfig(filepath.Join(tmpDir, "nonexistent.yml"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should return error if yaml is invalid", func() {
			err := os.WriteFile(cfgFile, []byte("invalid: yaml: content: :"), 0644)
			Expect(err).NotTo(HaveOccurred())

			_, err = LoadConfig(cfgFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse config"))
		})

		It("should use default alias as default command", func() {
			content := `
version: "1"
default: "vim {{.File}}"
`
			err := os.WriteFile(cfgFile, []byte(content), 0644)
			Expect(err).NotTo(HaveOccurred())

			cfg, err := LoadConfig(cfgFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.DefaultCommand).To(Equal("vim {{.File}}"))
		})
	})

	Describe("ValidateConfig", func() {
		It("should pass for valid config", func() {
			cfg := &Config{
				Version: "1",
				Rules: []Rule{
					{Name: "Rule", Command: "cmd"},
				},
			}
			err := ValidateConfig(cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail if command is missing", func() {
			cfg := &Config{
				Version: "1",
				Rules: []Rule{
					{Name: "Rule"}, // Missing Command
				},
			}
			err := ValidateConfig(cfg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Command"))
			Expect(err.Error()).To(ContainSubstring("required"))
		})

		It("should fail if regex is invalid", func() {
			cfg := &Config{
				Version: "1",
				Rules: []Rule{
					{Name: "Rule", Command: "cmd", Regex: "["},
				},
			}
			err := ValidateConfig(cfg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Regex"))
			Expect(err.Error()).To(ContainSubstring("is-regex"))
		})

		It("should fail if mime regex is invalid", func() {
			cfg := &Config{
				Version: "1",
				Rules: []Rule{
					{Name: "Rule", Command: "cmd", Mime: "["},
				},
			}
			err := ValidateConfig(cfg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Mime"))
			Expect(err.Error()).To(ContainSubstring("is-regex"))
		})
	})

	Describe("GetConfigPathWithProfile", func() {
		It("should return provided path if set", func() {
			path, err := GetConfigPathWithProfile("/custom/path", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal("/custom/path"))
		})

		It("should return default path if no profile", func() {
			// Mock UserHomeDir
			origHome := UserHomeDir
			UserHomeDir = func() (string, error) { return "/home/user", nil }
			defer func() { UserHomeDir = origHome }()

			path, err := GetConfigPathWithProfile("", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal(filepath.Join("/home/user", ".config", "entry", "config.yml")))
		})

		It("should return profile path", func() {
			// Mock UserHomeDir
			origHome := UserHomeDir
			UserHomeDir = func() (string, error) { return "/home/user", nil }
			defer func() { UserHomeDir = origHome }()

			path, err := GetConfigPathWithProfile("", "dev")
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal(filepath.Join("/home/user", ".config", "entry", "profiles", "dev.yml")))
		})
	})

	Describe("SaveConfig", func() {
		It("should save config to file", func() {
			cfg := &Config{Version: "1"}
			err := SaveConfig(cfgFile, cfg)
			Expect(err).NotTo(HaveOccurred())

			// Verify file exists
			_, err = os.Stat(cfgFile)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error if directory creation fails", func() {
			// Use a path where directory creation fails (e.g. file exists as directory name)
			// Create a file
			err := os.WriteFile(filepath.Join(tmpDir, "file"), []byte(""), 0644)
			Expect(err).NotTo(HaveOccurred())
			
			// Try to create config inside that file (treated as dir)
			err = SaveConfig(filepath.Join(tmpDir, "file", "config.yml"), &Config{})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ValidateRegex", func() {
		It("should pass for valid regex", func() {
			err := ValidateRegex("^test$")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail for invalid regex", func() {
			err := ValidateRegex("[")
			Expect(err).To(HaveOccurred())
		})
	})
})
