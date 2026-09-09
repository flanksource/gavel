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

var _ = Describe("Bulk runtime profile transport", func() {
	It("forwards the generated runtime-profile flag without making it a spec field", func() {
		flags, err := clicky.BuildOpts[RunFlags](map[string]string{"runtime-profile": "Review profile"})
		Expect(err).NotTo(HaveOccurred())
		opts, err := DefaultRunResolver(context.Background(), RunRequest{Step: "plan", Flags: flags})
		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(Equal(run.Options{Step: "plan", RuntimeProfile: "Review profile", Host: lifecycle.HostCLI}))
		Expect(api.IsEmpty(flags.Spec())).To(BeTrue())
	})

	DescribeTable("keeps each TODO's workspace and exact prepared profile through dispatch", func(location string) {
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
		fn, err := StartRun("run", RunFlags{RuntimeProfile: "review"}, run.NewRegistry(), batchDir, resolve, nil)
		Expect(err).NotTo(HaveOccurred())
		_, err = fn(context.Background(), nil, todo)
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{optionsDir, previewDir, dispatched.Dir}).To(Equal([]string{expectedDir, expectedDir, expectedDir}))
		Expect(dispatched.Prepared).To(BeIdenticalTo(prepared))
		Expect(dispatched.Options.RuntimeProfile).To(Equal("review"))
		Expect(dispatched.Todo).To(BeIdenticalTo(todo))
	}, Entry("absolute owning workspace", "absolute"), Entry("relative execution subdirectory", "relative"), Entry("selected batch workspace", "unset"))
})
