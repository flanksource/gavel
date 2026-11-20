// Tests the git connection comments provider over commons-db's thread writer: fetching an Azure
// DevOps PR's threads and a GitLab MR's discussions as remote comments, publishing on the captured
// diff's lines, replying against the revision a binding saw, and resolving.
package providers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/comments"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const azureThread = `{"id":7,"status":"active","lastUpdatedDate":"2026-11-04T10:00:00Z","comments":[
	{"id":11,"content":"Check this","commentType":"text","publishedDate":"2026-11-04T10:00:00Z","author":{"id":"grace-id","displayName":"Grace Hopper"}}]}`

// fakeAzure answers "METHOD path" routes and counts the writes it took.
type fakeAzure struct {
	*httptest.Server
	mu     sync.Mutex
	writes []string
}

func newFakeAzure(t *testing.T, routes map[string]string) *fakeAzure {
	t.Helper()
	fake := &fakeAzure{}
	fake.Server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		if r.Method != http.MethodGet {
			fake.mu.Lock()
			fake.writes = append(fake.writes, key)
			fake.mu.Unlock()
		}
		body, ok := routes[key]
		if !ok {
			http.Error(w, "no route "+key, http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(fake.Close)
	return fake
}

func (f *fakeAzure) provider(t *testing.T) (Connection, comments.Target) {
	t.Helper()
	client, err := git.FromConnection(dbcontext.New(), models.Connection{Name: "acme", Type: models.ConnectionTypeAzureDevops,
		URL: f.URL + "/acme", Password: "azure-pat", InsecureTLS: true})
	require.NoError(t, err)
	remote, err := client.Repository(f.URL + "/acme/Platform/_git/rules")
	require.NoError(t, err)
	resolve := func(context.Context, comments.Principal, comments.Target) (*git.Repository, error) {
		return remote, nil
	}
	return Connection{Kind: "azuredevops", Resolve: resolve}, comments.Target{URI: f.URL + "/acme/Platform/_git/rules/pullrequest/9"}
}

const pulls = "/acme/Platform/_apis/git/repositories/rules/pullrequests/9"

func TestAnAzurePullRequestsThreadsAreItsRemoteComments(t *testing.T) {
	fake := newFakeAzure(t, map[string]string{"GET " + pulls + "/threads": `{"value":[` + azureThread + `]}`})
	provider, target := fake.provider(t)

	fetched, err := provider.Fetch(context.Background(), comments.Principal{}, target)
	require.NoError(t, err)
	require.Len(t, fetched, 1)
	assert.Equal(t, "7", fetched[0].Binding.ThreadID)
	assert.Equal(t, "11", fetched[0].Binding.RootID)
	assert.Equal(t, target.URI+"#thread=7", fetched[0].Binding.RemoteKey)
	assert.True(t, fetched[0].Binding.Resolvable)
	assert.NotEmpty(t, fetched[0].Binding.Revision)
	assert.Equal(t, "Check this", fetched[0].Body)
	assert.Equal(t, comments.Actor{Kind: "user", ID: "grace-id", Label: "Grace Hopper"}, fetched[0].Author)
	assert.False(t, fetched[0].Resolved)
}

func TestAnAzureCommentIsPublishedOnTheCapturedDiffsLines(t *testing.T) {
	fake := newFakeAzure(t, map[string]string{
		"GET " + pulls:                           `{"pullRequestId":9,"status":"active","lastMergeSourceCommit":{"commitId":"head-a"},"repository":{"name":"rules","project":{"name":"Platform"}}}`,
		"GET " + pulls + "/iterations":           `{"value":[{"id":2,"sourceRefCommit":{"commitId":"head-a"},"commonRefCommit":{"commitId":"base-a"}}]}`,
		"GET " + pulls + "/iterations/2/changes": `{"changeEntries":[{"changeTrackingId":9,"item":{"path":"/widget.go"}}]}`,
		"POST " + pulls + "/threads":             azureThread,
	})
	provider, target := fake.provider(t)
	selector, err := json.Marshal(comments.CodeSelector{Kind: "diff-range", Version: 1, Path: "widget.go", BaseCommit: "base-a", HeadCommit: "head-a",
		Side: "new", StartLine: 7, EndLine: 7})
	require.NoError(t, err)
	target.Revision, target.Selector = "head-a", selector
	publication := comments.Publication{Comment: comments.Comment{Body: "Check this", Target: target}}

	published, err := provider.Apply(context.Background(), comments.Principal{}, "publish", publication)
	require.NoError(t, err)
	assert.Equal(t, "7", published.Binding.ThreadID)

	publication.Comment.Target.Revision = "old-head"
	_, err = provider.Apply(context.Background(), comments.Principal{}, "publish", publication)
	require.ErrorIs(t, err, comments.ErrConflict, "a comment captured on another head")
	assert.Equal(t, []string{"POST " + pulls + "/threads"}, fake.writes, "nothing is written for a stale capture")
}

func TestAnAzureReplyAndResolutionNeedTheRevisionTheBindingSaw(t *testing.T) {
	fake := newFakeAzure(t, map[string]string{
		"GET " + pulls + "/threads":             `{"value":[` + azureThread + `]}`,
		"GET " + pulls + "/threads/7":           azureThread,
		"POST " + pulls + "/threads/7/comments": `{"id":12,"parentCommentId":11,"content":"Done","commentType":"text","author":{"id":"ada-id","displayName":"Ada Lovelace"}}`,
		"PATCH " + pulls + "/threads/7":         strings.Replace(azureThread, `"status":"active"`, `"status":"fixed"`, 1),
	})
	provider, target := fake.provider(t)
	fetched, err := provider.Fetch(context.Background(), comments.Principal{}, target)
	require.NoError(t, err)
	require.Len(t, fetched, 1)
	binding := fetched[0].Binding

	reply, err := provider.Apply(context.Background(), comments.Principal{}, "publish-reply", comments.Publication{Comment: comments.Comment{Target: target},
		Binding: &binding, Message: &comments.Message{ID: "local-1", Body: "Done"}})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"12": "local-1"}, reply.Binding.MessageIDs)

	resolved, err := provider.Apply(context.Background(), comments.Principal{}, "resolve-native", comments.Publication{Comment: comments.Comment{Target: target}, Binding: &binding})
	require.NoError(t, err)
	assert.True(t, resolved.Resolved)

	stale := binding
	stale.Revision = "an-older-revision"
	_, err = provider.Apply(context.Background(), comments.Principal{}, "resolve-native", comments.Publication{Comment: comments.Comment{Target: target}, Binding: &stale})
	require.True(t, errors.Is(err, comments.ErrConflict), "a binding that saw another revision: %v", err)
}

func TestAGitLabMergeRequestsDiscussionsAreItsRemoteComments(t *testing.T) {
	const mr = "/api/v4/projects/acme/api/merge_requests/5"
	fake := newFakeAzure(t, map[string]string{
		"GET " + mr + "/discussions": `[
			{"id":"d1","notes":[{"id":11,"body":"Check this","resolvable":true,"resolved":false,"created_at":"2026-11-04T10:00:00Z",
			 "author":{"id":7,"name":"Grace Hopper","username":"grace"}}]},
			{"id":"d2","individual_note":true,"notes":[{"id":21,"body":"Nice.","resolvable":false,"created_at":"2026-11-04T11:00:00Z",
			 "author":{"id":8,"name":"Ada Lovelace","username":"ada"}}]}]`,
	})
	client, err := git.FromConnection(dbcontext.New(), models.Connection{Name: "acme", Type: models.ConnectionTypeGitlab,
		URL: fake.URL, Password: "glpat", InsecureTLS: true})
	require.NoError(t, err)
	remote, err := client.Repository(fake.URL + "/acme/api.git")
	require.NoError(t, err)
	provider := Connection{Kind: "gitlab", Resolve: func(context.Context, comments.Principal, comments.Target) (*git.Repository, error) {
		return remote, nil
	}}
	target := comments.Target{URI: fake.URL + "/acme/api/-/merge_requests/5"}

	assert.Equal(t, "gitlab", provider.Name())
	fetched, err := provider.Fetch(context.Background(), comments.Principal{}, target)
	require.NoError(t, err)
	require.Len(t, fetched, 2)
	assert.Equal(t, "gitlab", fetched[0].Binding.Provider)
	assert.Equal(t, "d1", fetched[0].Binding.ThreadID)
	assert.Equal(t, fake.URL+"/acme/api/-/merge_requests/5#note_11", fetched[0].Binding.URL, "the thread's own page")
	assert.True(t, fetched[0].Binding.Resolvable)
	assert.Equal(t, comments.Actor{Kind: "user", ID: "7", Label: "Grace Hopper"}, fetched[0].Author)
	assert.False(t, fetched[1].Binding.Resolvable, "an individual note can't be resolved")

	_, err = provider.Fetch(context.Background(), comments.Principal{}, comments.Target{URI: fake.URL + "/acme/Platform/_git/rules/pullrequest/9"})
	require.ErrorIs(t, err, comments.ErrInvalid, "an Azure DevOps pull request isn't the GitLab provider's")
}
