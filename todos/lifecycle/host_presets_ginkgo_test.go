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

var _ = Describe("Host runtime preset resolution", func() {
	var host *lifecycle.Host

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		host = newHost(&fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# Plan"}})
		dir := GinkgoT().TempDir()
		for name, document := range map[string]string{
			"requested": "name: requested\nscope: context\nspec:\n  budget: {maxTurns: 13}\n",
			"pinned":    "name: pinned\nscope: context\nspec:\n  budget: {maxTurns: 9}\n",
			"default":   "name: default\nscope: context\nspec:\n  budget: {maxTurns: 5}\n",
			"cost":      "name: cost\nscope: context\nspec:\n  budget: {cost: 4}\n",
			"nested":    "name: nested\nscope: context\npresets: [cost]\n",
		} {
			Expect(os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(document), 0o600)).To(Succeed())
		}
		source, err := runtimeprofiles.NewFileSource(runtimeprofiles.FileSourceOptions{Kind: runtimeprofiles.KindPreset, Dir: dir})
		Expect(err).NotTo(HaveOccurred())
		catalog, err := runtimeprofiles.NewCatalog(source)
		Expect(err).NotTo(HaveOccurred())
		host.Catalog = func(context.Context) (*runtimeprofiles.Catalog, error) { return catalog, nil }
		host.Config.Todos.Presets = []string{"default"}
	})

	It("selects ordered request presets and keeps request spec precedence", func() {
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{
			Presets: []string{"requested", "cost"}, PresetsSet: true,
			Request: api.Spec{Budget: api.Budget{MaxTurns: 17}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.RuntimePresets).NotTo(BeNil())
		Expect(resolution.RuntimePresets.Presets).To(HaveLen(2))
		Expect(resolution.RuntimePresets.Presets[0].Name).To(Equal("requested"))
		Expect(resolution.RuntimePresets.Presets[1].Name).To(Equal("cost"))
		Expect(resolution.Spec.Budget).To(Equal(api.Budget{MaxTurns: 17, Cost: 4, Timeout: lifecycle.DefaultTimeout.String()}))
		Expect(resolution.Trace).To(ContainElements(
			HaveField("Name", "requested"),
			HaveField("Name", "cost"),
		))
	})

	It("selects a prompt pin above step and global defaults", func() {
		path := filepath.Join(host.WorkDir, "run.prompt")
		Expect(os.WriteFile(path, []byte("---\npresets: [pinned]\n---\n{{{body}}}\n"), 0o600)).To(Succeed())
		host.Config.Todos.Run = verify.PromptSpec{File: path, Presets: []string{"requested"}, PresetsSet: true}
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.RuntimePresets.Presets[0].Name).To(Equal("pinned"))
	})

	It("allows an explicit empty request to clear configured defaults", func() {
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{Presets: []string{}, PresetsSet: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.RuntimePresets).To(BeNil())
	})

	It("stores but refuses nested preset execution", func() {
		_, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{Presets: []string{"nested"}, PresetsSet: true})
		Expect(err).To(MatchError(ContainSubstring("nested runtime presets are not supported")))
	})

	It("warns and ignores deprecated runtime profiles", func() {
		host.Config.Todos.Presets = nil
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{
			RuntimeProfile: "requested",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.RuntimeProfile).To(BeNil())
		Expect(resolution.RuntimePresets).To(BeNil())
		Expect(resolution.Warnings).To(ContainElement(api.RuntimeProfileDeprecationWarning))
	})
})
