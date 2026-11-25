package verify

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/slots"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Slot configuration", func() {
	It("keeps home limits authoritative while layering the project override", func() {
		home, repo := GinkgoT().TempDir(), GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		Expect(os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("slots: {global: 2, projectDefault: '50%'}\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, ".gavel.yaml"), []byte("slots: {global: 99, projectDefault: 99, project: 3}\n"), 0o600)).To(Succeed())
		config, err := LoadGavelConfig(repo)
		Expect(err).NotTo(HaveOccurred())
		for _, category := range []string{"test", "lint"} {
			resolved, _, err := config.RuntimeFor(category, "go", processruntime.Options{})
			Expect(err).NotTo(HaveOccurred())
			Expect(resolved.Slots).To(Equal(slots.Limits{Global: 2, Project: 3}))
		}
		Expect(os.Remove(filepath.Join(repo, ".gavel.yaml"))).To(Succeed())
		config, err = LoadGavelConfig(repo)
		Expect(err).NotTo(HaveOccurred())
		limits, err := config.SlotLimits()
		Expect(err).NotTo(HaveOccurred())
		Expect(limits).To(Equal(slots.Limits{Global: 2, Project: max(1, runtime.NumCPU()/2)}))
	})
	DescribeTable("fails on invalid slot values at the configuration boundary", func(key, raw string) {
		_, err := parseGavelConfig(".gavel.yaml", []byte("slots: {"+key+": "+raw+"}"))
		Expect(err).To(MatchError(ContainSubstring("slots." + key)))
	}, Entry("global", "global", "0"), Entry("default", "projectDefault", "'-1'"), Entry("project", "project", "'150%'"), Entry("object", "project", "{}"))
})
