package lifecycle_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Host prompt context preview", func() {
	const prURL = "https://github.com/acme/service/pull/42"
	var host *lifecycle.Host

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		host = newHost(&fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# Plan"}})
	})

	linkedTodo := func() *types.TODO {
		todo := hostTodo()
		todo.PR = &types.PR{Number: 42, URL: prURL}
		return todo
	}

	It("previews the prompt with gavel's PR context hook applied exactly once", func() {
		resolution, err := host.Resolve(context.Background(), linkedTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Prompt).To(ContainSubstring("gavel pr status " + prURL + " --json --logs"))
		Expect(strings.Count(resolution.Prompt, "<!-- captain:context:pr-status")).To(Equal(1))
		Expect(resolution.Prompt).NotTo(ContainSubstring("captain:context:pr-write"), "a plan run does not publish to the PR")
		Expect(resolution.Spec.Prompt.User).To(Equal(resolution.Prompt))
	})

	It("lets a workspace prompt dir disable the shipped hook", func() {
		dir := filepath.Join(host.WorkDir, ".captain", "prompts")
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "pr-status.prompt"), []byte("---\nhook: {on: prompt.context, disabled: true}\n---\n"), 0o600)).To(Succeed())
		resolution, err := host.Resolve(context.Background(), linkedTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Prompt).NotTo(ContainSubstring("captain:context:pr-status"))
		Expect(resolution.Prompt).To(ContainSubstring("## Linked PR"))
	})
})
