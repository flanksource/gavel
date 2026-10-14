package verify

import (
	"github.com/flanksource/gavel/processruntime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Shared runtime layers", func() {
	It("lets repository globals override home type settings before repository type overrides", func() {
		home, err := parseGavelConfig("home/.gavel.yaml", []byte("runtime: {cpu: 300, env: {KEEP: home}}\nlint:\n  linters:\n    golangci-lint:\n      args: [--concurrency=2]\n      runtime: {cpu: 100, env: {RESET: home}}\n"))
		Expect(err).NotTo(HaveOccurred())
		repo, err := parseGavelConfig("repo/.gavel.yaml", []byte("runtime: {cpu: 200, env: {RESET: ''}}\nlint:\n  runtime: {nice: 10}\n  linters:\n    golangci-lint:\n      runtime: {mem: 2GiB}\n"))
		Expect(err).NotTo(HaveOccurred())
		cfg := MergeGavelConfig(home, repo)
		effective, args, err := cfg.RuntimeFor("lint", "golangci-lint", processruntime.Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(*effective.CPU).To(Equal(200.0))
		Expect(*effective.Nice).To(Equal(10))
		Expect(*effective.Mem).To(Equal("2GiB"))
		Expect(effective.Env).To(Equal(map[string]string{"KEEP": "home", "RESET": ""}))
		Expect(args).To(Equal([]string{"--concurrency=2"}))
	})

	DescribeTable("rejects invalid runtime configuration at load", func(document string) {
		_, err := parseGavelConfig(".gavel.yaml", []byte(document))
		Expect(err).To(HaveOccurred())
	},
		Entry("unknown linter", "lint: {linters: {typo: {runtime: {cpu: 100}}}}"),
		Entry("unknown framework", "test: {frameworks: {typo: {runtime: {cpu: 100}}}}"),
		Entry("invalid global duration", "runtime: {timeout: forever}"),
	)
})
