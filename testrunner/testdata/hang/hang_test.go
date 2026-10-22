// Package hang is a Go test package that never finishes: the timeout
// tests run it under a short per-package deadline to see its goroutines
// dumped and its process killed.
package hang

import (
	"testing"
	"time"
)

func TestHang(t *testing.T) {
	<-time.After(time.Hour)
}
