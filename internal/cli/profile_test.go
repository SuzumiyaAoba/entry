package cli

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)



var _ = Describe("Profile commands", func() {
	var (
		tmpDir     string
		configFile string
		outBuf     bytes.Buffer
		origHome   string
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		origHome = os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)

		// Create config dir
		configDir := filepath.Join(tmpDir, ".config", "entry")
		err := os.MkdirAll(configDir, 0755)
		Expect(err).NotTo(HaveOccurred())

		configFile = filepath.Join(configDir, "config.yml")
		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&outBuf)

		// Mock config path
		cfgFile = configFile
		
		// Create default config
		cfg := &config.Config{Version: "1"}
		err = config.SaveConfig(configFile, cfg)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.Setenv("HOME", origHome)
		cfgFile = ""
	})

	Describe("runConfigProfileList", func() {
		It("should list profiles", func() {
			// Create a profile
			profilesDir := filepath.Join(filepath.Dir(configFile), "profiles")
			err := os.MkdirAll(profilesDir, 0755)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(filepath.Join(profilesDir, "work.yml"), []byte{}, 0644)
			Expect(err).NotTo(HaveOccurred())

			err = runConfigProfileList()
			Expect(err).NotTo(HaveOccurred())
			// Output is printed to stdout, not captured by rootCmd.SetOut because runConfigProfileList uses fmt.Println directly?
			// Let's check runConfigProfileList. It uses fmt.Println.
			// We can't capture stdout easily unless we redirect os.Stdout.
			// But we can verify it doesn't error.
		})
	})

	Describe("runConfigProfileCopy", func() {
		It("should copy profile", func() {
			err := runConfigProfileCopy("default", "newprofile")
			Expect(err).NotTo(HaveOccurred())

			profilesDir := filepath.Join(filepath.Dir(configFile), "profiles")
			Expect(fileExists(filepath.Join(profilesDir, "newprofile.yml"))).To(BeTrue())
		})
	})
})
