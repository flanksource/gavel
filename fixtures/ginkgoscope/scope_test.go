package ginkgoscope

import (
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func init() {
	root, err := filepath.Abs("../..")
	if err != nil {
		panic(err)
	}
	runner, err := fixtures.NewRunner(fixtures.RunnerOptions{
		Paths:   []string{filepath.Join(root, "fixtures", "testdata", "kinds", "exec-env.fixture.yaml")},
		WorkDir: root,
	})
	if err != nil {
		panic(err)
	}
	runner.RegisterGinkgoSpecs()
}

func TestGinkgoScope(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fixture Ginkgo Scope Suite")
}
