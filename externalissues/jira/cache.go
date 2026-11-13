package jira

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/flanksource/gavel/externalissues/syncdoc"
)

type ReadOptions struct {
	Key     string
	Dir     string
	TTL     time.Duration
	Full    bool
	NoCache bool
	Client  ClientOptions
	Store   CatalogStore
	Now     func() time.Time
}

func SyncDir(dir string) (string, error) {
	if dir != "" {
		return filepath.Abs(dir)
	}
	root := os.Getenv("XDG_CACHE_HOME")
	if root != "" && !filepath.IsAbs(root) {
		return "", fmt.Errorf("XDG_CACHE_HOME must be absolute: %q", root)
	}
	if root == "" {
		var err error
		root, err = os.UserCacheDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(root, ".gavel", "jira"), nil
}

func siteDomain(site string) (string, error) {
	normalized, err := normalizeSite(site)
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", err
	}
	return strings.ToLower(parsed.Host), nil
}

func Read(ctx context.Context, options ReadOptions) (Issue, error) {
	if err := ctx.Err(); err != nil {
		return Issue{}, err
	}
	if strings.TrimSpace(options.Key) == "" || options.TTL < 0 {
		return Issue{}, fmt.Errorf("jira get requires an issue key and a nonnegative TTL")
	}
	dir, err := SyncDir(options.Dir)
	if err != nil {
		return Issue{}, err
	}
	site := options.Client.Site
	if site == "" {
		site = firstEnv(os.Getenv, "JIRA_URL", "JIRA_BASE_URL", "JIRA_SITE", "ATLASSIAN_SITE", "ATLASSIAN_URL", "TWG_SITE", "TWG_DOMAIN")
	}
	if site != "" {
		site, err = normalizeSite(site)
		if err != nil {
			return Issue{}, err
		}
	}
	path, doc, err := findSavedIssue(dir, site, options.Key)
	if err != nil {
		return Issue{}, err
	}
	now := time.Now()
	if options.Now != nil {
		now = options.Now()
	}
	if doc != nil {
		issue, err := readSnapshot(path, doc)
		if err != nil {
			return Issue{}, err
		}
		at := doc.Metadata.Sync.SyncedAt
		if issue != nil && !options.NoCache && options.TTL > 0 && !at.IsZero() && !now.Before(at) && now.Sub(at) < options.TTL && slices.Equal(issue.cachedIgnoreFields, options.Client.IgnoreFields) {
			issue.Full = options.Full
			return *issue, nil
		}
		if site == "" {
			site = doc.Metadata.Source.Site
		}
	}
	options.Client.Site = site
	if doc != nil {
		options.Client.Credentials.CloudID = doc.Metadata.Source.CloudID
	}
	client, err := NewClient(ctx, options.Client)
	if err != nil {
		return Issue{}, err
	}
	syncOptions := SyncOptions{Keys: []string{options.Key}, Dir: dir, NoCache: options.NoCache, Store: options.Store}
	if doc != nil {
		syncOptions = SyncOptions{From: path, NoCache: options.NoCache, Store: options.Store}
	}
	results, err := client.Sync(ctx, syncOptions)
	if err != nil {
		return Issue{}, err
	}
	return loadSavedIssue(results[0].Path, options.Full)
}

func findSavedIssue(dir, site, key string) (path string, doc *syncdoc.Document, returnErr error) {
	err := filepath.WalkDir(dir, func(name string, entry fs.DirEntry, err error) error {
		if os.IsNotExist(err) && name == dir {
			return nil
		}
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if name != dir && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "attachments" || entry.Name() == "issues" || strings.HasSuffix(entry.Name(), ".attachments")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || filepath.Ext(name) != ".md" {
			return nil
		}
		saved, err := readSyncDocument(name)
		if err != nil {
			return err
		}
		source := saved.Metadata.Source
		if source.Provider != "jira" || (site != "" && source.Site != site) || (source.Key != key && source.ID != key) {
			return nil
		}
		if doc != nil {
			return fmt.Errorf("ambiguous saved Jira issue %s: %s and %s", key, path, name)
		}
		path, doc = name, saved
		return nil
	})
	return path, doc, err
}

func readSyncDocument(path string) (*syncdoc.Document, error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return syncdoc.Read(root, filepath.Base(path))
}
