package fixtures

import (
	"context"
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var _ = Describe("Fixture YAML source diagnostics", func() {
	It("stops the runner before executing any step", func() {
		calls := 0
		useKind(strictKind{fakeKind{name: "probe", execute: func(_ context.Context, _ *StepContext) (StepOutcome, error) {
			calls++
			return StepOutcome{}, nil
		}}})
		path := writeDocument("invalid.fixture.yaml", "kind: Fixture\nsteps:\n  - {kind: probe, url: ok}\n  - {id: scaleDown, kind: probe, url: ok, cleanup: true}\n")
		runner, err := NewRunner(RunnerOptions{Paths: []string{path}, WorkDir: filepath.Dir(path)})
		Expect(err).NotTo(HaveOccurred())
		_, err = runner.Run()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).NotTo(ContainSubstring("failed to parse fixture"))
		Expect(calls).To(BeZero())
	})
	It("points at cleanup in scaleDown in the second document and retains the schema cause", func() {
		useKind(strictKind{fakeKind{name: "probe", execute: passing(nil, nil)}})
		path := writeDocument("purge.fixture.yaml", "kind: Fixture\nsteps:\n  - {id: first, kind: probe, url: ok}\n---\nkind: Fixture\nsteps:\n  - id: scaleDown\n    kind: probe\n    url: ok\n    cleanup: true\n")
		_, err := ParseFixtureFile(path)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(path + ":10:5"))
		Expect(err.Error()).To(ContainSubstring(`step "scaleDown"`))
		Expect(err.Error()).To(ContainSubstring("> 10 |     cleanup: true"))
		var invalid *jsonschema.ValidationError
		Expect(errors.As(err, &invalid)).To(BeTrue())
	})
})
