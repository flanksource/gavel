package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/gavel/models"
	"gorm.io/gorm"
)

// FileScan records the content fingerprint used to invalidate cached findings.
type FileScan struct {
	FilePath     string `gorm:"primaryKey;column:file_path"`
	LastScanTime int64  `gorm:"column:last_scan_time;not null"`
	FileModTime  int64  `gorm:"column:file_mod_time;not null"`
	FileHash     string `gorm:"column:file_hash;not null"`
}

func (FileScan) TableName() string { return "file_scans" }

// violationRecord keeps persistence concerns out of the public violation
// model. In particular, Rule is encoded explicitly as JSONB.
type violationRecord struct {
	ID               int64 `gorm:"primaryKey;autoIncrement"`
	FilePath         string
	Line             int
	Column           int
	Message          *string
	Source           string
	Rule             []byte `gorm:"type:jsonb"`
	Severity         string
	Fixable          bool
	FixApplicability string
	Code             *string
	CreatedAt        time.Time
}

func (violationRecord) TableName() string { return "violations" }

// ViolationCache manages cached linter findings in PostgreSQL.
type ViolationCache struct{ db *gorm.DB }

func NewViolationCache(db *gorm.DB) (*ViolationCache, error) {
	if db == nil {
		return nil, fmt.Errorf("violation cache database is nil")
	}
	return &ViolationCache{db: db}, nil
}

func GetFileHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

func (c *ViolationCache) NeedsRescan(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return true, nil
	}

	var scan FileScan
	err = c.db.Where("file_path = ?", filePath).First(&scan).Error
	if err == gorm.ErrRecordNotFound {
		return true, nil
	}
	if err != nil {
		return true, err
	}
	if info.ModTime().Unix() > scan.FileModTime {
		return true, nil
	}
	currentHash, err := GetFileHash(filePath)
	if err != nil {
		return true, err
	}
	return currentHash != scan.FileHash, nil
}

func (c *ViolationCache) GetCachedViolations(filePath string) ([]models.Violation, error) {
	return c.find(c.db.Where("file_path = ?", filePath))
}

func (c *ViolationCache) GetAllViolations() ([]models.Violation, error) {
	return c.find(c.db.Order("file_path, line, column"))
}

func (c *ViolationCache) GetViolationsBySource(source string) ([]models.Violation, error) {
	return c.find(c.db.Where("source = ?", source).Order("file_path, line, column"))
}

func (c *ViolationCache) GetViolationsBySources(sources []string) ([]models.Violation, error) {
	if len(sources) == 0 {
		return []models.Violation{}, nil
	}
	return c.find(c.db.Where("source IN ?", sources).Order("file_path, line, column"))
}

func (c *ViolationCache) find(query *gorm.DB) ([]models.Violation, error) {
	var records []violationRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}
	violations := make([]models.Violation, 0, len(records))
	for _, record := range records {
		violation, err := record.violation()
		if err != nil {
			return nil, err
		}
		violations = append(violations, violation)
	}
	return violations, nil
}

func (r violationRecord) violation() (models.Violation, error) {
	violation := models.Violation{
		File:             r.FilePath,
		Line:             r.Line,
		Column:           r.Column,
		Message:          r.Message,
		Source:           r.Source,
		Severity:         models.ViolationSeverity(r.Severity),
		Fixable:          r.Fixable,
		FixApplicability: r.FixApplicability,
		Code:             r.Code,
		CreatedAt:        r.CreatedAt,
	}
	if len(r.Rule) > 0 {
		var rule models.Rule
		if err := json.Unmarshal(r.Rule, &rule); err != nil {
			return models.Violation{}, fmt.Errorf("decode cached rule for %s:%d: %w", r.FilePath, r.Line, err)
		}
		violation.Rule = &rule
	}
	return violation, nil
}

func newViolationRecord(filePath string, violation models.Violation) (violationRecord, error) {
	var rule []byte
	var err error
	if violation.Rule != nil {
		rule, err = json.Marshal(violation.Rule)
		if err != nil {
			return violationRecord{}, fmt.Errorf("encode rule for %s:%d: %w", filePath, violation.Line, err)
		}
	}
	createdAt := violation.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	return violationRecord{
		FilePath:         filePath,
		Line:             violation.Line,
		Column:           violation.Column,
		Message:          violation.Message,
		Source:           violation.Source,
		Rule:             rule,
		Severity:         string(violation.Severity),
		Fixable:          violation.Fixable,
		FixApplicability: violation.FixApplicability,
		Code:             violation.Code,
		CreatedAt:        createdAt,
	}, nil
}

func (c *ViolationCache) StoreViolations(filePath string, violations []models.Violation) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	hash, err := GetFileHash(filePath)
	if err != nil {
		return err
	}
	records := make([]violationRecord, 0, len(violations))
	for _, violation := range violations {
		record, err := newViolationRecord(filePath, violation)
		if err != nil {
			return err
		}
		records = append(records, record)
	}

	return c.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("file_path = ?", filePath).Delete(&FileScan{}).Error; err != nil {
			return err
		}
		scan := FileScan{
			FilePath:     filePath,
			LastScanTime: time.Now().Unix(),
			FileModTime:  info.ModTime().Unix(),
			FileHash:     hash,
		}
		if err := tx.Create(&scan).Error; err != nil {
			return err
		}
		if len(records) > 0 {
			return tx.Create(&records).Error
		}
		return nil
	})
}

func (c *ViolationCache) GetAllCachedFiles() ([]string, error) {
	var scans []FileScan
	if err := c.db.Select("file_path").Find(&scans).Error; err != nil {
		return nil, err
	}
	files := make([]string, len(scans))
	for i, scan := range scans {
		files[i] = scan.FilePath
	}
	return files, nil
}

func (c *ViolationCache) ClearCache() error {
	return c.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&FileScan{}).Error
}

func (c *ViolationCache) ClearFileCache(filePaths []string) error {
	if len(filePaths) == 0 {
		return nil
	}
	return c.db.Where("file_path IN ?", filePaths).Delete(&FileScan{}).Error
}

func (c *ViolationCache) GetStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{}
	var fileCount, violationCount int64
	if err := c.db.Model(&FileScan{}).Count(&fileCount).Error; err != nil {
		return nil, err
	}
	if err := c.db.Model(&violationRecord{}).Count(&violationCount).Error; err != nil {
		return nil, err
	}
	var size int64
	if err := c.db.Raw(`SELECT pg_total_relation_size('public.file_scans'::regclass) + pg_total_relation_size('public.violations'::regclass)`).Scan(&size).Error; err != nil {
		return nil, err
	}
	stats["cached_files"] = fileCount
	stats["total_violations"] = violationCount
	stats["cache_size_bytes"] = size
	return stats, nil
}

func (c *ViolationCache) ClearViolations(olderThan time.Time, pathPattern string) (int64, error) {
	var files []string
	if pathPattern != "" {
		violations, err := c.GetAllViolations()
		if err != nil {
			return 0, err
		}
		seen := map[string]struct{}{}
		for _, violation := range violations {
			if matchesViolationPath(pathPattern, violation.File) {
				seen[violation.File] = struct{}{}
			}
		}
		for file := range seen {
			files = append(files, file)
		}
		if len(files) == 0 {
			return 0, nil
		}
	}

	var deleted int64
	err := c.db.Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&violationRecord{})
		if !olderThan.IsZero() {
			query = query.Where("created_at < ?", olderThan)
		}
		if len(files) > 0 {
			query = query.Where("file_path IN ?", files)
		}
		if err := query.Count(&deleted).Error; err != nil {
			return err
		}
		if olderThan.IsZero() && len(files) == 0 {
			if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&violationRecord{}).Error; err != nil {
				return err
			}
		} else if err := query.Delete(&violationRecord{}).Error; err != nil {
			return err
		}

		scans := tx.Model(&FileScan{})
		switch {
		case len(files) > 0:
			scans = scans.Where("file_path IN ?", files)
		case !olderThan.IsZero():
			scans = scans.Where("last_scan_time < ?", olderThan.Unix())
		default:
			scans = scans.Session(&gorm.Session{AllowGlobalUpdate: true})
		}
		return scans.Delete(&FileScan{}).Error
	})
	return deleted, err
}

func matchesViolationPath(pattern, file string) bool {
	if matched, err := doublestar.Match(pattern, file); err == nil && matched {
		return true
	}
	if matched, err := doublestar.Match(pattern, filepath.Base(file)); err == nil && matched {
		return true
	}
	if !filepath.IsAbs(pattern) {
		if relative, err := filepath.Rel(filepath.Dir(file), file); err == nil {
			matched, _ := doublestar.Match(pattern, relative)
			return matched
		}
	}
	return false
}
