package jira

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("site-wide field metadata", func() {
	It("uses one site/account snapshot across projects and replaces it on no-cache refresh", func() {
		calls := 0
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account-1"}`)
			case "/rest/api/3/field":
				calls++
				if calls == 1 {
					fmt.Fprint(w, `[{"id":"stale","name":"Removed"},{"id":"current","name":"Before"}]`)
				} else {
					fmt.Fprint(w, `[{"id":"current","name":"After"}]`)
				}
			default:
				http.Error(w, "unexpected metadata endpoint", http.StatusForbidden)
			}
		})
		options := CatalogOptions{Store: CatalogStore{Dir: GinkgoT().TempDir()}}
		for range 2 {
			fields, err := client.Catalog(context.Background(), options)
			Expect(err).NotTo(HaveOccurred())
			Expect(fields).To(HaveKey("stale"))
		}
		Expect(calls).To(Equal(1))
		options.NoCache = true
		fields, err := client.Catalog(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(fields).To(Equal(map[string]FieldDefinition{"current": {ID: "current", Name: "After"}}))
		Expect(calls).To(Equal(2))
	})
})
