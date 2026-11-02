package resultstore

import (
	"context"
	"testing"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestResultStores(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Result store adapters") }

var _ = DescribeTable("durable result store contract", func(open func(context.Context, string) (Store, error)) {
	ctx, dir := context.Background(), GinkgoT().TempDir()
	store, err := open(ctx, dir)
	Expect(err).NotTo(HaveOccurred())
	original := store
	DeferCleanup(func() {
		if original != nil {
			Expect(original.Close()).To(Succeed())
		}
	})
	run := Run{Metadata: Metadata{ExecutionID: uuid.NewString(), Workspace: dir, Kind: "test+lint", Started: time.Unix(100, 0)}, Status: Status{State: "queued", Running: true}, Admission: "request-key", InputHash: "input-one"}
	accepted, created, err := store.Create(ctx, run)
	Expect(err).NotTo(HaveOccurred())
	Expect(created).To(BeTrue())
	Expect(accepted.Metadata.Revision).To(Equal(uint64(1)))
	retry := run
	retry.Metadata.ExecutionID = uuid.NewString()
	replayed, created, err := store.Create(ctx, retry)
	Expect(err).NotTo(HaveOccurred())
	Expect(created).To(BeFalse())
	Expect(replayed.Metadata.ExecutionID).To(Equal(accepted.Metadata.ExecutionID))
	retry.InputHash = "different"
	_, _, err = store.Create(ctx, retry)
	Expect(err).To(MatchError(ErrConflict))
	accepted.Tests = []parsers.Test{{Name: "nested attempt", Passed: true}}
	message := "visible finding"
	accepted.Lint = []*linters.LinterResult{{Linter: "prompt", Violations: []models.Violation{{Message: &message, Line: 7}}}}
	accepted.Status = Status{State: "completed", LintRun: true}
	accepted.Metadata.Ended = time.Unix(110, 0)
	finished, err := store.Update(ctx, accepted, 1)
	Expect(err).NotTo(HaveOccurred())
	Expect(finished.Metadata.Revision).To(Equal(uint64(2)))
	_, err = store.Update(ctx, accepted, 1)
	Expect(err).To(MatchError(ErrTerminal))
	Expect(store.Close()).To(Succeed())
	original = nil
	store, err = open(ctx, dir)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(store.Close)
	loaded, err := store.Get(ctx, accepted.Metadata.ExecutionID)
	Expect(err).NotTo(HaveOccurred())
	Expect(loaded).To(Equal(finished))
	Expect(loaded.Public().Admission).To(BeEmpty())
	Expect(loaded.Public().InputHash).To(BeEmpty())
	tests, err := store.Tests(ctx, loaded.Metadata.ExecutionID)
	Expect(err).NotTo(HaveOccurred())
	Expect(tests).To(Equal(finished.Tests))
	findings, err := store.Lint(ctx, loaded.Metadata.ExecutionID)
	Expect(err).NotTo(HaveOccurred())
	Expect(findings).To(Equal(finished.Lint))
	target, err := open(ctx, GinkgoT().TempDir())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(target.Close)
	imported, err := target.Import(ctx, finished)
	Expect(err).NotTo(HaveOccurred())
	Expect(imported).To(BeTrue())
	copy, err := target.Get(ctx, finished.Metadata.ExecutionID)
	Expect(err).NotTo(HaveOccurred())
	Expect(copy).To(Equal(finished))
	imported, err = target.Import(ctx, finished)
	Expect(err).NotTo(HaveOccurred())
	Expect(imported).To(BeFalse())
}, Entry("snapshots", func(_ context.Context, dir string) (Store, error) { return OpenSnapshots(dir) }), Entry("SQLite", OpenSQLite))
