package bulk

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBulkRun(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bulk Run Suite")
}

var _ = Describe("Bulk runtime preset transport", func() {
	It("forwards generated preset flags without making them spec fields", func() {
		flags, err := clicky.BuildOpts[RunFlags](map[string]string{"preset": "organization,review"})
		Expect(err).NotTo(HaveOccurred())
		opts, err := DefaultRunResolver(context.Background(), RunRequest{Step: "plan", Flags: flags})
		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(Equal(run.Options{Step: "plan", Presets: []string{"organization", "review"}, PresetsSet: true, Host: lifecycle.HostCLI}))
		Expect(api.IsEmpty(flags.Spec())).To(BeTrue())
	})

	DescribeTable("keeps each TODO's workspace and exact prepared presets through dispatch", func(location string) {
		batchDir, ownedDir := GinkgoT().TempDir(), GinkgoT().TempDir()
		todo, expectedDir := &types.TODO{}, batchDir
		switch location {
		case "absolute":
			todo.CWD, expectedDir = ownedDir, ownedDir
		case "relative":
			todo.CWD = filepath.Join("packages", "worker")
		}
		prepared := &run.Prepared{Resolution: &lifecycle.Resolution{Prompt: "Prepared once"}}
		var optionsDir, previewDir string
		var dispatched run.Request
		oldResolve, oldStart := run.Resolve, run.Start
		DeferCleanup(func() { run.Resolve, run.Start = oldResolve, oldStart })
		run.Resolve = func(_ context.Context, req run.Request) (*run.Prepared, error) {
			previewDir = req.Dir
			return prepared, nil
		}
		run.Start = func(req run.Request) (run.StartResult, error) {
			dispatched = req
			return run.StartResult{Status: "started"}, nil
		}
		resolve := func(ctx context.Context, req RunRequest) (run.Options, error) {
			optionsDir = req.Dir
			return DefaultRunResolver(ctx, req)
		}
		fn, err := StartRun(RunSpec{
			Step: "run", Flags: RunFlags{Presets: []string{"review"}}, Batch: []string{"ab12cd", "ff0011"},
			Registry: run.NewRegistry(), Dir: batchDir, Resolve: resolve,
		})
		Expect(err).NotTo(HaveOccurred())
		_, err = fn(context.Background(), nil, todo)
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{optionsDir, previewDir, dispatched.Dir}).To(Equal([]string{expectedDir, expectedDir, expectedDir}))
		Expect(dispatched.Prepared).To(BeIdenticalTo(prepared))
		Expect(dispatched.Options.Presets).To(Equal([]string{"review"}))
		Expect(dispatched.Options.PresetsSet).To(BeTrue())
		Expect(dispatched.Todo).To(BeIdenticalTo(todo))
		Expect(dispatched.Options.Batch).To(Equal([]string{"ab12cd", "ff0011"}),
			"every run is told the whole selection, so a triage render can mark the batch")
	}, Entry("absolute owning workspace", "absolute"), Entry("relative execution subdirectory", "relative"), Entry("selected batch workspace", "unset"))
})
