// Pre-tenancy GitHub cache models, frozen from main@edc252b, so migration tests can build
// the schema gorm AutoMigrate created before the HCL bundle and organization keys existed.
package database_test

import "time"

type legacyHTTPCacheEntry struct {
	URL          string `gorm:"primaryKey;size:1024"`
	Method       string `gorm:"primaryKey;size:8"`
	ETag         string `gorm:"size:256"`
	LastModified string `gorm:"size:64"`
	StatusCode   int
	Body         []byte
	Headers      []byte
	FetchedAt    time.Time `gorm:"index"`
	ExpiresAt    *time.Time
}

func (legacyHTTPCacheEntry) TableName() string { return "http_cache_entries" }

type legacyWorkflowRunCache struct {
	RunID      int64  `gorm:"primaryKey"`
	Repo       string `gorm:"index;size:255"`
	Status     string `gorm:"size:32"`
	Conclusion string `gorm:"size:32"`
	Payload    []byte
	FetchedAt  time.Time
}

func (legacyWorkflowRunCache) TableName() string { return "workflow_run_caches" }

type legacyJobLogCache struct {
	JobID     int64  `gorm:"primaryKey"`
	Repo      string `gorm:"size:255"`
	LogsGz    []byte // gzip-compressed
	FetchedAt time.Time
}

func (legacyJobLogCache) TableName() string { return "job_log_caches" }

type legacyWorkflowDefCache struct {
	Repo       string `gorm:"primaryKey;size:255"`
	WorkflowID int64  `gorm:"primaryKey"`
	SHA        string `gorm:"primaryKey;size:64"`
	Path       string `gorm:"size:512"`
	YAML       string
	FetchedAt  time.Time
}

func (legacyWorkflowDefCache) TableName() string { return "workflow_def_caches" }

type legacySeenPR struct {
	Repo      string    `gorm:"primaryKey;size:255"`
	Number    int       `gorm:"primaryKey"`
	SeenAt    time.Time `gorm:"not null"`
	UpdatedAt time.Time
}

func (legacySeenPR) TableName() string { return "seen_prs" }

type legacyTestRunCache struct {
	WorkspaceDir string `gorm:"primaryKey;size:512"`
	RunID        string `gorm:"primaryKey;size:255"` // run-*.json file stem
	Path         string `gorm:"size:1024"`           // absolute path to the run-*.json
	Kind         string `gorm:"size:16;index"`       // test | lint | test+lint
	Repo         string `gorm:"size:255"`
	SHA          string `gorm:"size:64"`
	StartedTS    int64  `gorm:"index"` // unix nanos, the scan watermark
	EndedTS      int64

	Passed  int
	Failed  int
	Skipped int
	Warned  int
	Total   int

	LintViolations int
	LintLinters    int

	Frameworks []byte // JSON-encoded []string
	SyncedAt   time.Time
}

func (legacyTestRunCache) TableName() string { return "test_run_caches" }

type legacyTestRunCursor struct {
	WorkspaceDir string `gorm:"primaryKey;size:512"`
	LastRunTS    int64
	SyncedAt     time.Time
}

func (legacyTestRunCursor) TableName() string { return "test_run_cursors" }
