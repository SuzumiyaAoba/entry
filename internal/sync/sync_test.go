package sync_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/sync"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sync Suite")
}

var _ = Describe("Sync Client", func() {
	var (
		server *httptest.Server
		client *sync.Client
		cfg    *config.Config
	)

	BeforeEach(func() {
		cfg = &config.Config{
			Version: "1",
			Rules: []config.Rule{
				{Command: "echo test"},
			},
		}
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("CreateGist", func() {
		It("should create gist successfully", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("POST"))
				Expect(r.URL.Path).To(Equal("/gists"))
				Expect(r.Header.Get("Authorization")).To(Equal("token token"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]string{
					"id": "new-gist-id",
				})
			}))
			sync.GitHubAPIURL = server.URL
			client = sync.NewClient("token")

			id, err := client.CreateGist(cfg, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(id).To(Equal("new-gist-id"))
		})
	})

	Describe("GetGist", func() {
		It("should get gist successfully", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("GET"))
				Expect(r.URL.Path).To(Equal("/gists/gist-id"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"files": map[string]interface{}{
						"config.yml": map[string]string{
							"content": "version: \"1\"\nrules:\n  - command: echo test\n",
						},
					},
				})
			}))
			sync.GitHubAPIURL = server.URL
			client = sync.NewClient("token")

			gotCfg, err := client.GetGist("gist-id")
			Expect(err).NotTo(HaveOccurred())
			Expect(gotCfg.Rules).To(HaveLen(1))
			Expect(gotCfg.Rules[0].Command).To(Equal("echo test"))
		})
	})

	Describe("UpdateGist", func() {
		It("should update gist successfully", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("PATCH"))
				Expect(r.URL.Path).To(Equal("/gists/gist-id"))

				w.WriteHeader(http.StatusOK)
			}))
			sync.GitHubAPIURL = server.URL
			client = sync.NewClient("token")

			err := client.UpdateGist("gist-id", cfg)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
