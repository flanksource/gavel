package services_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/onsi/gomega/gexec"

	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/kinds/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var svcBin string

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services Kind Suite")
}

var _ = BeforeSuite(func() {
	var err error
	svcBin, err = gexec.Build("github.com/flanksource/gavel/fixtures/kinds/services/testdata/svc")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() { gexec.CleanupBuildArtifacts() })

// newStep builds the StepContext a `kind: services` step with the given id and
// JSON config runs under, with a fresh fixture scope and tmp dir.
func newStep(id, config string) *fixtures.StepContext {
	return &fixtures.StepContext{
		Spec:    fixtures.StepSpec{ID: id, Kind: "services", Config: json.RawMessage(config)},
		Scope:   fixtures.NewStepScope(os.Environ()),
		TmpDir:  GinkgoT().TempDir(),
		WorkDir: GinkgoT().TempDir(),
	}
}

func servicesKind() fixtures.Kind {
	kind, err := fixtures.DefaultKinds.Lookup("services")
	Expect(err).NotTo(HaveOccurred())
	return kind
}

// execute runs the step and registers its cleanup with Ginkgo so a failing
// spec never leaks a service process.
func execute(step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	outcome, err := servicesKind().Execute(context.Background(), step)
	if err == nil && outcome.Cleanup != nil {
		cleaned := false
		DeferCleanup(func() {
			if !cleaned {
				Expect(outcome.Cleanup(context.Background())).To(Succeed())
			}
		})
		cleanup := outcome.Cleanup
		outcome.Cleanup = func(ctx context.Context) error {
			cleaned = true
			return cleanup(ctx)
		}
	}
	return outcome, err
}

// envValue reads a fixture Env variable, failing the spec when it is unset.
func envValue(step *fixtures.StepContext, name string) string {
	value, ok := step.Scope.Env.Get(name)
	Expect(ok).To(BeTrue(), "%s is not set in the fixture Env", name)
	return value
}

func entryOutputs(outcome fixtures.StepOutcome, name string) map[string]any {
	services, ok := outcome.Outputs["services"].(map[string]any)
	Expect(ok).To(BeTrue(), "outputs.services is %T", outcome.Outputs["services"])
	entry, ok := services[name].(map[string]any)
	Expect(ok).To(BeTrue(), "outputs.services.%s is %T", name, services[name])
	return entry
}
