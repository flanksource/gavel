package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira sync selection command", func() {
	It("leaves an in-place resync byte-identical when its status is filtered", func() {
		status := "Open"
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/myself"):
				fmt.Fprint(w, `{"accountId":"account"}`)
			case strings.HasSuffix(r.URL.Path, "/field"), strings.HasSuffix(r.URL.Path, "/remotelink"):
				fmt.Fprint(w, `[]`)
			case strings.HasSuffix(r.URL.Path, "/comment"):
				Expect(status).To(Equal("Open"))
				fmt.Fprint(w, `{"total":0,"comments":[]}`)
			default:
				fmt.Fprintf(w, `{"key":"APP-1","fields":{"summary":"In-place selection","project":{"id":"1","key":"APP"},"status":{"name":%q}}}`, status)
			}
		})
		dir := GinkgoT().TempDir()
		_, err := commandOutput([]string{"sync", "APP-1", "--sync-dir", dir}, "json")
		Expect(err).NotTo(HaveOccurred())
		path := filepath.Join(dir, syncTestDomain(), "APP", "APP-1.md")
		before, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		status = "Done"
		output, err := commandOutput([]string{"sync", "--from", path, "--status", "!Done"}, "json")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(MatchJSON(`[]`))
		after, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
	})
	It("paginates mixed targets, filters before enrichment, and preserves excluded files", func() {
		catalogs, pages, comments := 0, 0, 0
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/myself"):
				fmt.Fprint(w, `{"accountId":"account"}`)
			case strings.HasSuffix(r.URL.Path, "/field"):
				catalogs++
				fmt.Fprint(w, `[{"id":"customfield_1","name":"Countries"}]`)
			case strings.HasSuffix(r.URL.Path, "/search/jql"):
				var request struct {
					JQL   string `json:"jql"`
					After string `json:"nextPageToken"`
				}
				Expect(json.NewDecoder(r.Body).Decode(&request)).To(Succeed())
				Expect(request.JQL).To(ContainSubstring("updated >= 1793491200000"))
				pages++
				if request.After == "" {
					fmt.Fprint(w, `{"issues":[{"key":"APP-1"}],"isLast":false,"nextPageToken":"next"}`)
				} else {
					Expect(request.After).To(Equal("next"))
					fmt.Fprint(w, `{"issues":[{"key":"APP-2"},{"key":"APP-3"}],"isLast":true}`)
				}
			case strings.HasSuffix(r.URL.Path, "/comment"):
				Expect(r.URL.Path).To(ContainSubstring("APP-1"))
				comments++
				fmt.Fprint(w, `{"total":0,"comments":[]}`)
			case strings.HasSuffix(r.URL.Path, "/remotelink"):
				Expect(r.URL.Path).To(ContainSubstring("APP-1"))
				fmt.Fprint(w, `[]`)
			case strings.Contains(r.URL.Path, "/issue/"):
				key := filepath.Base(r.URL.Path)
				if key == "10001" {
					key = "APP-1"
				}
				status, updated := "Open", "2026-11-01T00:00:00Z"
				if key == "APP-2" {
					status = "Done"
				}
				if key == "APP-3" {
					updated = "2026-10-01T00:00:00Z"
				}
				fmt.Fprintf(w, `{"id":"10001","key":%q,"fields":{"summary":"Selection example","project":{"id":"1","key":"APP"},"updated":%q,"status":{"name":%q},"customfield_1":[{"value":"Kenya"}]}}`, key, updated, status)
			default:
				http.NotFound(w, r)
			}
		})
		dir := GinkgoT().TempDir()
		excluded := filepath.Join(dir, syncTestDomain(), "APP", "APP-2.md")
		Expect(os.MkdirAll(filepath.Dir(excluded), 0o700)).To(Succeed())
		Expect(os.WriteFile(excluded, []byte("keep existing document"), 0o600)).To(Succeed())
		output, err := commandOutput([]string{"sync", "APP", "APP-1..APP-3", "10001", "--since", "2026-11-01", "--status", "!Done,!Closed", "--filter", "Countries=Ken*,Uganda", "--filter", "Countries=!Blocked", "--sync-dir", dir, "--no-cache"}, "json")
		Expect(err).NotTo(HaveOccurred())
		var results []struct {
			Key string `json:"key"`
		}
		Expect(json.Unmarshal([]byte(output), &results)).To(Succeed())
		Expect(results).To(Equal([]struct {
			Key string `json:"key"`
		}{{"APP-1"}}))
		Expect(catalogs).To(Equal(1))
		Expect(pages).To(Equal(4))
		Expect(comments).To(Equal(1))
		data, err := os.ReadFile(excluded)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("keep existing document"))
	})
	It("rejects invalid selection syntax before credentials", func() {
		for _, args := range [][]string{{"APP-3..APP-1"}, {"APP", "--since", "invalid"}, {"APP", "--filter", "Priority"}} {
			_, err := commandOutput(append([]string{"sync"}, args...), "json")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).NotTo(ContainSubstring("credential"))
		}
	})
	It("returns an empty JSON array for an empty project", func() {
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/rest/api/3/search/jql"))
			fmt.Fprint(w, `{"issues":[],"isLast":true}`)
		})
		output, err := commandOutput([]string{"sync", "APP"}, "json")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(MatchJSON(`[]`))
	})
})
