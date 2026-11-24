// Tests that gavel's query log treats a query whose caller went away as no failure, keeps every
// other failure, and is the log gavel's database is opened with.
package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// recordedLog records the errors queries were traced with and the levels it was set to.
type recordedLog struct {
	errs   *[]error
	levels *[]gormlogger.LogLevel
}

func (r recordedLog) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	*r.levels = append(*r.levels, level)
	return r
}
func (recordedLog) Info(context.Context, string, ...any)  {}
func (recordedLog) Warn(context.Context, string, ...any)  {}
func (recordedLog) Error(context.Context, string, ...any) {}
func (r recordedLog) Trace(_ context.Context, _ time.Time, _ func() (string, int64), err error) {
	*r.errs = append(*r.errs, err)
}

var _ = Describe("Query log", func() {
	var (
		errs   []error
		levels []gormlogger.LogLevel
		inner  recordedLog
	)

	BeforeEach(func() {
		errs, levels = nil, nil
		inner = recordedLog{errs: &errs, levels: &levels}
	})

	trace := func(log gormlogger.Interface, err error) {
		log.Trace(context.Background(), time.Now(), func() (string, int64) { return "SELECT 1", 1 }, err)
	}

	It("traces a query whose caller went away as no failure", func() {
		trace(queryLog{inner}, fmt.Errorf("read: %w", context.Canceled))

		Expect(errs).To(Equal([]error{nil}))
	})

	It("keeps every other failure, a timeout too", func() {
		broken := errors.New("relation does not exist")
		trace(queryLog{inner}, broken)
		trace(queryLog{inner}, context.DeadlineExceeded)

		Expect(errs).To(Equal([]error{broken, context.DeadlineExceeded}))
	})

	It("stays itself when its level changes", func() {
		quiet := queryLog{inner}.LogMode(gormlogger.Silent)
		trace(quiet, context.Canceled)

		Expect(levels).To(Equal([]gormlogger.LogLevel{gormlogger.Silent}))
		Expect(errs).To(Equal([]error{nil}))
	})

	It("is the log gavel's database is opened with", func(ctx SpecContext) {
		var config *gorm.Config
		deps := openDependencies{
			disabled: func() (string, bool) { return "", false },
			resolve:  func() (string, string, error) { return "postgres://gavel", "test", nil },
			open: func(_ string, c *gorm.Config) (*gorm.DB, error) {
				config = c
				return &gorm.DB{}, nil
			},
		}

		_, err := open(ctx, deps)

		Expect(err).NotTo(HaveOccurred())
		Expect(config.Logger).To(BeAssignableToTypeOf(queryLog{}))
	})
})
