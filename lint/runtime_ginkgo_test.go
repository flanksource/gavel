package lint

import (
	"bytes"
	"context"
	"encoding/json"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLint(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lint")
}

func TestMain(m *testing.M) {
	if mode := os.Getenv("GAVEL_RUNTIME_LINTER_CHILD"); mode != "" {
		if mode == "env" {
			if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"Issues": []any{}, "runtimeEnv": os.Getenv("GAVEL_INVOCATION_TEST")}); err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		}
		if mode == "output" {
			_, _ = os.Stdout.WriteString(`{"diagnostics":[]}`)
		}
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

var _ = Describe("Shared linter runtime", func() {
	It("retains invocation configuration while executing at the project root", func() {
		root := GinkgoT().TempDir()
		pkg := filepath.Join(root, "pkg")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		Expect(os.MkdirAll(filepath.Join(root, ".git"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(pkg, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(pkg, ".gavel.yaml"), []byte("runtime: {env: {GAVEL_INVOCATION_TEST: child}}\n"), 0o600)).To(Succeed())
		executable, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		linter, exists := buildLinterRegistry(root).Get("golangci-lint")
		Expect(exists).To(BeTrue())
		var output bytes.Buffer
		result := linters.RunLinterWithTask(context.Background(), nil, linter, linters.RunOptions{WorkDir: root, ConfigDir: pkg, Executable: executable, Runtime: processruntime.Options{Env: map[string]string{"GAVEL_RUNTIME_LINTER_CHILD": "env"}}, OutputTee: &output})
		Expect(result.Success).To(BeTrue(), result.Error)
		Expect(output.String()).To(ContainSubstring(`"runtimeEnv":"child"`))
	})
	It("fails react-doctor after timeout even when it emitted valid JSON", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("runtime: {timeout: 300ms, env: {GAVEL_RUNTIME_LINTER_CHILD: output}}\n"), 0o600)).To(Succeed())
		executable, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		linter, exists := buildLinterRegistry(dir).Get("react-doctor")
		Expect(exists).To(BeTrue())
		linter.(linters.OptionsMixin).SetOptions(linters.RunOptions{WorkDir: dir, Executable: executable})
		_, err = linter.Run(commonsContext.NewContext(context.Background()), nil)
		Expect(err).To(MatchError(ContainSubstring("timed out")))
	})
	for _, name := range verify.BuiltinLinterNames {
		It("enforces memory limits for "+name, func() {
			dir := GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("runtime:\n  mem: 1B\n  env: {GAVEL_RUNTIME_LINTER_CHILD: '1'}\n"), 0o600)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(dir, "node_modules", "typescript"), 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dir, "node_modules", "typescript", "index.js"), []byte("Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 30000)\n"), 0o600)).To(Succeed())
			executable, err := os.Executable()
			Expect(err).NotTo(HaveOccurred())
			linter, exists := buildLinterRegistry(dir).Get(name)
			Expect(exists).To(BeTrue())
			result := linters.RunLinterWithTask(context.Background(), nil, linter, linters.RunOptions{WorkDir: dir, Executable: executable})
			Expect(result.Success).To(BeFalse())
			Expect(result.Error).To(ContainSubstring("resource limit exceeded"))
		})
	}
})
