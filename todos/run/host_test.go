// Tests the host a run resolves through: a local run's, or, for a run whose prompt runs
// elsewhere, a hosted repository's with the request's RemoteRunner.
package run_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type idleRunner struct{}

func (idleRunner) Run(context.Context, lifecycle.RemoteRun) (promptrun.Result, error) {
	return promptrun.Result{}, nil
}

var _ = Describe("the host a run resolves through", func() {
	It("is a hosted repository's, running its prompts elsewhere, when the request has a RemoteRunner", func() {
		home, repo := GinkgoT().TempDir(), GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		Expect(os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("ssh:\n  cmd: make deploy-from-the-host\n"), 0o644)).To(Succeed())
		runner := idleRunner{}
		prs := func(context.Context, string) (provider.Client, int, error) { return nil, 7, nil }

		host, err := run.NewHost(run.Request{Dir: repo, Remote: runner, PRs: prs, Options: run.Options{Host: lifecycle.HostDashboard}})

		Expect(err).NotTo(HaveOccurred())
		Expect(host.Remote).To(Equal(lifecycle.RemoteRunner(runner)))
		Expect(host.PRs).NotTo(BeNil(), "the host reads linked PRs through the request's client")
		Expect(host.Config.SSH.Cmd).To(BeEmpty(), "the serve host's own configuration doesn't apply")
	})

	It("is a local one otherwise", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		Expect(os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("ssh:\n  cmd: make check\n"), 0o644)).To(Succeed())

		host, err := run.NewHost(run.Request{Dir: GinkgoT().TempDir()})

		Expect(err).NotTo(HaveOccurred())
		Expect(host.Remote).To(BeNil())
		Expect(host.Config.SSH.Cmd).To(Equal("make check"))
	})
})
