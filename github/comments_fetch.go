package github

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/flanksource/gavel/comments"
)

type nativeComment struct {
	DatabaseID        int64     `json:"databaseId"`
	ID                int64     `json:"id"`
	Body              string    `json:"body"`
	URL               string    `json:"url"`
	HTMLURL           string    `json:"html_url"`
	CreatedAt         time.Time `json:"createdAt"`
	RESTCreatedAt     time.Time `json:"created_at"`
	UpdatedAt         string    `json:"updatedAt"`
	RESTUpdatedAt     string    `json:"updated_at"`
	Path              string    `json:"path,omitempty"`
	Line              int       `json:"line,omitempty"`
	StartLine         int       `json:"startLine,omitempty"`
	OriginalLine      int       `json:"originalLine,omitempty"`
	OriginalStartLine int       `json:"originalStartLine,omitempty"`
	DiffSide          string    `json:"diffSide,omitempty"`
	StartDiffSide     string    `json:"startDiffSide,omitempty"`
	DiffHunk          string    `json:"diffHunk,omitempty"`
	OriginalCommit    struct {
		OID string `json:"oid"`
	} `json:"originalCommit"`
	Commit struct {
		OID string `json:"oid"`
	} `json:"commit"`
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	ReplyTo *struct {
		DatabaseID int64 `json:"databaseId"`
	} `json:"replyTo"`
}

const nativeCommentFields = `databaseId body url createdAt updatedAt author { login } replyTo { databaseId } path line startLine diffHunk originalLine originalStartLine diffSide startDiffSide originalCommit { oid } commit { oid }`
const nativeThreadFields = `id viewerCanResolve isResolved isOutdated comments(first:100,after:$commentAfter){pageInfo{hasNextPage endCursor} nodes{` + nativeCommentFields + `}}`

type nativeThread struct {
	ID         string `json:"id"`
	CanResolve bool   `json:"viewerCanResolve"`
	Resolved   bool   `json:"isResolved"`
	Outdated   bool   `json:"isOutdated"`
	Comments   struct {
		PageInfo graphQLPageInfo `json:"pageInfo"`
		Nodes    []nativeComment `json:"nodes"`
	} `json:"comments"`
}

func nativeSnapshot(target comments.Target, t nativeThread) (comments.RemoteComment, error) {
	if t.ID == "" || len(t.Comments.Nodes) == 0 {
		return comments.RemoteComment{}, fmt.Errorf("native GitHub thread is incomplete")
	}
	raw, err := json.Marshal(t)
	if err != nil {
		return comments.RemoteComment{}, err
	}
	sum := sha256.Sum256(raw)
	root := t.Comments.Nodes[0]
	r := comments.RemoteComment{Binding: comments.Binding{Provider: "github", RemoteKey: target.URI + "#thread=" + t.ID, ThreadID: t.ID, RootID: strconv.FormatInt(root.DatabaseID, 10), URL: root.URL, Revision: hex.EncodeToString(sum[:]), Resolvable: t.CanResolve, Outdated: t.Outdated, Native: raw}, Body: root.Body, Author: comments.Actor{Kind: "user", ID: root.Author.Login}, CreatedAt: root.CreatedAt, Resolved: t.Resolved}
	for _, m := range t.Comments.Nodes[1:] {
		parent := ""
		if m.ReplyTo != nil {
			parent = strconv.FormatInt(m.ReplyTo.DatabaseID, 10)
		}
		r.Messages = append(r.Messages, comments.RemoteMessage{ID: strconv.FormatInt(m.DatabaseID, 10), ParentID: parent, Body: m.Body, Author: comments.Actor{Kind: "user", ID: m.Author.Login}, CreatedAt: m.CreatedAt})
	}
	return r, nil
}
func (a CommentAPI) Fetch(ctx context.Context, target comments.Target) ([]comments.RemoteComment, error) {
	repo, number, err := a.target(target)
	if err != nil {
		return nil, err
	}
	owner, name := splitCommentRepo(repo)
	var after any
	seen := map[string]bool{}
	var out []comments.RemoteComment
	for {
		var response struct {
			Data struct {
				Repository struct {
					PullRequest *struct {
						ReviewThreads struct {
							PageInfo graphQLPageInfo `json:"pageInfo"`
							Nodes    []nativeThread  `json:"nodes"`
						} `json:"reviewThreads"`
					} `json:"pullRequest"`
				} `json:"repository"`
			} `json:"data"`
		}
		query := `query($owner:String!,$repo:String!,$number:Int!,$after:String,$commentAfter:String){repository(owner:$owner,name:$repo){pullRequest(number:$number){reviewThreads(first:100,after:$after){pageInfo{hasNextPage endCursor}nodes{` + nativeThreadFields + `}}}}}`
		if err = a.graph(ctx, query, map[string]any{"owner": owner, "repo": name, "number": number, "after": after, "commentAfter": nil}, &response); err != nil {
			return nil, err
		}
		pr := response.Data.Repository.PullRequest
		if pr == nil {
			return nil, fmt.Errorf("GitHub PR not found")
		}
		for _, thread := range pr.ReviewThreads.Nodes {
			if err = a.completeNativeThread(ctx, &thread); err != nil {
				return nil, err
			}
			snapshot, err := nativeSnapshot(target, thread)
			if err != nil {
				return nil, err
			}
			out = append(out, snapshot)
		}
		page := pr.ReviewThreads.PageInfo
		if !page.HasNextPage {
			break
		}
		if page.EndCursor == "" || seen[page.EndCursor] {
			return nil, fmt.Errorf("GitHub thread cursor did not advance")
		}
		after = page.EndCursor
		seen[page.EndCursor] = true
	}
	general, err := a.generalComments(ctx, target, repo, number)
	if err != nil {
		return nil, err
	}
	return append(out, general...), nil
}
func (a CommentAPI) completeNativeThread(ctx context.Context, t *nativeThread) error {
	seen := map[string]bool{}
	for t.Comments.PageInfo.HasNextPage {
		after := t.Comments.PageInfo.EndCursor
		if after == "" || seen[after] {
			return fmt.Errorf("GitHub reply cursor did not advance")
		}
		seen[after] = true
		var response struct {
			Data struct {
				Node *nativeThread `json:"node"`
			} `json:"data"`
		}
		if err := a.graph(ctx, `query($id:ID!,$commentAfter:String){node(id:$id){... on PullRequestReviewThread{`+nativeThreadFields+`}}}`, map[string]any{"id": t.ID, "commentAfter": after}, &response); err != nil {
			return err
		}
		if response.Data.Node == nil {
			return fmt.Errorf("GitHub thread disappeared")
		}
		t.Comments.Nodes = append(t.Comments.Nodes, response.Data.Node.Comments.Nodes...)
		t.Comments.PageInfo = response.Data.Node.Comments.PageInfo
	}
	return nil
}
func splitCommentRepo(repo string) (string, string) {
	for i := range repo {
		if repo[i] == '/' {
			return repo[:i], repo[i+1:]
		}
	}
	return repo, ""
}
func (a CommentAPI) generalComments(ctx context.Context, target comments.Target, repo string, number int) ([]comments.RemoteComment, error) {
	var out []comments.RemoteComment
	for page := 1; ; page++ {
		var rows []nativeComment
		if err := a.request(ctx, "GET", fmt.Sprintf("/repos/%s/issues/%d/comments?per_page=100&page=%d", repo, number, page), nil, &rows); err != nil {
			return nil, err
		}
		for _, row := range rows {
			out = append(out, generalSnapshot(target, row))
		}
		if len(rows) < 100 {
			return out, nil
		}
	}
}
func generalSnapshot(t comments.Target, n nativeComment) comments.RemoteComment {
	raw, _ := json.Marshal(n)
	sum := sha256.Sum256(raw)
	return comments.RemoteComment{Binding: comments.Binding{Provider: "github", RemoteKey: t.URI + "#issuecomment=" + strconv.FormatInt(n.ID, 10), RootID: strconv.FormatInt(n.ID, 10), URL: n.HTMLURL, Revision: hex.EncodeToString(sum[:]), Native: raw}, Body: n.Body, Author: comments.Actor{Kind: "user", ID: n.User.Login}, CreatedAt: n.RESTCreatedAt}
}
