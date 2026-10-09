package runtime

import (
	"sync"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

// statementRecorder records every statement a GORM pool executes while it is
// recording, whether or not the statement's context carries a Server-Timing
// accumulator — so, unlike the production "sql" metric, a query issued without
// a context still counts.
type statementRecorder struct {
	recording atomic.Bool
	mu        sync.Mutex
	sql       []string
}

func recordStatements(db *gorm.DB) *statementRecorder {
	GinkgoHelper()
	recorder := &statementRecorder{}
	callbacks := db.Callback()
	for name, processor := range map[string]interface {
		Register(string, func(*gorm.DB)) error
	}{
		"create": callbacks.Create().After("gorm:create"),
		"query":  callbacks.Query().After("gorm:query"),
		"update": callbacks.Update().After("gorm:update"),
		"delete": callbacks.Delete().After("gorm:delete"),
		"row":    callbacks.Row().After("gorm:row"),
		"raw":    callbacks.Raw().After("gorm:raw"),
	} {
		Expect(processor.Register("test:statement-recorder:"+name, recorder.record)).To(Succeed())
	}
	return recorder
}

func (r *statementRecorder) record(tx *gorm.DB) {
	if !r.recording.Load() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sql = append(r.sql, tx.Statement.SQL.String())
}

// during returns the statements fn executed.
func (r *statementRecorder) during(fn func()) []string {
	GinkgoHelper()
	r.mu.Lock()
	r.sql = nil
	r.mu.Unlock()
	r.recording.Store(true)
	fn()
	r.recording.Store(false)
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.sql...)
}
