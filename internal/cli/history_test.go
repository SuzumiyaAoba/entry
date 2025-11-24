package cli

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/SuzumiyaAoba/entry/internal/history"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)



var _ = Describe("History command", func() {
	var (
		tmpDir      string
		historyFile string
		outBuf      bytes.Buffer
	)

	BeforeEach(func() {
		resetGlobals()
		tmpDir = GinkgoT().TempDir()
		historyFile = filepath.Join(tmpDir, "history.json")
		history.SetHistoryPath(historyFile)
		outBuf.Reset()
		rootCmd.SetOut(&outBuf)
		rootCmd.SetErr(&outBuf)
	})

	AfterEach(func() {
		history.SetHistoryPath("")
	})

	It("should show empty message if no history", func() {
		err := rootCmd.RunE(rootCmd, []string{":history"})
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("No history available"))
	})

	It("should clear history", func() {
		// Add some history first
		err := history.AddEntry("cmd1", "rule1")
		Expect(err).NotTo(HaveOccurred())

		// Run clear command
		err = rootCmd.RunE(rootCmd, []string{":history", "clear"})
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("History cleared"))

		// Verify history is empty
		entries, err := history.LoadHistory()
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(BeEmpty())
	})

	Describe("runHistory", func() {
		It("should show message when no history", func() {
			// Ensure history is empty
			history.ClearHistory()

			err := runHistory(rootCmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(ContainSubstring("No history available"))
		})

		It("should fail if history load fails", func() {
			// Mock history load failure by making directory unreadable?
			// Or just rely on LoadHistory error handling which checks for file existence.
			// If file exists but is invalid JSON, it returns error?
			// Let's write invalid JSON to history file.
			
			// We need to use the path set in BeforeEach
			histFile := historyFile
			err := os.WriteFile(histFile, []byte("{invalid json}"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = runHistory(rootCmd)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to load history"))
		})
	})
})
