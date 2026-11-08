package services

import (
	"context"
	"github.com/flanksource/gavel/fixtures"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
)

var _ = ginkgo.Describe("connection selection", func() {
	ginkgo.It("returns absence only when the saved record is missing", func() {
		ginkgo.GinkgoT().Setenv("XDG_CONFIG_HOME", ginkgo.GinkgoT().TempDir())
		step := &fixtures.StepContext{WorkDir: ginkgo.GinkgoT().TempDir()}
		entry := Entry{Deps: "mssql", Connection: &ConnectionLookup{Ref: "connection://missing"}}
		svc, outputs, err := existingConnection(context.Background(), step, entry)
		Expect(err).NotTo(HaveOccurred())
		Expect(svc).To(BeNil())
		Expect(outputs).To(BeNil())
		Expect(os.Mkdir(filepath.Join(step.WorkDir, ".connections"), 0700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(step.WorkDir, ".connections", "missing.yaml"), []byte("invalid: ["), 0600)).To(Succeed())
		_, _, err = existingConnection(context.Background(), step, entry)
		Expect(err).To(HaveOccurred())
	})
	ginkgo.It("accepts an optional named connection alongside provisioning options", func() {
		_, err := decodeEntry("database", map[string]any{
			"deps": "mssql", "runtime": "docker",
			"connection": map[string]any{"ref": "connection://lab", "app": "sqlglot-go"},
		})
		Expect(err).NotTo(HaveOccurred())
	})
})
