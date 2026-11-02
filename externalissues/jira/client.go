package jira

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	httpclient "github.com/flanksource/commons/http"
	"github.com/flanksource/gavel/internal/httpx"
)

type ClientOptions struct {
	Site         string
	Profile      string
	Credentials  Credentials
	HTTP         *httpclient.Client
	IgnoreFields []string
}

type Client struct {
	Site         string
	baseURL      string
	http         *httpclient.Client
	credentials  Credentials
	account      string
	ignoreFields []string
}

func NewClient(ctx context.Context, options ClientOptions) (*Client, error) {
	credentials, site, err := discoverCredentials(ctx, options)
	if err != nil {
		return nil, err
	}
	if options.Credentials.CloudID != "" {
		credentials.CloudID = options.Credentials.CloudID
	}
	if credentials.APIToken == "" && credentials.AccessToken == "" {
		return nil, fmt.Errorf("jira credentials require an API token or an OAuth access token")
	}
	if credentials.APIToken != "" && credentials.AccessToken != "" {
		return nil, fmt.Errorf("configure either a Jira API token or an OAuth access token")
	}
	if credentials.APIToken != "" && credentials.Username == "" {
		return nil, fmt.Errorf("jira API token requires a username or email")
	}
	if strings.IndexFunc(credentials.AccessToken, func(r rune) bool { return r <= ' ' || r >= 127 }) >= 0 {
		return nil, fmt.Errorf("jira OAuth access token contains invalid characters")
	}
	client := options.HTTP
	if client == nil {
		client = httpclient.NewClient().InsecureSkipVerify(false)
	}
	client.Timeout(45 * time.Second).RedirectPolicy(0)
	if err := httpx.Configure(ctx, "jira", client); err != nil {
		return nil, err
	}
	result := &Client{http: client, credentials: credentials, ignoreFields: slices.Clone(options.IgnoreFields)}
	if credentials.AccessToken != "" {
		if err := result.resolveOAuthSite(ctx, site); err != nil {
			return nil, err
		}
	} else {
		result.Site, err = normalizeSite(site)
		if err != nil {
			return nil, err
		}
		result.baseURL = result.Site
		if credentials.CloudID != "" {
			result.baseURL = "https://api.atlassian.com/ex/jira/" + url.PathEscape(credentials.CloudID)
		}
	}
	return result, nil
}

func normalizeSite(site string) (string, error) {
	if site == "" {
		return "", fmt.Errorf("jira site is required: use --site, JIRA_URL, ATLASSIAN_SITE or TWG_SITE")
	}
	if !strings.Contains(site, "://") {
		if !strings.Contains(site, ".") && !strings.Contains(site, ":") {
			site += ".atlassian.net"
		}
		site = "https://" + site
	}
	parsed, err := url.Parse(site)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("invalid Jira site URL")
	}
	if parsed.Scheme != "https" && (parsed.Scheme != "http" || (parsed.Hostname() != "localhost" && parsed.Hostname() != "127.0.0.1" && parsed.Hostname() != "::1")) {
		return "", fmt.Errorf("jira site requires HTTPS")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func (c *Client) resolveOAuthSite(ctx context.Context, site string) error {
	var resources []struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := c.json(ctx, requestOptions{Path: "https://api.atlassian.com/oauth/token/accessible-resources"}, &resources); err != nil {
		return fmt.Errorf("discover Jira OAuth sites: %w", err)
	}
	var target string
	if site != "" {
		var err error
		target, err = normalizeSite(site)
		if err != nil {
			return err
		}
	}
	for _, resource := range resources {
		if resource.ID == "" || resource.URL == "" || (target != "" && !strings.EqualFold(strings.TrimRight(resource.URL, "/"), target)) || (c.credentials.CloudID != "" && resource.ID != c.credentials.CloudID) {
			continue
		}
		if c.baseURL != "" && c.Site != resource.URL {
			return fmt.Errorf("multiple Jira OAuth sites are available; select one with --site or JIRA_URL")
		}
		c.Site = strings.TrimRight(resource.URL, "/")
		c.credentials.CloudID = resource.ID
		c.baseURL = "https://api.atlassian.com/ex/jira/" + url.PathEscape(resource.ID)
	}
	if c.baseURL == "" {
		return fmt.Errorf("selected Jira site is not available to the OAuth credentials")
	}
	return nil
}

type requestOptions struct {
	Method  string
	Path    string
	Query   url.Values
	Body    any
	Headers map[string]string
}

func (c *Client) request(ctx context.Context, options requestOptions) (*httpclient.Response, error) {
	method := options.Method
	if method == "" {
		method = http.MethodGet
	}
	endpoint := c.baseURL + "/rest/api/3" + options.Path
	if options.Path == "https://api.atlassian.com/oauth/token/accessible-resources" {
		endpoint = options.Path
	}
	request := c.http.R(ctx).Header("Accept", "application/json")
	if c.credentials.AccessToken != "" {
		request.Header("Authorization", "Bearer "+c.credentials.AccessToken)
	} else {
		request.Header("Authorization", "Basic "+basicToken(c.credentials.Username, c.credentials.APIToken))
	}
	for key, values := range options.Query {
		for _, value := range values {
			request.QueryParamAdd(key, value)
		}
	}
	for key, value := range options.Headers {
		request.Header(key, value)
	}
	if options.Body != nil {
		if options.Headers["Content-Type"] == "" {
			request.Header("Content-Type", "application/json")
		}
		if err := request.Body(options.Body); err != nil {
			return nil, fmt.Errorf("prepare Jira %s %s: %w", method, options.Path, err)
		}
	}
	response, err := request.Do(method, endpoint)
	if err != nil {
		return nil, fmt.Errorf("jira %s %s: %w", method, options.Path, err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		defer response.Body.Close()
		body, readErr := io.ReadAll(io.LimitReader(response.Body, 4096))
		if readErr != nil {
			return nil, fmt.Errorf("jira %s %s: HTTP %d; read error response: %w", method, options.Path, response.StatusCode, readErr)
		}
		message := string(body)
		for _, secret := range []string{c.credentials.APIToken, c.credentials.AccessToken, basicToken(c.credentials.Username, c.credentials.APIToken)} {
			if secret != "" {
				message = strings.ReplaceAll(message, secret, "[REDACTED]")
			}
		}
		return nil, fmt.Errorf("jira %s %s: HTTP %d: %s", method, options.Path, response.StatusCode, strings.TrimSpace(message))
	}
	return response, nil
}

func (c *Client) json(ctx context.Context, options requestOptions, result any) error {
	response, err := c.request(ctx, options)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if result == nil {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		return fmt.Errorf("decode Jira %s: %w", options.Path, err)
	}
	return nil
}

func basicToken(user, token string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + token))
}
