package lifecycle_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("deprecated runtime profile selection", func() {
	var host *lifecycle.Host

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		host = newHost(&fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# Plan"}})
		host.Catalog = func(context.Context) (*runtimeprofiles.Catalog, error) {
			Fail("deprecated runtime profiles must not open the catalog")
			return nil, nil
		}
	})

	It("warns and ignores a requested profile", func() {
		withoutProfile, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		withProfile, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{RuntimeProfile: "implementer"})
		Expect(err).NotTo(HaveOccurred())
		Expect(withProfile.RuntimeProfile).To(BeNil())
		Expect(withProfile.Spec).To(Equal(withoutProfile.Spec))
		Expect(withProfile.Warnings).To(ContainElement(api.RuntimeProfileDeprecationWarning))
	})

	It("warns and ignores project and step profile defaults", func() {
		host.Config.Todos.RuntimeProfile = "project-default"
		host.Config.Todos.Run.RuntimeProfile = "step-default"
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.RuntimeProfile).To(BeNil())
		Expect(resolution.Warnings).To(ContainElement(api.RuntimeProfileDeprecationWarning))
	})

	It("warns and ignores a prompt frontmatter profile", func() {
		path := filepath.Join(host.WorkDir, "run.prompt")
		Expect(os.WriteFile(path, []byte("---\nruntimeProfile: reviewer\n---\n{{{body}}}\n"), 0o600)).To(Succeed())
		host.Config.Todos.Run = verify.PromptSpec{File: path}
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.RuntimeProfile).To(BeNil())
		Expect(resolution.Warnings).To(ContainElement(api.RuntimeProfileDeprecationWarning))
	})
})
