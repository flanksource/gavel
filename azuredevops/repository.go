package azuredevops

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Repository struct {
	Organization string
	Project      string
	Name         string
}

func ParseRepository(raw string) (Repository, error) {
	u, err := parseURL(raw)
	if err != nil {
		return Repository{}, err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	host := strings.ToLower(u.Hostname())
	if u.Scheme == "ssh" {
		if (host != "ssh.dev.azure.com" && host != "vs-ssh.visualstudio.com") || len(parts) != 4 || parts[0] != "v3" {
			return Repository{}, fmt.Errorf("invalid hosted Azure SSH repository")
		}
		return validateRepository(Repository{Organization: parts[1], Project: parts[2], Name: parts[3]})
	}
	if u.Scheme != "https" || u.Port() != "" {
		return Repository{}, fmt.Errorf("azure repository requires hosted HTTPS or SSH")
	}
	var org string
	switch {
	case host == "dev.azure.com":
		if len(parts) < 1 {
			return Repository{}, fmt.Errorf("missing Azure organization")
		}
		org, parts = parts[0], parts[1:]
	case strings.HasSuffix(host, ".visualstudio.com"):
		org = strings.TrimSuffix(host, ".visualstudio.com")
		if strings.Contains(org, ".") {
			return Repository{}, fmt.Errorf("invalid Azure organization host")
		}
		if len(parts) > 0 && strings.EqualFold(parts[0], "DefaultCollection") {
			parts = parts[1:]
		}
	default:
		return Repository{}, fmt.Errorf("unsupported Azure repository host %q", host)
	}
	if len(parts) == 2 && parts[0] == "_git" {
		return validateRepository(Repository{Organization: org, Name: parts[1]})
	}
	if len(parts) != 3 || parts[1] != "_git" {
		return Repository{}, fmt.Errorf("expected Azure project/_git/repository path")
	}
	return validateRepository(Repository{Organization: org, Project: parts[0], Name: parts[2]})
}

func ParsePRURL(raw string) (Repository, int, error) {
	u, err := parseURL(raw)
	if err != nil {
		return Repository{}, 0, err
	}
	base, number, found := strings.Cut(u.Path, "/pullrequest/")
	n, err := strconv.Atoi(number)
	if !found || err != nil || n <= 0 {
		return Repository{}, 0, fmt.Errorf("expected a positive Azure PR identifier")
	}
	u.Path, u.RawPath, u.RawQuery, u.Fragment = base, "", "", ""
	repo, err := ParseRepository(u.String())
	return repo, n, err
}

func (r Repository) URL() string {
	base := "https://dev.azure.com/" + url.PathEscape(r.Organization)
	if r.Project != "" {
		base += "/" + url.PathEscape(r.Project)
	}
	return base + "/_git/" + url.PathEscape(r.Name)
}

func parseURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if !strings.Contains(raw, "://") {
		if host, path, ok := strings.Cut(raw, ":"); ok {
			raw = "ssh://" + host + "/" + path
		}
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid Azure URL: %w", err)
	}
	return u, nil
}

func validateRepository(repo Repository) (Repository, error) {
	if repo.Organization == "" || repo.Name == "" {
		return Repository{}, fmt.Errorf("azure organization and repository are required")
	}
	for _, part := range []string{repo.Organization, repo.Project, repo.Name} {
		if part == "." || part == ".." || strings.ContainsAny(part, "/\\") {
			return Repository{}, fmt.Errorf("invalid Azure repository path component")
		}
	}
	return repo, nil
}
