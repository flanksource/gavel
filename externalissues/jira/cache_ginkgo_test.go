package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/externalissues/syncdoc"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira saved issue reads", func() {
	It("isolates identical issue keys by site and rejects ambiguous site selection", func() {
		root := GinkgoT().TempDir()
		for _, site := range []string{"https://alpha.example.test", "https://beta.example.test"} {
			issue, err := ParseIssue(json.RawMessage(`{"id":"123","key":"EXAMPLE-1","summary":"Cached site","project":{"id":"10001","key":"EXAMPLE"}}`), ParseOptions{Site: site})
			Expect(err).NotTo(HaveOccurred())
			domain, err := siteDomain(site)
			Expect(err).NotTo(HaveOccurred())
			dir := filepath.Join(root, domain, "EXAMPLE")
			Expect(os.MkdirAll(dir, 0700)).To(Succeed())
			doc := &syncdoc.Document{}
			Expect(prepareSyncMetadata(doc, issue, SyncOptions{})).To(Succeed())
			_, err = saveSnapshot(issue, nil, dir, doc)
			Expect(err).NotTo(HaveOccurred())
			body, err := doc.Markdown("Saved site body")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(filepath.Join(dir, "EXAMPLE-1.md"), body, 0600)).To(Succeed())
		}
		issue, err := Read(context.Background(), ReadOptions{Dir: root, Key: "123", TTL: 24 * time.Hour, Client: ClientOptions{Site: "https://alpha.example.test"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(issue.Site).To(Equal("https://alpha.example.test"))
		_, _, err = findSavedIssue(root, "", "EXAMPLE-1")
		Expect(err).To(MatchError(ContainSubstring("ambiguous")))
	})
	It("syncs on a miss, reuses fresh data without credentials, and refreshes at expiry", func() {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("XDG_CACHE_HOME", root)
		reads := 0
		failIssue := false
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case r.URL.Path == "/rest/api/3/field":
				fmt.Fprint(w, `[{"id":"customfield_1","name":"Enabled","schema":{"type":"boolean"}}]`)
			case filepath.Base(r.URL.Path) == "remotelink":
				fmt.Fprint(w, `[]`)
			case filepath.Base(r.URL.Path) == "comment":
				fmt.Fprint(w, `{"total":1,"comments":[{"id":"c1","body":"Saved comment"}]}`)
			default:
				reads++
				if failIssue {
					http.Error(w, "unavailable", http.StatusServiceUnavailable)
					return
				}
				fmt.Fprintf(w, `{"id":"123","key":"EXAMPLE-1","fields":{"summary":"Version %d","project":{"id":"10001","key":"EXAMPLE"},"description":"Saved description","customfield_1":false}}`, reads)
			}
		})
		opts := ReadOptions{Key: "EXAMPLE-1", TTL: 24 * time.Hour, Full: true, Client: ClientOptions{Site: client.Site, Credentials: client.credentials}, Store: CatalogStore{Dir: GinkgoT().TempDir()}}
		issue, err := Read(context.Background(), opts)
		Expect(err).NotTo(HaveOccurred())
		site, err := url.Parse(client.Site)
		Expect(err).NotTo(HaveOccurred())
		Expect(filepath.Join(root, ".gavel", "jira", site.Host, "EXAMPLE", "EXAMPLE-1.md")).To(BeAnExistingFile())
		Expect(issue.Title).To(Equal("Version 1"))
		Expect(issue.Fields["customfield_1"].Name).To(Equal("Enabled"))
		Expect(string(issue.Fields["customfield_1"].Schema)).To(MatchJSON(`{"type":"boolean"}`))
		data, err := json.Marshal(issue)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring(`"enabled":false`))
		full, err := clicky.Format(issue, clicky.FormatOptions{Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(full).To(ContainSubstring("Saved comment"))
		opts.Client.Credentials = Credentials{}
		opts.Key = "123"
		issue, err = Read(context.Background(), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(issue.Title).To(Equal("Version 1"))
		Expect(reads).To(Equal(1))
		opts.Client.Credentials = client.credentials
		cacheDir, err := SyncDir("")
		Expect(err).NotTo(HaveOccurred())
		path, doc, err := findSavedIssue(cacheDir, client.Site, "EXAMPLE-1")
		Expect(err).NotTo(HaveOccurred())
		Expect(doc).NotTo(BeNil())
		opts.Now = func() time.Time { return doc.Metadata.Sync.SyncedAt.Add(24 * time.Hour) }
		issue, err = Read(context.Background(), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(issue.Title).To(Equal("Version 2"))
		Expect(reads).To(Equal(2))
		opts.Now = nil
		opts.TTL = 0
		_, err = Read(context.Background(), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(reads).To(Equal(3))
		opts.TTL = 24 * time.Hour
		opts.Client.IgnoreFields = []string{"Enabled"}
		issue, err = Read(context.Background(), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(issue.Fields).NotTo(HaveKey("customfield_1"))
		Expect(reads).To(Equal(4))
		before, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		failIssue = true
		opts.TTL = 0
		_, err = Read(context.Background(), opts)
		Expect(err).To(MatchError(ContainSubstring("HTTP 503")))
		after, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
		failIssue = false
		opts.TTL = 24 * time.Hour
		opts.NoCache = true
		_, err = Read(context.Background(), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(reads).To(Equal(6))
		opts.NoCache = false
		doc, err = readSyncDocument(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(filepath.Dir(path), doc.Metadata.Snapshot.Path), []byte("corrupt"), 0600)).To(Succeed())
		_, err = Read(context.Background(), opts)
		Expect(err).To(MatchError(ContainSubstring("hash mismatch")))
		Expect(reads).To(Equal(6))
	})

	It("respects explicit directories and validates XDG cache paths", func() {
		GinkgoT().Setenv("XDG_CACHE_HOME", "relative")
		_, err := SyncDir("")
		Expect(err).To(MatchError(ContainSubstring("must be absolute")))
		custom := GinkgoT().TempDir()
		actual, err := SyncDir(custom)
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(custom))
		GinkgoT().Setenv("XDG_CACHE_HOME", "")
		platform, err := os.UserCacheDir()
		Expect(err).NotTo(HaveOccurred())
		actual, err = SyncDir("")
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(filepath.Join(platform, ".gavel", "jira")))
	})
})
