package verify

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/commons/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGavelConfigDecoding(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "gavel config decoding")
}

var _ = Describe("loading config fields", func() {
	It("warns about unknown fields with their source and retains known settings", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, ".gavel.yaml")
		Expect(os.WriteFile(path, []byte("todos:\n  futureSetting: true\n  driver: cmux\n  timeout: 45m\n"), 0o600)).To(Succeed())
		var warnings bytes.Buffer
		logger.SetOutput(&warnings)
		DeferCleanup(func() { logger.SetOutput(nil) })
		cfg, err := LoadSingleGavelConfig(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Todos.Timeout).To(Equal("45m"))
		Expect(warnings.String()).To(SatisfyAll(ContainSubstring(path), ContainSubstring("todos.futureSetting"), ContainSubstring("todos.driver")))
	})

	It("inspects flat prompt overrides through their custom decoder", func() {
		var warnings []string
		var cfg GavelConfig
		err := DecodeJSON([]byte(`{"todos":{"run":{"model":"cli:sonnet","file":"run.prompt","futureSetting":true}}}`), &cfg,
			DecodeOptions{Source: "project.yaml", Warn: func(message string) { warnings = append(warnings, message) }})
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Todos.Run.Spec.Name).To(Equal("cli:sonnet"))
		Expect(cfg.Todos.Run.File).To(Equal("run.prompt"))
		Expect(warnings).To(Equal([]string{"project.yaml: unknown field todos.run.futureSetting"}))
	})

	DescribeTable("rejects malformed known settings",
		func(document string) {
			path := filepath.Join(GinkgoT().TempDir(), ".gavel.yaml")
			Expect(os.WriteFile(path, []byte(document), 0o600)).To(Succeed())
			_, err := LoadSingleGavelConfig(path)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(path))
		},
		Entry("field type", "todos:\n  checkConcurrency: many\n"),
		Entry("invalid YAML", "todos: [\n"),
		Entry("duplicate key", "todos:\n  timeout: 45m\n  timeout: 30m\n"),
		Entry("trailing document", "todos: {}\n---\ntodos: {}\n"),
		Entry("conflicting lifecycle forms", "todos:\n  lifecycle:\n    file: lifecycle.yaml\n    name: duplicate\n"),
	)
})
