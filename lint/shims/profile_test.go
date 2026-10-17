package shims

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const twoGB = 2_000_000_000

var _ = Describe("Profiles", func() {
	twoCores := Limits{Slots: 2, MemBytes: twoGB}

	DescribeTable("Slots",
		func(cpu float64, want int) { Expect(Slots(cpu)).To(Equal(want)) },
		Entry("disabled", 0.0, 0),
		Entry("half a core rounds up", 50.0, 1),
		Entry("two cores", 200.0, 2),
		Entry("partial core rounds up", 250.0, 3),
	)

	DescribeTable("golangci-lint arguments",
		func(args, want []string) {
			got, err := Profiles["golangci-lint"].Inject(args, twoCores)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(want))
		},
		Entry("run gets fast-only and concurrency",
			[]string{"run", "./..."}, []string{"run", "--fast-only", "--concurrency=2", "./..."}),
		Entry("global flags before run are kept in place",
			[]string{"--color=never", "run"}, []string{"--color=never", "run", "--fast-only", "--concurrency=2"}),
		Entry("explicit --fast-only=false is preserved",
			[]string{"run", "--fast-only=false"}, []string{"run", "--concurrency=2", "--fast-only=false"}),
		Entry("--enable-only suppresses fast-only",
			[]string{"run", "--enable-only=govet"}, []string{"run", "--concurrency=2", "--enable-only=govet"}),
		Entry("explicit concurrency is preserved",
			[]string{"run", "-j", "8"}, []string{"run", "--fast-only", "-j", "8"}),
		Entry("other subcommands are untouched",
			[]string{"version"}, []string{"version"}),
		Entry("no subcommand is untouched",
			[]string{"--help"}, []string{"--help"}),
	)

	It("omits --concurrency when the CPU limit is disabled", func() {
		got, err := Profiles["golangci-lint"].Inject([]string{"run"}, Limits{})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]string{"run", "--fast-only"}))
	})

	It("adds an eslint cache outside the repository", func() {
		cache, err := os.UserCacheDir()
		Expect(err).NotTo(HaveOccurred())
		got, err := Profiles["eslint"].Inject([]string{"src"}, twoCores)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]string{"--cache", "--cache-location", filepath.Join(cache, "gavel", "eslint") + "/", "src"}))
	})

	DescribeTable("arguments left alone",
		func(tool string, args []string) {
			got, err := Profiles[tool].Inject(args, twoCores)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(args))
		},
		Entry("eslint --no-cache", "eslint", []string{"--no-cache", "src"}),
		Entry("eslint --cache", "eslint", []string{"--cache", "src"}),
		Entry("tsc", "tsc", []string{"--noEmit", "-p", "tsconfig.json"}),
	)

	DescribeTable("environment",
		func(tool string, limits Limits, environ, want map[string]string) {
			Expect(Profiles[tool].Env(limits, environ)).To(Equal(want))
		},
		Entry("go caps procs and soft memory at 90%", "golangci-lint", twoCores, map[string]string{},
			map[string]string{"GOMAXPROCS": "2", "GOMEMLIMIT": "1800000000B"}),
		Entry("go keeps a caller GOMEMLIMIT", "golangci-lint", twoCores, map[string]string{"GOMEMLIMIT": "1GiB"},
			map[string]string{"GOMAXPROCS": "2"}),
		Entry("node caps the heap at 75%", "eslint", twoCores, map[string]string{},
			map[string]string{"UV_THREADPOOL_SIZE": "2", "NODE_OPTIONS": "--max-old-space-size=1430"}),
		Entry("node keeps existing NODE_OPTIONS", "tsc", twoCores, map[string]string{"NODE_OPTIONS": "--enable-source-maps"},
			map[string]string{"UV_THREADPOOL_SIZE": "2", "NODE_OPTIONS": "--enable-source-maps --max-old-space-size=1430"}),
		Entry("node leaves an explicit heap size alone", "tsc", twoCores, map[string]string{"NODE_OPTIONS": "--max-old-space-size=4096"},
			map[string]string{"UV_THREADPOOL_SIZE": "2"}),
		Entry("disabled limits set nothing", "golangci-lint", Limits{}, map[string]string{}, map[string]string{}),
	)
})
