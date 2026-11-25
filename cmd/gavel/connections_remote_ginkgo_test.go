package main

import (
	"context"
	shared "github.com/flanksource/commons-db/cmd/query/connections"
	"github.com/flanksource/gavel/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("connection catalog remote", func() {
	It("sends the selected organization and bearer token through the existing client", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			Expect(r.Header.Get("X-Gavel-Org")).To(Equal("example"))
			Expect(r.Header.Get("Authorization")).To(Equal("Bearer test-access"))
			Expect(r.URL.Path).To(Equal("/api/connections"))
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`[]`))
			Expect(err).NotTo(HaveOccurred())
		}))
		DeferCleanup(server.Close)
		remote := connectionRemote{ctx: context.Background(), server: gavelServer{session: &api.Session{Server: server.URL, Org: "example", Token: &oauth2.Token{AccessToken: "test-access"}}}}
		rows, err := remote.Inventory(shared.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(rows).To(BeEmpty())
	})
})
