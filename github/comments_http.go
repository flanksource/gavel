package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/internal/httpx"
)

type CommentAPI struct{ Options Options }

func (a CommentAPI) target(t comments.Target) (string, int, error) {
	u, err := url.Parse(t.URI)
	if err != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", 0, fmt.Errorf("invalid GitHub PR target")
	}
	p := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(p) != 4 || p[2] != "pull" {
		return "", 0, fmt.Errorf("invalid GitHub PR path")
	}
	n, err := strconv.Atoi(p[3])
	if err != nil || n < 1 {
		return "", 0, fmt.Errorf("invalid GitHub PR number")
	}
	repo := p[0] + "/" + p[1]
	if a.Options.Repo != repo {
		return "", 0, fmt.Errorf("GitHub credentials are not scoped to target repository")
	}
	return repo, n, nil
}
func (a CommentAPI) request(ctx context.Context, method, path string, input, output any) error {
	token, err := a.Options.token()
	if err != nil {
		return err
	}
	var data []byte
	if input != nil {
		data, err = json.Marshal(input)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, githubAPIBase()+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpx.Shared.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GitHub comments %s %s: HTTP %d: %s", method, path, resp.StatusCode, body)
	}
	if output != nil {
		return json.Unmarshal(body, output)
	}
	return nil
}
func (a CommentAPI) graph(ctx context.Context, query string, variables map[string]any, out any) error {
	var raw json.RawMessage
	if err := a.request(ctx, http.MethodPost, "/graphql", map[string]any{"query": query, "variables": variables}, &raw); err != nil {
		return err
	}
	if err := graphQLErrors(raw); err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}
