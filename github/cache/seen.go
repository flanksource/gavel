package cache

import (
	"fmt"
	"time"

	"gorm.io/gorm/clause"
)

// SeenPR records when the user last acknowledged a PR. A PR is considered
// "unread" when the PR's own UpdatedAt is newer than this SeenAt, or when no
// row exists at all. The primary key is (Repo, Number) so each PR has at most
// one row regardless of how many times it has been clicked.
type SeenPR struct {
	Repo      string    `gorm:"primaryKey;size:255"`
	Number    int       `gorm:"primaryKey"`
	SeenAt    time.Time `gorm:"not null"`
	UpdatedAt time.Time
}

// SeenKey identifies a PR in the seen tracking API without dragging in the
// full PRListItem type.
type SeenKey struct {
	Repo   string
	Number int
}

// MarkSeen upserts a (Repo, Number) row with SeenAt = now. A disabled store
// returns nil without touching the database, so callers can ignore the error
// when persistence isn't configured.
func (s *Store) MarkSeen(repo string, number int) error {
	if s.Disabled() {
		return nil
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	row := SeenPR{Repo: repo, Number: number, SeenAt: time.Now()}
	return s.gorm().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "repo"}, {Name: "number"}},
		DoUpdates: clause.AssignmentColumns([]string{"seen_at"}),
	}).Create(&row).Error
}

// MarkManySeen upserts SeenAt = now for every key in one statement. Used by
// "mark all as read" actions. Disabled stores are a silent no-op.
func (s *Store) MarkManySeen(keys []SeenKey) error {
	if s.Disabled() || len(keys) == 0 {
		return nil
	}
	now := time.Now()
	rows := make([]SeenPR, len(keys))
	for i, k := range keys {
		rows[i] = SeenPR{Repo: k.Repo, Number: k.Number, SeenAt: now}
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.gorm().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "repo"}, {Name: "number"}},
		DoUpdates: clause.AssignmentColumns([]string{"seen_at"}),
	}).Create(&rows).Error
}

// LoadSeenMap fetches SeenAt timestamps for the given keys. A disabled store
// returns an empty map (every PR will read as unread). Keys not present in
// the database are simply omitted from the returned map.
func (s *Store) LoadSeenMap(keys []SeenKey) (map[SeenKey]time.Time, error) {
	out := map[SeenKey]time.Time{}
	if s.Disabled() || len(keys) == 0 {
		return out, nil
	}

	repos := make([]string, 0, len(keys))
	numbers := make([]int, 0, len(keys))
	seenSet := make(map[SeenKey]struct{}, len(keys))
	for _, k := range keys {
		if _, ok := seenSet[k]; ok {
			continue
		}
		seenSet[k] = struct{}{}
		repos = append(repos, k.Repo)
		numbers = append(numbers, k.Number)
	}

	var rows []SeenPR
	// Scope by repo IN (...) AND number IN (...) for a single round trip.
	// This overfetches on cross-repo number collisions (#42 in two repos)
	// which we filter out below; that's cheaper than N individual queries.
	err := s.gorm().
		Where("repo IN ? AND number IN ?", repos, numbers).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("load seen map: %w", err)
	}
	for _, r := range rows {
		k := SeenKey{Repo: r.Repo, Number: r.Number}
		if _, ok := seenSet[k]; ok {
			out[k] = r.SeenAt
		}
	}
	return out, nil
}
