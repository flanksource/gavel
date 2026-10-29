package jira

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

const catalogTTL = 24 * time.Hour

type CatalogScope struct {
	Site    string `json:"site"`
	Account string `json:"account"`
}

type CatalogStore struct {
	Dir string
	Now func() time.Time
}

type fieldCatalog struct {
	Version   int                        `json:"version"`
	Scope     CatalogScope               `json:"scope"`
	Fields    map[string]FieldDefinition `json:"fields"`
	FetchedAt time.Time                  `json:"fetchedAt"`
}

func (s CatalogStore) Resolve(ctx context.Context, scope CatalogScope, refresh bool, fetch func(context.Context) ([]FieldDefinition, error)) (result map[string]FieldDefinition, returnErr error) {
	path, lock, err := s.lock(ctx, scope)
	if err != nil {
		return nil, err
	}
	defer func() { returnErr = errors.Join(returnErr, lock.Unlock()) }()
	cache := fieldCatalog{Version: 2, Scope: scope, Fields: map[string]FieldDefinition{}}
	data, err := os.ReadFile(path)
	if err == nil && !refresh {
		if err := json.Unmarshal(data, &cache); err != nil {
			return nil, fmt.Errorf("decode Jira catalog %s: %w", path, err)
		}
		if cache.Version != 2 || cache.Scope != scope || cache.Fields == nil {
			return nil, fmt.Errorf("invalid Jira catalog %s", path)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read Jira catalog %s: %w", path, err)
	}
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	if !refresh && !cache.FetchedAt.IsZero() && !now.Before(cache.FetchedAt) && now.Sub(cache.FetchedAt) < catalogTTL {
		return cache.Fields, nil
	}
	definitions, err := fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch Jira field definitions: %w", err)
	}
	cache.Fields = make(map[string]FieldDefinition, len(definitions))
	for _, field := range definitions {
		if strings.TrimSpace(field.ID) == "" || strings.TrimSpace(field.Name) == "" {
			return nil, errors.New("invalid Jira field definition: id and name are required")
		}
		if _, exists := cache.Fields[field.ID]; exists {
			return nil, fmt.Errorf("duplicate Jira field definition %s", field.ID)
		}
		cache.Fields[field.ID] = FieldDefinition{ID: field.ID, Name: field.Name, Schema: field.Schema}
	}
	cache.FetchedAt = now
	data, err = json.Marshal(cache)
	if err != nil {
		return nil, fmt.Errorf("encode Jira catalog: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".field-catalog-*")
	if err != nil {
		return nil, fmt.Errorf("create Jira catalog file: %w", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil && !os.IsNotExist(err) {
			returnErr = errors.Join(returnErr, fmt.Errorf("remove temporary Jira catalog: %w", err))
		}
	}()
	if _, err := file.Write(data); err != nil {
		return nil, errors.Join(fmt.Errorf("write Jira catalog: %w", err), file.Close())
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close Jira catalog: %w", err)
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return nil, fmt.Errorf("replace Jira catalog: %w", err)
	}
	return cache.Fields, nil
}

func (s CatalogStore) lock(ctx context.Context, scope CatalogScope) (string, *flock.Flock, error) {
	if strings.TrimSpace(scope.Site) == "" || strings.TrimSpace(scope.Account) == "" {
		return "", nil, errors.New("Jira field catalog requires site and account")
	}
	dir := s.Dir
	if dir == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return "", nil, fmt.Errorf("resolve Jira cache directory: %w", err)
		}
		dir = filepath.Join(cacheDir, "gavel", "jira")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", nil, fmt.Errorf("create Jira catalog directory: %w", err)
	}
	key, err := json.Marshal(scope)
	if err != nil {
		return "", nil, fmt.Errorf("encode Jira catalog scope: %w", err)
	}
	hash := sha256.Sum256(key)
	path := filepath.Join(dir, "fields-v2-"+hex.EncodeToString(hash[:])+".json")
	lock := flock.New(path + ".lock")
	locked, err := lock.TryLockContext(ctx, 50*time.Millisecond)
	if err != nil {
		return "", nil, fmt.Errorf("lock Jira catalog: %w", err)
	}
	if !locked {
		return "", nil, errors.New("lock Jira catalog: lock not acquired")
	}
	return path, lock, nil
}
