package testrunner

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nested test configuration", func() {
	It("separates root-discovered packages with different environments", func() {
		repo := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		Expect(os.Mkdir(filepath.Join(repo, ".git"), 0o755)).To(Succeed())
		for _, name := range []string{"a", "b"} {
			dir := filepath.Join(repo, name)
			Expect(os.Mkdir(dir, 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("runtime: {env: {PACKAGE_ENV: "+name+"}}\n"), 0o600)).To(Succeed())
		}
		o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: repo}, registry: DefaultRegistry(repo)}
		plans, err := o.planGoBatches(map[Framework][]string{parsers.GoTest: {"./a", "./b"}}, nil, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(plans).To(HaveLen(2))
		for _, plan := range plans {
			command, err := o.buildGoBatchCommand(plan)
			Expect(err).NotTo(HaveOccurred())
			Expect(command.Env["PACKAGE_ENV"]).To(Equal(filepath.Base(plan.packages[0])))
		}
	})
})
