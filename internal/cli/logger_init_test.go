package cli

import (
	"os"

	"github.com/SuzumiyaAoba/entry/internal/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)



var _ = Describe("Logger Init", func() {
	var (
		origVerboseEnv string
		origVerbose    bool
	)

	BeforeEach(func() {
		origVerboseEnv = os.Getenv("ENTRY_VERBOSE")
		origVerbose = verbose
		// Reset global logger to Nop before each test to ensure clean state
		logger.SetGlobal(logger.NewNopLogger())
	})

	AfterEach(func() {
		os.Setenv("ENTRY_VERBOSE", origVerboseEnv)
		verbose = origVerbose
	})

	It("should use nop logger by default", func() {
		verbose = false
		os.Setenv("ENTRY_VERBOSE", "")
		
		err := initLogger()
		Expect(err).NotTo(HaveOccurred())
		
		// We can't easily check the type of the global logger without exporting it or adding a getter.
		// But we can check if it didn't error.
	})

	It("should enable verbose mode via env var", func() {
		os.Setenv("ENTRY_VERBOSE", "true")
		// verbose flag is usually set by cobra, but initLogger checks env var too
		
		err := initLogger()
		Expect(err).NotTo(HaveOccurred())
		Expect(verbose).To(BeTrue())
	})

	It("should initialize logger when verbose is true", func() {
		verbose = true
		
		err := initLogger()
		Expect(err).NotTo(HaveOccurred())
	})
})
