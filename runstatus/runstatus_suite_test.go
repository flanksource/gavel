package runstatus_test

import (
	"testing"

	"github.com/flanksource/gavel/slots"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRunStatus(t *testing.T) {
	// A gavel running these specs exports its own lease, which Acquire would inherit.
	t.Setenv(slots.HeldEnv, "")
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run status")
}
