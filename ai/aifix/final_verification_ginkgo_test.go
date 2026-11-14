package aifix

import (
	"context"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/linters"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRepairVerification(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Repair verification")
}

var _ = Describe("Final repair verification", func() {
	It("rechecks edits from the final allowed iteration", func() {
		provider := &fakeStreaming{model: "gpt-5.5", runtime: fakeRuntime}
		captainai.RegisterProvider(fakeRuntime, func(captainai.Config) (captainai.Provider, error) { return provider, nil })
		checks := 0
		result, err := Run(context.Background(), Request{
			Initial:        resultsWith("fakelint", violation("sample.go", "missing comma", "syntax", 1)),
			MaxIterations:  1,
			AIConfig:       captainai.Config{Model: api.Model{Name: "gpt-5.5", Mode: fakeRuntime.Mode}, APIKey: "example-key"},
			AIRequestProto: captainai.Request{Prompt: api.Prompt{User: "Repair the syntax error"}},
			BuildRequest:   unexpectedRequestBuild,
			ReLint:         func(context.Context) ([]*linters.LinterResult, error) { checks++; return nil, nil },
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(checks).To(Equal(1))
		Expect(result.FinalResults).To(BeEmpty())
	})
})
