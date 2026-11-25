package testrunner

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

// TestMain lets this test binary serve as the fixture worker. Run executes
// discovered fixtures in an isolated worker process spawned from
// os.Executable(), which under `go test` is this binary; without the dispatch
// the worker would rerun the whole suite instead of answering, and hang.
func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == fixtures.WorkerCommand {
		wio, err := fixtures.WorkerProcessIO()
		if err == nil {
			err = fixtures.RunWorker(wio)
		}
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestTestRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Runner Suite")
}
