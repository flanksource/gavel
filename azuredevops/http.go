package azuredevops

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
)

func (c *Client) request(ctx context.Context, method, path string, query url.Values, payload any) ([]byte, http.Header, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, nil, err
	}
	q := u.Query()
	for key, values := range query {
		q[key] = values
	}
	if q.Get("api-version") == "" {
		q.Set("api-version", "7.1")
	}
	u.RawQuery = q.Encode()
	var body []byte
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("encode Azure request: %w", err)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	auth, err := c.auth.authorization(ctx)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if strings.Contains(path, "/logs/") {
		req.Header.Set("Accept", "text/plain")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("azure %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read Azure response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("azure %s %s: HTTP %d: %s", method, path, resp.StatusCode, data)
	}
	return data, resp.Header, nil
}

func collection[T any](ctx context.Context, c *Client, path string, query url.Values) ([]T, error) {
	q := url.Values{}
	for key, values := range query {
		q[key] = append([]string{}, values...)
	}
	q.Set("$top", "100")
	var all []T
	seen := map[string]bool{}
	previousPage := ""
	for {
		data, headers, err := c.request(ctx, http.MethodGet, path, q, nil)
		if err != nil {
			return nil, err
		}
		var page struct {
			Value *[]T `json:"value"`
		}
		if err := json.Unmarshal(data, &page); err != nil {
			return nil, fmt.Errorf("decode Azure collection %s: %w", path, err)
		}
		if page.Value == nil {
			return nil, fmt.Errorf("azure collection %s is missing value", path)
		}
		if q.Get("$skip") != "" && string(data) == previousPage {
			return nil, fmt.Errorf("azure collection %s did not advance its page", path)
		}
		previousPage = string(data)
		all = append(all, (*page.Value)...)
		if next := headers.Get("x-ms-continuationtoken"); next != "" {
			if seen[next] {
				return nil, fmt.Errorf("azure collection %s repeated continuation token", path)
			}
			seen[next] = true
			q.Set("continuationToken", next)
		} else {
			if len(*page.Value) < 100 || q.Get("continuationToken") != "" || (!strings.HasSuffix(path, "/pullrequests") && !strings.HasSuffix(path, "/evaluations")) {
				return all, nil
			}
			q.Set("$skip", strconv.Itoa(len(all)))
		}
	}
}

func (c *Client) apiPath(resource string) string {
	return c.projectPath(c.repo.Project, resource)
}

func (c *Client) projectPath(project, resource string) string {
	path := "/" + url.PathEscape(c.repo.Organization)
	if project != "" {
		path += "/" + url.PathEscape(project)
	}
	return path + "/_apis/" + resource
}

func (c *Client) repositoryPath() string {
	id := c.repo.Name
	if c.info != nil {
		id = c.info.ID
	}
	return c.apiPath("git/repositories/" + url.PathEscape(id))
}
