package lint

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/tsc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TypeScript compiler selection", func() {
	It("runs the Node wrapper when the config selects TypeScript and tsrs is installed", func() {
		root, bin := GinkgoT().TempDir(), GinkgoT().TempDir()
		git := exec.Command("git", "init", "--quiet")
		git.Dir = root
		Expect(git.Run()).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, ".gavel.yaml"), []byte("lint:\n  linters:\n    tsc:\n      engine: typescript\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "tsconfig.json"), []byte(`{"files":["app.ts"]}`), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "app.ts"), []byte("export const port: number = 8080\n"), 0o644)).To(Succeed())
		node := filepath.Join(bin, executableFileName("node"))
		Expect(os.WriteFile(node, []byte("#!/bin/sh\nprintf '[]\\n'\n"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(bin, executableFileName("tsrs")), []byte("#!/bin/sh\nexit 37\n"), 0o755)).To(Succeed())
		GinkgoT().Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

		results, err := Execute(Options{WorkDir: root, Linters: []string{"tsc"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(ConsistOf(And(
			HaveField("Command", node),
			HaveField("Args", ConsistOf(HaveSuffix(".cjs"))),
			HaveField("Success", true),
		)))
	})

	It("prefers the workspace tsrs, then the PATH tsrs, then Node", func() {
		root, bin := GinkgoT().TempDir(), GinkgoT().TempDir()
		project := filepath.Join(root, "apps", "web")
		localBin := filepath.Join(root, "node_modules", ".bin")
		Expect(os.MkdirAll(filepath.Join(root, ".git"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(project, 0o755)).To(Succeed())
		Expect(os.MkdirAll(localBin, 0o755)).To(Succeed())
		GinkgoT().Setenv("PATH", bin)

		localTSRS := filepath.Join(localBin, "tsrs")
		if runtime.GOOS == "windows" {
			localTSRS += ".cmd"
		}
		adapter := tsc.NewTSC(project)
		for _, executable := range []string{
			filepath.Join(bin, executableFileName("node")),
			filepath.Join(bin, executableFileName("tsrs")),
			localTSRS,
		} {
			Expect(os.WriteFile(executable, []byte("#!/bin/sh\nexit 0\n"), 0o755)).To(Succeed())
			binary, reason, err := resolveLinterExecutable(context.Background(), adapter, root, project, true, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(reason).To(BeEmpty())
			Expect(binary).To(Equal(executable))
			command, args, err := linters.PrepareCommand(adapter, linters.RunOptions{WorkDir: project, Executable: binary})
			Expect(err).NotTo(HaveOccurred())
			Expect(command).To(Equal(binary))
			if filepath.Base(binary) == executableFileName("node") {
				Expect(args).To(ConsistOf(HaveSuffix(".cjs")))
			} else {
				Expect(args).To(ContainElement("--api"))
			}
		}
	})
})
