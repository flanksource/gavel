package todos

import (
	"context"
	"sync"

	capapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type recordingProgressProvider struct {
	mu      sync.Mutex
	reports []capapi.VerifyReport
}

func (p *recordingProgressProvider) RecordRunProgress(_ context.Context, _ *types.TODO, report capapi.VerifyReport) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reports = append(p.reports, report)
	return nil
}

func (p *recordingProgressProvider) recorded() []capapi.VerifyReport {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]capapi.VerifyReport(nil), p.reports...)
}

var _ = Describe("verification progress persistence", func() {
	It("writes the first report immediately and flushes the latest report on close", func() {
		provider := &recordingProgressProvider{}
		persister := newProgressPersister(context.Background(), provider, &types.TODO{ID: "progress-todo"})

		for iteration := 1; iteration <= 3; iteration++ {
			Expect(persister.Sink(capapi.VerifyReport{Kind: "fixture", Iteration: iteration})).To(Succeed())
		}
		Expect(provider.recorded()).To(HaveLen(1))

		Expect(persister.Close()).To(Succeed())
		Expect(provider.recorded()).To(HaveLen(2))
		Expect(provider.recorded()[1].Iteration).To(Equal(3))
	})
})
