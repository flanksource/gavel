package jira

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira sync selection pagination", func() {
	It("rejects missing, repeated, and inconsistent continuation metadata", func() {
		for _, response := range []string{
			`{"issues":[{"key":"APP-1"}]}`,
			`{"issues":[{"key":"APP-1"}],"isLast":false,"nextPageToken":"same"}`,
			`{"issues":[],"isLast":true,"nextPageToken":"unexpected"}`,
			`{"isLast":true}`,
		} {
			client := testClient(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, response) })
			selection, err := (SyncOptions{}).selection(time.Now())
			Expect(err).NotTo(HaveOccurred())
			_, err = client.syncKeys(context.Background(), SyncOptions{Keys: []string{"APP"}}, selection)
			Expect(err).To(HaveOccurred(), response)
		}
	})
})
