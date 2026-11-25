package uir

import (
	"context"
	"os"
	"path/filepath"
	"time"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Gavel-owned UIR fragments", func() {
	It("passes native config through with a deadline and removes only its fragment after execution", func() {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		Expect(os.WriteFile(filepath.Join(root, ".gavel.yaml"), []byte("lint:\n  linters:\n    uir:\n      config:\n        version: 1\n        lint:\n          rules:\n            - {id: sample, query: 'SELECT path FROM files()', message: '{{.path}}', severity: warning}\n"), 0o600)).To(Succeed())
		Expect(os.Mkdir(filepath.Join(root, ".uir.d"), 0o700)).To(Succeed())
		authored := filepath.Join(root, ".uir.d", "authored.yaml")
		Expect(os.WriteFile(authored, []byte("authored"), 0o600)).To(Succeed())
		content, err := os.ReadFile("testdata/fragment.sh")
		Expect(err).NotTo(HaveOccurred())
		binary := filepath.Join(root, "uir")
		Expect(os.WriteFile(binary, content, 0o700)).To(Succeed())
		capture := filepath.Join(root, "captured.yaml")
		GinkgoT().Setenv("UIR_TEST_CAPTURE", capture)
		adapter := New(root)
		adapter.SetOptions(linters.RunOptions{WorkDir: root, Executable: binary})
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		_, err = adapter.Run(commonsContext.NewContext(ctx), nil)
		Expect(err).NotTo(HaveOccurred())
		data, err := os.ReadFile(capture)
		Expect(err).NotTo(HaveOccurred())
		var document struct {
			Expires time.Time `yaml:"_expires"`
			Lint    struct {
				Rules []map[string]any `yaml:"rules"`
			} `yaml:"lint"`
		}
		Expect(yaml.Unmarshal(data, &document)).To(Succeed())
		deadline, _ := ctx.Deadline()
		Expect(document.Expires).To(BeTemporally("~", deadline, time.Second))
		Expect(document.Lint.Rules).To(ConsistOf(map[string]any{"id": "sample", "query": "SELECT path FROM files()", "message": "{{.path}}", "severity": "warning"}))
		entries, err := os.ReadDir(filepath.Join(root, ".uir.d"))
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))
		Expect(authored).To(BeAnExistingFile())
	})

	DescribeTable("removes generated fragments when execution cannot finish", func(cancelled bool) {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		Expect(os.WriteFile(filepath.Join(root, ".gavel.yaml"), []byte("lint:\n  linters:\n    uir:\n      config:\n        version: 1\n        lint: {}\n"), 0o600)).To(Succeed())
		adapter := New(root)
		adapter.SetOptions(linters.RunOptions{WorkDir: root, Executable: filepath.Join(root, "missing-uir")})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if cancelled {
			cancel()
		}
		_, err := adapter.Run(commonsContext.NewContext(ctx), nil)
		Expect(err).To(HaveOccurred())
		entries, err := os.ReadDir(filepath.Join(root, ".uir.d"))
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(BeEmpty())
	}, Entry("startup failure", false), Entry("cancelled context", true))

	It("does not create a fragment while rendering a dry-run command", func() {
		root := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(root, ".gavel.yaml"), []byte("lint:\n  linters:\n    uir:\n      config:\n        version: 1\n        lint: {}\n"), 0o600)).To(Succeed())
		binary, args := New(root).DryRunCommand()
		Expect(binary).To(Equal("uir"))
		Expect(args).To(Equal([]string{"lint", "--format", "json", "--no-progress"}))
		Expect(filepath.Join(root, ".uir.d")).NotTo(BeADirectory())
	})
})
