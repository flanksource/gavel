// Tests the merged results page: /results/{repo}/{id,id,...} serves several
// shard artifacts as one run and hands the page the PR's shard list to filter by.
package ui

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var shardsScript = regexp.MustCompile(`window\.__gavelShards = (.*?);</script>`)

var _ = Describe("merged shard results", func() {
	const mergedKey = "9001,9002"

	BeforeEach(func() {
		srv := testui.NewServer()
		srv.LoadSnapshot(testui.Snapshot{Tests: []parsers.Test{{Name: "TestFromShardA", Passed: true}, {Name: "TestFromShardB", Failed: true}}})
		srv.MarkDone()
		globalArtifactCache.put(mergedKey, &artifactEntry{srv: srv, handler: srv.Handler()})
		DeferCleanup(func() {
			globalArtifactCache.mu.Lock()
			delete(globalArtifactCache.entries, mergedKey)
			globalArtifactCache.mu.Unlock()
		})
	})

	It("serves the merged run's API under the comma-separated artifact prefix", func() {
		base := newDashboardTestServer(&Server{}).URL
		resp := doRequest(newRequest(http.MethodGet, base+"/results/acme/widgets/"+mergedKey+"/api/tests", nil, nil))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		body, err := io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(body)).To(And(ContainSubstring("TestFromShardA"), ContainSubstring("TestFromShardB")))
	})

	It("hands the page every shard, marking the ones in the path selected", func() {
		base := newDashboardTestServer(&Server{}).URL
		resp := doRequest(newRequest(http.MethodGet, base+"/results/acme/widgets/"+mergedKey+"/lint?shard=9001:gavel-lint&shard=9002:gavel-test&shard=9003:gavel-bench", nil, nil))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		body, err := io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		m := shardsScript.FindSubmatch(body)
		Expect(m).NotTo(BeNil(), string(body))
		var got struct {
			Base   string         `json:"base"`
			Shards []resultsShard `json:"shards"`
		}
		Expect(json.Unmarshal(m[1], &got)).To(Succeed())
		Expect(got.Base).To(Equal("/results/acme/widgets"))
		Expect(got.Shards).To(Equal([]resultsShard{
			{ArtifactID: 9001, Label: "gavel-lint", Selected: true},
			{ArtifactID: 9002, Label: "gavel-test", Selected: true},
			{ArtifactID: 9003, Label: "gavel-bench", Selected: false},
		}))
	})

	It("rejects a shard parameter without a label", func() {
		base := newDashboardTestServer(&Server{}).URL
		resp := doRequest(newRequest(http.MethodGet, base+"/results/acme/widgets/"+mergedKey+"/lint?shard=9001", nil, nil))
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
	})
})
