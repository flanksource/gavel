// gavel's query log: the shared SQL log, except that a query canceled because its caller went
// away, such as a browser dropping a request, is traced as a query, not logged as a failure.
package database

import (
	"context"
	"errors"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

// queryLog passes every query on to its log, a canceled one without its error.
type queryLog struct {
	gormlogger.Interface
}

func (l queryLog) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return queryLog{l.Interface.LogMode(level)}
}

func (l queryLog) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	l.Interface.Trace(ctx, begin, fc, err)
}
