package ui

import (
	"encoding/json"
	"time"

	"github.com/flanksource/gavel/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// With no GitHub token the poller's first full fetch fails before any PR list
// exists. The stream must still send a well-formed snapshot (prs as an array)
// that names the state explicitly, or the client rejects it as an invalid
// update and the operator never learns the token is missing.
var _ = Describe("PR list stream without a GitHub token", func() {
	It("sends an empty PR list with the no-token state and an actionable message", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("GITHUB_TOKEN", "")
		GinkgoT().Setenv("GH_TOKEN", "")
		useStreamInterval(&prStreamInterval, idleStreamCadence)

		server := &Server{}
		search := func(time.Time) (github.PRSearchResults, *github.RateLimit, error) {
			return github.SearchPRs(github.Options{}, github.PRSearchOptions{})
		}
		NewPoller(server, search, time.Minute).fetchFull()

		frame := nextStreamFrame(serveStream(server.handleSSE, "/api/prs/stream"), time.Second)
		var payload map[string]any
		Expect(json.Unmarshal([]byte(frame.Data), &payload)).To(Succeed())
		Expect(payload).To(HaveKeyWithValue("prs", []any{}))
		Expect(payload).To(HaveKeyWithValue("config", HaveKeyWithValue("repos", []any{})))
		Expect(payload).To(HaveKeyWithValue("errorKind", "no-token"))
		Expect(payload).To(HaveKeyWithValue("error", noTokenMessage))
	})
})
