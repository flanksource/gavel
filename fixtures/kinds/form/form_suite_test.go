package form_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

// TestMain lets this test binary serve as the fixture worker an isolated run
// spawns from os.Executable().
func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == fixtures.WorkerCommand {
		wio, err := fixtures.WorkerProcessIO()
		if err == nil {
			err = fixtures.RunWorker(wio)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestForm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Form Kind Suite")
}
