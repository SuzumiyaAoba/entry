package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SuzumiyaAoba/entry/internal/config"
	. "github.com/SuzumiyaAoba/entry/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("LoadConfig", func() {
	var (
		tmpDir     string
		configPath string
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		configPath = filepath.Join(tmpDir, "config.yml")
	})

	Context("when config file exists", func() {
		BeforeEach(func() {
			configContent := `
version: "1"
aliases:
  grep: rg
rules:
  - extensions: [txt]
    command: "echo text"
  - regex: ".*\\.log$"
    command: "echo log"
`
			err := os.WriteFile(configPath, []byte(configContent), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should load config successfully", func() {
			cfg, err := LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())
			Expect(cfg.Version).To(Equal("1"))
			Expect(cfg.Rules).To(HaveLen(2))
			Expect(cfg.Aliases).To(HaveKeyWithValue("grep", "rg"))

			foundRegex := false
			for _, rule := range cfg.Rules {
				if rule.Regex == ".*\\.log$" {
					foundRegex = true
					break
				}
			}
			Expect(foundRegex).To(BeTrue(), "LoadConfig() did not load regex rule")
		})
	})

	Context("when config file does not exist", func() {
		It("should return an error", func() {
			_, err := LoadConfig(filepath.Join(tmpDir, "nonexistent.yml"))
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("ValidateConfig", func() {
	It("should return nil for valid config", func() {
		cfg := &config.Config{
			Rules: []config.Rule{
				{Command: "echo", Regex: "^test$", Mime: "text/plain"},
			},
		}
		Expect(config.ValidateConfig(cfg)).To(Succeed())
	})

	It("should return error for missing command", func() {
		cfg := &config.Config{
			Rules: []config.Rule{
				{Command: ""},
			},
		}
		Expect(config.ValidateConfig(cfg)).To(HaveOccurred())
	})

	It("should return error for invalid regex", func() {
		cfg := &config.Config{
			Rules: []config.Rule{
				{Command: "echo", Regex: "["},
			},
		}
		Expect(config.ValidateConfig(cfg)).To(HaveOccurred())
	})

	It("should return error for invalid mime regex", func() {
		cfg := &config.Config{
			Rules: []config.Rule{
				{Command: "echo", Mime: "["},
			},
		}
		Expect(config.ValidateConfig(cfg)).To(HaveOccurred())
	})
})

var _ = Describe("GetConfigPath", func() {
	It("should return provided path if not empty", func() {
		path, err := GetConfigPath("/tmp/config.yml")
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal("/tmp/config.yml"))
	})

	It("should return default path if empty", func() {
		path, err := GetConfigPath("")
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(ContainSubstring(".config/entry/config.yml"))
	})
})

var _ = Describe("GetConfigPathWithProfile", func() {
	It("should return provided path if not empty", func() {
		path, err := GetConfigPathWithProfile("/tmp/config.yml", "profile")
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal("/tmp/config.yml"))
	})

	It("should return default path if profile is empty", func() {
		path, err := GetConfigPathWithProfile("", "")
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(ContainSubstring(".config/entry/config.yml"))
	})

	It("should return profile path if profile is provided", func() {
		path, err := GetConfigPathWithProfile("", "work")
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(ContainSubstring(".config/entry/profiles/work.yml"))
	})
})

var _ = Describe("SaveConfig", func() {
	var tmpDir string

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
	})

	It("should save config to file", func() {
		cfg := &config.Config{
			Version: "1",
			Rules: []config.Rule{
				{Command: "echo"},
			},
		}
		path := filepath.Join(tmpDir, "config.yml")
		err := SaveConfig(path, cfg)
		Expect(err).NotTo(HaveOccurred())

		// Verify file content
		content, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("version: \"1\""))
		Expect(string(content)).To(ContainSubstring("command: echo"))
	})

	It("should return error if directory is not writable", func() {
		// Create read-only directory
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		err := os.Mkdir(readOnlyDir, 0555)
		Expect(err).NotTo(HaveOccurred())

		path := filepath.Join(readOnlyDir, "config.yml")
		cfg := &config.Config{Version: "1"}
		err = SaveConfig(path, cfg)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("LoadConfig Error Paths", func() {
	var tmpDir string
	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
	})

	It("should return error for invalid YAML", func() {
		path := filepath.Join(tmpDir, "invalid.yml")
		err := os.WriteFile(path, []byte("invalid: yaml: :"), 0644)
		Expect(err).NotTo(HaveOccurred())

		_, err = LoadConfig(path)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("yaml: "))
	})
})

var _ = Describe("ValidateRegex", func() {
	It("should return nil for valid regex", func() {
		Expect(ValidateRegex(".*")).To(Succeed())
	})

	It("should return error for invalid regex", func() {
		Expect(ValidateRegex("[")).To(HaveOccurred())
	})
})
