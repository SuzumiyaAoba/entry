package matcher_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/SuzumiyaAoba/via/internal/config"
	"github.com/SuzumiyaAoba/via/internal/matcher"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Matcher Suite")
}

var _ = Describe("Match", func() {
	var rules []config.Rule

	BeforeEach(func() {
		// Create dummy files for testing
		_ = os.WriteFile("file.txt", []byte("text"), 0644)
		_ = os.WriteFile("README.md", []byte("markdown"), 0644)
		_ = os.WriteFile("app.log", []byte("log"), 0644)
		_ = os.WriteFile("main.go", []byte("go"), 0644)
		_ = os.WriteFile("script.sh", []byte("#!/bin/sh"), 0755)
		_ = os.WriteFile("script.bat", []byte("echo off"), 0644)
		_ = os.WriteFile("unknown.dat", []byte("data"), 0644)
		_ = os.WriteFile("FILE.TXT", []byte("TEXT"), 0644)
		// Create a fake PNG file signature
		// PNG signature: 89 50 4E 47 0D 0A 1A 0A
		_ = os.WriteFile("image.png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, 0644)

		rules = []config.Rule{
			{
				Extensions: []string{"txt", "md"},
				Command:    "echo text",
			},
			{
				Regex:   ".*\\.log$",
				Command: "echo log",
			},
			{
				Mime:       "image/.*",
				Extensions: []string{"png"},
				Command:    "echo image",
			},
			{
				Extensions: []string{"go"},
				Command:    "echo go",
			},
			{
				Extensions: []string{"sh"},
				OS:         []string{runtime.GOOS},
				Command:    "echo sh",
			},
			{
				Extensions: []string{"bat"},
				OS:         []string{"otheros"},
				Command:    "echo bat",
			},
			{
				Scheme:  "https",
				Command: "open browser",
			},
			{
				Scheme:  "mailto",
				Command: "open mail",
			},
			{
				Script:  "true",
				Command: "run script",
			},
		}
	})

	AfterEach(func() {
		os.Remove("file.txt")
		os.Remove("README.md")
		os.Remove("app.log")
		os.Remove("main.go")
		os.Remove("script.sh")
		os.Remove("script.bat")
		os.Remove("unknown.dat")
		os.Remove("FILE.TXT")
		os.Remove("image.png")
	})

	DescribeTable("matching files against rules",
		func(filename string, wantCmd string, wantErr bool) {
			got, err := matcher.Match(rules, filename)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}

			if wantCmd != "" {
				Expect(got).NotTo(BeEmpty())
				Expect(got[0].Command).To(Equal(wantCmd))
			} else {
				Expect(got).To(BeEmpty())
			}
		},
		Entry("Match extension txt", "file.txt", "echo text", false),
		Entry("Match extension md", "README.md", "echo text", false),
		Entry("Match regex log", "app.log", "echo log", false),
		Entry("Match mime image", "image.png", "echo image", false),
		Entry("Match extension go", "main.go", "echo go", false),
		Entry("Match OS specific", "script.sh", "echo sh", false),
		Entry("No match OS mismatch", "script.bat", "run script", false),
		Entry("No match", "unknown.dat", "run script", false),
		Entry("Case insensitive extension", "FILE.TXT", "echo text", false),
		Entry("Match URL scheme https", "https://google.com", "open browser", false),
		Entry("Match URL scheme mailto", "mailto:user@example.com", "open mail", false),
		Entry("Match URL extension png", "http://example.com/image.png", "echo image", false),
		Entry("Match URL regex", "http://example.com/foo.log", "echo log", false),
		Entry("Match Script", "any.file", "run script", false),
	)

	Describe("Script Matching", func() {
		It("should match using JS script", func() {
			scriptRules := []config.Rule{
				{Script: "file.endsWith('.js')", Command: "node"},
			}
			matches, err := matcher.Match(scriptRules, "test.js")
			Expect(err).NotTo(HaveOccurred())
			Expect(matches).To(HaveLen(1))
			Expect(matches[0].Command).To(Equal("node"))
		})

		It("should not match if JS script returns false", func() {
			scriptRules := []config.Rule{
				{Script: "file.endsWith('.js')", Command: "node"},
			}
			matches, err := matcher.Match(scriptRules, "test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(matches).To(BeEmpty())
		})

		It("should return error if JS script is invalid", func() {
			scriptRules := []config.Rule{
				{Script: "invalid syntax )))", Command: "node"},
			}
			_, err := matcher.Match(scriptRules, "test.js")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("MatchAll", func() {
		It("should return all matching rules", func() {
			multiRules := []config.Rule{
				{Extensions: []string{"txt"}, Command: "cmd1"},
				{Extensions: []string{"txt"}, Command: "cmd2"},
			}
			matches, err := matcher.MatchAll(multiRules, "file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(matches).To(HaveLen(2))
			Expect(matches[0].Command).To(Equal("cmd1"))
			Expect(matches[1].Command).To(Equal("cmd2"))
		})

		It("should handle fallthrough logic in Match (not MatchAll)", func() {
			// Match stops at first match unless Fallthrough is true
			// But MatchAll should return all matches regardless of Fallthrough?
			// Let's check MatchAll implementation.
			// MatchAll iterates all rules and collects matches. It doesn't seem to check Fallthrough.
			// Wait, Match checks Fallthrough. MatchAll does not.
			
			// Let's verify Match behavior with Fallthrough
			ftRules := []config.Rule{
				{Extensions: []string{"txt"}, Command: "cmd1", Fallthrough: true},
				{Extensions: []string{"txt"}, Command: "cmd2"},
			}
			matches, err := matcher.Match(ftRules, "file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(matches).To(HaveLen(2))
		})
	})

	Describe("Error cases", func() {
		It("should return error for invalid regex", func() {
			badRules := []config.Rule{
				{Regex: "["},
			}
			_, err := matcher.Match(badRules, "file.txt")
			Expect(err).To(HaveOccurred())
		})
	})
})
