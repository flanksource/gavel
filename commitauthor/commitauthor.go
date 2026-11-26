// Package commitauthor is who a hosted run's commits are by: the user who started the run, as
// GitHub knows them, unless they set their own name or email.
package commitauthor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrInvalid is a name or email git would misread.
var ErrInvalid = errors.New("invalid commit identity")

// Author is a commit's author.
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Identity is a user's commit identity: the name and email they set, empty when they set none,
// and the default each empty one falls back to.
type Identity struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Default Author `json:"default"`
}

// Author is who the user's runs commit as: each field they set, else its default. ok is false
// when no email is known, and gavel is the author instead; without a name, the email names them.
func (i Identity) Author() (Author, bool) {
	author := Author{Name: firstNonEmpty(i.Name, i.Default.Name), Email: firstNonEmpty(i.Email, i.Default.Email)}
	if author.Email == "" {
		return Author{}, false
	}
	if author.Name == "" {
		author.Name = author.Email
	}
	return author, true
}

// Store reads and keeps users' commit identities.
type Store struct{ DB *gorm.DB }

// Get reads user's commit identity.
func (s Store) Get(ctx context.Context, user uuid.UUID) (Identity, error) {
	var row struct {
		Name, Email             string
		GitHubLogin             *string `gorm:"column:github_login"`
		GitHubID                *int64  `gorm:"column:github_id"`
		CommitName, CommitEmail *string
	}
	result := s.DB.WithContext(ctx).Raw(`SELECT name, email, github_login, github_id, commit_name, commit_email FROM users WHERE id = ?`, user).Scan(&row)
	if result.Error != nil {
		return Identity{}, fmt.Errorf("read user %s's commit identity: %w", user, result.Error)
	}
	if result.RowsAffected == 0 {
		return Identity{}, fmt.Errorf("read user %s's commit identity: no such user", user)
	}
	identity := Identity{Name: deref(row.CommitName), Email: deref(row.CommitEmail), Default: Author{Name: row.Name, Email: row.Email}}
	if login := deref(row.GitHubLogin); login != "" {
		identity.Default = Author{Name: login, Email: login + "@users.noreply.github.com"}
		if row.GitHubID != nil && *row.GitHubID > 0 {
			identity.Default.Email = fmt.Sprintf("%d+%s@users.noreply.github.com", *row.GitHubID, login)
		}
	}
	return identity, nil
}

// Set keeps name and email as user's commit identity, trimmed; an empty one falls back to its
// default.
func (s Store) Set(ctx context.Context, user uuid.UUID, name, email string) (Identity, error) {
	name, email = strings.TrimSpace(name), strings.TrimSpace(email)
	if err := validate(name, email); err != nil {
		return Identity{}, err
	}
	if err := s.DB.WithContext(ctx).Exec(`UPDATE users SET commit_name = NULLIF(?, ''), commit_email = NULLIF(?, ''), updated_at = now() WHERE id = ?`,
		name, email, user).Error; err != nil {
		return Identity{}, fmt.Errorf("save user %s's commit identity: %w", user, err)
	}
	return s.Get(ctx, user)
}

// validate refuses what git would misread in "name <email>".
func validate(name, email string) error {
	if strings.ContainsAny(name, "<>") {
		return fmt.Errorf(`%w: the name can't contain "<" or ">"`, ErrInvalid)
	}
	if strings.ContainsAny(name, "\r\n") {
		return fmt.Errorf("%w: the name must be one line", ErrInvalid)
	}
	if email == "" {
		return nil
	}
	local, domain, ok := strings.Cut(email, "@")
	if !ok || local == "" || domain == "" || strings.Contains(domain, "@") || strings.ContainsAny(email, "<> \t\r\n") {
		return fmt.Errorf("%w: the email must be one address, like ada@example.com", ErrInvalid)
	}
	return nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
