package todos

import (
	"context"
	"sync"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type recordingProgressProvider struct {
	mu        sync.Mutex
	snapshots []fixtures.ExecutionSnapshot
}

func (p *recordingProgressProvider) RecordRunProgress(_ context.Context, _ *types.TODO, snapshot fixtures.ExecutionSnapshot) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.snapshots = append(p.snapshots, snapshot)
	return nil
}

func (p *recordingProgressProvider) recorded() []fixtures.ExecutionSnapshot {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]fixtures.ExecutionSnapshot(nil), p.snapshots...)
}

var _ = Describe("verification progress persistence", func() {
	It("writes the first snapshot immediately and flushes the latest snapshot on close", func() {
		provider := &recordingProgressProvider{}
		persister := newProgressPersister(context.Background(), provider, &types.TODO{ID: "progress-todo"})

		for iteration := int64(1); iteration <= 3; iteration++ {
			Expect(persister.Sink(context.Background(), fixtures.ExecutionSnapshot{Iteration: iteration})).To(Succeed())
		}
		Expect(provider.recorded()).To(HaveLen(1))

		Expect(persister.Close()).To(Succeed())
		Expect(provider.recorded()).To(HaveLen(2))
		Expect(provider.recorded()[1].Iteration).To(Equal(int64(3)))
	})
})
