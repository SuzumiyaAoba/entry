package cli

import (
	"bytes"
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
		err := runHistory(historyCmd)
		Expect(err).NotTo(HaveOccurred())
		Expect(outBuf.String()).To(ContainSubstring("No history available"))
	})
})
