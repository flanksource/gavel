package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/lint/shims"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("lint shim-exec", func() {
	var binDir string

	fake := func(name, body string) {
		Expect(os.WriteFile(filepath.Join(binDir, name), []byte("#!/bin/sh\n"+body+"\n"), 0o755)).To(Succeed())
	}

	run := func(tool string, args ...string) (int, string, string) {
		var stdout, stderr bytes.Buffer
		code := runShimExec(context.Background(), tool, args, nil, &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}

	BeforeEach(func() {
		binDir = GinkgoT().TempDir()
		GinkgoT().Setenv("PATH", binDir+string(os.PathListSeparator)+"/usr/bin"+string(os.PathListSeparator)+"/bin")
		GinkgoT().Setenv("GOMEMLIMIT", "")
		GinkgoT().Setenv("NODE_OPTIONS", "")
		GinkgoT().Setenv(shims.EnvBypass, "")
		GinkgoT().Setenv(shims.EnvTimeout, "30s")
		GinkgoT().Setenv(shims.EnvCPU, "200")
		GinkgoT().Setenv(shims.EnvMem, "2GB")
		GinkgoT().Setenv(shims.EnvNice, "0")
	})

	It("runs golangci-lint fast-only with capped parallelism and memory, and propagates its exit code", func() {
		fake("golangci-lint", `echo "args=$*"; echo "GOMAXPROCS=$GOMAXPROCS GOMEMLIMIT=$GOMEMLIMIT"; exit 3`)

		code, stdout, stderr := run("golangci-lint", "run", "./...")

		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(Equal("args=run --fast-only --concurrency=2 --allow-parallel-runners ./...\nGOMAXPROCS=2 GOMEMLIMIT=1800000000B\n"))
		Expect(code).To(Equal(3))
	})

	It("caps the node heap for tsc", func() {
		fake("tsc", `echo "NODE_OPTIONS=$NODE_OPTIONS UV_THREADPOOL_SIZE=$UV_THREADPOOL_SIZE"`)

		code, stdout, _ := run("tsc", "--noEmit")

		Expect(code).To(Equal(0))
		Expect(stdout).To(Equal("NODE_OPTIONS=--max-old-space-size=1430 UV_THREADPOOL_SIZE=2\n"))
	})

	It("kills a linter that exceeds the timeout with exit code 124", func() {
		GinkgoT().Setenv(shims.EnvTimeout, "500ms")
		fake("eslint", "sleep 10")

		started := time.Now()
		code, _, stderr := run("eslint", ".")

		Expect(code).To(Equal(124))
		Expect(stderr).To(Equal("gavel shim: eslint exceeded the 500ms timeout\n"))
		Expect(time.Since(started)).To(BeNumerically("<", 7*time.Second))
	})

	It("passes arguments through unchanged when GAVEL_SHIMS=off", func() {
		GinkgoT().Setenv(shims.EnvBypass, "off")
		// The bypass must not inject GOMAXPROCS; clear the host's so it cannot leak in.
		GinkgoT().Setenv("GOMAXPROCS", "")
		fake("golangci-lint", `echo "args=$* GOMAXPROCS=$GOMAXPROCS"`)

		code, stdout, _ := run("golangci-lint", "run")

		Expect(code).To(Equal(0))
		Expect(stdout).To(Equal("args=run GOMAXPROCS=\n"))
	})

	It("fails loudly when the shim settings are missing", func() {
		GinkgoT().Setenv(shims.EnvCPU, "")
		fake("tsc", "exit 0")

		code, _, stderr := run("tsc")

		Expect(code).To(Equal(126))
		Expect(stderr).To(ContainSubstring("missing [GAVEL_SHIM_CPU]"))
	})
})
