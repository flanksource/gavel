package cache

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LinterExecution struct {
	ID             int64 `gorm:"primaryKey;autoIncrement"`
	LinterName     string
	WorkDir        string
	ExecutedAt     time.Time
	DurationMS     int64
	ViolationCount int
	Success        bool
}

func (LinterExecution) TableName() string { return "linter_executions" }

type DebounceMetadata struct {
	LinterName              string `gorm:"primaryKey"`
	WorkDir                 string `gorm:"primaryKey"`
	LastDebounceUsedMS      *int64
	ConsecutiveNoViolations int
	ConsecutiveViolations   int
	AdaptationFactor        float64
	UpdatedAt               time.Time
}

func (DebounceMetadata) TableName() string { return "debounce_metadata" }

// LinterStats tracks execution metrics for intelligent debouncing.
type LinterStats struct{ db *gorm.DB }

type ExecutionStats struct {
	LinterName     string
	WorkDir        string
	LastRun        time.Time
	LastDuration   time.Duration
	RunCount       int64
	AvgDuration    time.Duration
	ViolationCount int64
	SuccessRate    float64
}

func NewLinterStats(db *gorm.DB) (*LinterStats, error) {
	if db == nil {
		return nil, fmt.Errorf("linter stats database is nil")
	}
	return &LinterStats{db: db}, nil
}

func (ls *LinterStats) RecordExecution(linterName, workDir string, duration time.Duration, violations int, success bool) error {
	now := time.Now()
	return ls.db.Transaction(func(tx *gorm.DB) error {
		execution := LinterExecution{
			LinterName:     linterName,
			WorkDir:        workDir,
			ExecutedAt:     now,
			DurationMS:     duration.Milliseconds(),
			ViolationCount: violations,
			Success:        success,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "linter_name"}, {Name: "work_dir"}, {Name: "executed_at"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"duration_ms", "violation_count", "success",
			}),
		}).Create(&execution).Error; err != nil {
			return err
		}
		return updateDebounceMetadata(tx, linterName, workDir, violations, now)
	})
}

func updateDebounceMetadata(tx *gorm.DB, linterName, workDir string, violations int, now time.Time) error {
	metadata := DebounceMetadata{LinterName: linterName, WorkDir: workDir, AdaptationFactor: 1}
	err := tx.Where("linter_name = ? AND work_dir = ?", linterName, workDir).First(&metadata).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if metadata.AdaptationFactor == 0 {
		metadata.AdaptationFactor = 1
	}

	if violations == 0 {
		metadata.ConsecutiveNoViolations++
		metadata.ConsecutiveViolations = 0
	} else {
		metadata.ConsecutiveNoViolations = 0
		metadata.ConsecutiveViolations++
	}
	if metadata.ConsecutiveNoViolations >= 5 {
		metadata.AdaptationFactor = minFloat(2, metadata.AdaptationFactor*1.1)
	} else if metadata.ConsecutiveViolations >= 3 {
		metadata.AdaptationFactor = maxFloat(0.5, metadata.AdaptationFactor*0.9)
	}
	metadata.UpdatedAt = now

	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "linter_name"}, {Name: "work_dir"}},
		DoUpdates: clause.AssignmentColumns([]string{"consecutive_no_violations", "consecutive_violations", "adaptation_factor", "updated_at"}),
	}).Create(&metadata).Error
}

func (ls *LinterStats) GetIntelligentDebounce(linterName, workDir string) (time.Duration, error) {
	var aggregate struct {
		AvgDuration float64
		RunCount    int64
	}
	err := ls.db.Model(&LinterExecution{}).
		Select("COALESCE(AVG(duration_ms), 1000) AS avg_duration, COUNT(*) AS run_count").
		Where("linter_name = ? AND work_dir = ? AND executed_at > ?", linterName, workDir, time.Now().Add(-7*24*time.Hour)).
		Scan(&aggregate).Error
	if err != nil {
		return 5 * time.Minute, nil
	}

	adaptationFactor := 1.0
	var metadata DebounceMetadata
	if err := ls.db.Where("linter_name = ? AND work_dir = ?", linterName, workDir).First(&metadata).Error; err == nil && metadata.AdaptationFactor != 0 {
		adaptationFactor = metadata.AdaptationFactor
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	avg := time.Duration(aggregate.AvgDuration) * time.Millisecond
	var debounce time.Duration
	switch {
	case aggregate.RunCount == 0:
		debounce = 5 * time.Minute
	case avg < 100*time.Millisecond:
		debounce = 0
	case avg < time.Second:
		debounce = 5 * time.Second
	case avg < 30*time.Second:
		debounce = 5 * time.Minute
	case avg < 5*time.Minute:
		debounce = time.Hour
	case avg < 15*time.Minute:
		debounce = 3 * time.Hour
	default:
		debounce = 8 * time.Hour
	}
	if debounce > 0 {
		debounce = time.Duration(float64(debounce) * adaptationFactor)
	}
	if debounce > 24*time.Hour {
		debounce = 24 * time.Hour
	}
	return debounce, nil
}

func (ls *LinterStats) ShouldSkipLinter(linterName, workDir, configured string) (bool, time.Duration, error) {
	var debounce time.Duration
	var err error
	if configured != "" && configured != "auto" {
		debounce, err = time.ParseDuration(configured)
		if err != nil {
			return false, 0, fmt.Errorf("invalid debounce duration: %w", err)
		}
	} else {
		debounce, err = ls.GetIntelligentDebounce(linterName, workDir)
		if err != nil {
			return false, 0, err
		}
	}

	var result struct{ LastRun *time.Time }
	if err := ls.db.Model(&LinterExecution{}).
		Select("MAX(executed_at) AS last_run").
		Where("linter_name = ? AND work_dir = ?", linterName, workDir).
		Scan(&result).Error; err != nil {
		return false, debounce, err
	}
	if result.LastRun == nil {
		return false, debounce, nil
	}
	return time.Since(*result.LastRun) < debounce, debounce, nil
}

func (ls *LinterStats) GetStats(linterName, workDir string) (*ExecutionStats, error) {
	stats := &ExecutionStats{LinterName: linterName, WorkDir: workDir}
	var aggregate struct {
		RunCount      int64
		AvgDurationMS float64
		Violations    int64
		SuccessRate   float64
	}
	if err := ls.db.Model(&LinterExecution{}).
		Select(`COUNT(*) AS run_count,
			COALESCE(AVG(duration_ms), 0) AS avg_duration_ms,
			COALESCE(SUM(violation_count), 0) AS violations,
			COALESCE(AVG(CASE WHEN success THEN 1.0 ELSE 0.0 END), 0) AS success_rate`).
		Where("linter_name = ? AND work_dir = ?", linterName, workDir).
		Scan(&aggregate).Error; err != nil {
		return nil, err
	}
	stats.RunCount = aggregate.RunCount
	stats.AvgDuration = time.Duration(aggregate.AvgDurationMS) * time.Millisecond
	stats.ViolationCount = aggregate.Violations
	stats.SuccessRate = aggregate.SuccessRate

	if stats.RunCount > 0 {
		var latest LinterExecution
		if err := ls.db.Where("linter_name = ? AND work_dir = ?", linterName, workDir).
			Order("executed_at DESC").First(&latest).Error; err != nil {
			return nil, err
		}
		stats.LastRun = latest.ExecutedAt
		stats.LastDuration = time.Duration(latest.DurationMS) * time.Millisecond
	}
	return stats, nil
}

func (ls *LinterStats) GetLinterHistory(workDir string) ([]string, error) {
	var linters []string
	if err := ls.db.Model(&LinterExecution{}).
		Distinct("linter_name").
		Where("work_dir = ?", workDir).
		Order("linter_name").
		Pluck("linter_name", &linters).Error; err != nil {
		return nil, err
	}
	return linters, nil
}

func (ls *LinterStats) Close() error { return nil }

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
