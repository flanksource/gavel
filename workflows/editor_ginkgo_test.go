package workflows_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/gavel/workflows"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEditor(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Lifecycle editor") }

var _ = Describe("scoped lifecycle editing", func() {
	It("previews and saves a confined file without changing its reference", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, ".gavel.yaml")
		file := filepath.Join(dir, "workflow.yaml")
		Expect(os.WriteFile(path, []byte("todos:\n  lifecycle:\n    file: workflow.yaml\n"), 0600)).To(Succeed())
		original := []byte("order: [run, plan]\n")
		Expect(os.WriteFile(file, original, 0600)).To(Succeed())
		view, err := workflows.Inspect("todos", dir)
		Expect(err).NotTo(HaveOccurred())
		disabled := []string{"triage"}
		edit := workflows.Edit{Revision: view.Revision, Source: view.Source, FileSource: &verify.LifecycleConfig{Disabled: &disabled}}
		_, err = workflows.Preview("todos", dir, &edit)
		Expect(err).NotTo(HaveOccurred())
		data, err := os.ReadFile(file)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(original))
		saved, err := workflows.Save("todos", dir, edit)
		Expect(err).NotTo(HaveOccurred())
		Expect(saved.Source.File).To(Equal("workflow.yaml"))
		Expect(saved.Effective.StepNames()).NotTo(ContainElement("triage"))
	})
	It("preserves unrelated configuration and rejects stale saves", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, ".gavel.yaml")
		Expect(os.WriteFile(path, []byte("# keep this comment\npr:\n  draft: true\n"), 0600)).To(Succeed())
		view, err := workflows.Inspect("todos", dir)
		Expect(err).NotTo(HaveOccurred())
		order := []string{"run", "plan"}
		request := workflows.Edit{Revision: view.Revision, Source: verify.LifecycleConfig{Order: &order}}
		_, err = workflows.Save("todos", dir, request)
		Expect(err).NotTo(HaveOccurred())
		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("# keep this comment"))
		Expect(string(data)).To(ContainSubstring("draft: true"))
		_, err = workflows.Save("todos", dir, request)
		Expect(err).To(MatchError(ContainSubstring("changed")))
	})
	It("rejects disabling mandatory verification without changing the source", func() {
		dir := GinkgoT().TempDir()
		view, err := workflows.Inspect("todos", dir)
		Expect(err).NotTo(HaveOccurred())
		disabled := []string{"verify"}
		_, err = workflows.Save("todos", dir, workflows.Edit{Revision: view.Revision, Source: verify.LifecycleConfig{Disabled: &disabled}})
		Expect(err).To(HaveOccurred())
		_, err = os.Stat(filepath.Join(dir, ".gavel.yaml"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})
})
