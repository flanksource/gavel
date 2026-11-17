package azuredevops

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	repo    Repository
	baseURL string
	http    *http.Client
	auth    authenticator
	info    *repositoryInfo
}

func (c *Client) Kind() string { return "azuredevops" }

func New(repo Repository) *Client {
	return &Client{repo: repo, baseURL: "https://dev.azure.com", http: &http.Client{Timeout: 45 * time.Second}, auth: authenticator{
		pat:        strings.TrimSpace(os.Getenv("AZURE_DEVOPS_EXT_PAT")),
		connection: func(ctx context.Context) (string, error) { return connectionPAT(ctx, repo) },
		fetch:      cliAccessToken, now: time.Now,
	}}
}
