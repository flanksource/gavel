package types

import (
	"context"
	"os"
	"path/filepath"
	"runtime/pprof"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture Go profile artifacts", func() {
	It("passes the destination to an exec only with --profile and fails emitted invalid data", func() {
		root := GinkgoT().TempDir()
		fixture := fixtures.FixtureTest{
			Name: "emitter", SourceDir: root,
			ExecFixtureBase: fixtures.ExecFixtureBase{
				Exec: "sh", Args: []string{"-c", "if [ -n \"$PROFILE_OUTPUT\" ]; then printf invalid > \"$PROFILE_OUTPUT\"; fi"},
				GoProfiles: map[string]fixtures.GoProfileOutput{"cpu": {File: "cpu.pprof", Env: "PROFILE_OUTPUT"}},
			},
		}
		without := (&ExecFixture{}).Run(context.Background(), fixture, fixtures.RunOptions{WorkDir: root})
		Expect(without.GoProfiles).To(BeEmpty())
		Expect(without.Status).To(Equal(task.StatusPASS))
		with := (&ExecFixture{}).Run(context.Background(), fixture, fixtures.RunOptions{WorkDir: root, Profile: true})
		Expect(with.GoProfiles).To(HaveLen(1))
		Expect(with.GoProfiles[0].Status).To(Equal("invalid"))
		Expect(with.Error).To(ContainSubstring("go profile cpu"))
	})

	It("captures a Go test CPU profile from a fixture exec", func() {
		root := GinkgoT().TempDir()
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		fixture := fixtures.FixtureTest{
			Name: "Go CPU", SourceDir: filepath.Dir(cwd),
			ExecFixtureBase: fixtures.ExecFixtureBase{
				Exec: "go", Args: []string{"test", "./record", "-run", "TestParseShorthand"},
				GoProfiles: map[string]fixtures.GoProfileOutput{"cpu": {File: "cpu.pprof", Args: []string{"-cpuprofile={{.path}}"}}},
			},
		}
		result := (&ExecFixture{}).Run(context.Background(), fixture, fixtures.RunOptions{WorkDir: root, Profile: true})
		Expect(result.Error).To(BeEmpty())
		Expect(result.GoProfiles).To(HaveLen(1))
		Expect(result.GoProfiles[0].Status).To(Equal("captured"))
		Expect(result.GoProfiles[0].Bytes).To(BeNumerically(">", 0))
	})

	It("keeps an absent declared profile as not emitted", func() {
		root := GinkgoT().TempDir()
		prepared, err := prepareGoProfiles(root, map[string]fixtures.GoProfileOutput{
			"cpu": {File: "cpu.pprof", Args: []string{"-cpuprofile={{.path}}"}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(prepared.Args).To(HaveLen(1))
		Expect(prepared.Args[0]).To(HavePrefix("-cpuprofile="))
		artifacts, err := collectGoProfiles(prepared)
		Expect(err).NotTo(HaveOccurred())
		Expect(artifacts).To(HaveLen(1))
		Expect(artifacts[0].Status).To(Equal("not_emitted"))
	})

	It("accepts emitted pprof data and rejects an emitted invalid file", func() {
		root := GinkgoT().TempDir()
		prepared, err := prepareGoProfiles(root, map[string]fixtures.GoProfileOutput{
			"heap": {File: "heap.pprof", Env: "HEAP_PROFILE"},
		})
		Expect(err).NotTo(HaveOccurred())
		file, err := os.Create(prepared.Env["HEAP_PROFILE"])
		Expect(err).NotTo(HaveOccurred())
		Expect(pprof.Lookup("goroutine").WriteTo(file, 0)).To(Succeed())
		Expect(file.Close()).To(Succeed())
		artifacts, err := collectGoProfiles(prepared)
		Expect(err).NotTo(HaveOccurred())
		Expect(artifacts[0].Status).To(Equal("captured"))
		Expect(artifacts[0].Bytes).To(BeNumerically(">", 0))

		Expect(os.WriteFile(filepath.Join(filepath.Dir(prepared.Env["HEAP_PROFILE"]), "heap.pprof"), []byte("invalid"), 0o600)).To(Succeed())
		artifacts, err = collectGoProfiles(prepared)
		Expect(err).To(HaveOccurred())
		Expect(artifacts[0].Status).To(Equal("invalid"))
	})

	It("accepts an absent profile directory and rejects escaping destinations", func() {
		root := GinkgoT().TempDir()
		prepared, err := prepareGoProfiles(root, map[string]fixtures.GoProfileOutput{
			"samples": {Directory: "samples", Env: "PROFILE_DIRECTORY"},
		})
		Expect(err).NotTo(HaveOccurred())
		artifacts, err := collectGoProfiles(prepared)
		Expect(err).NotTo(HaveOccurred())
		Expect(artifacts).To(Equal([]fixtures.GoProfileArtifact{{Name: "samples", Status: "not_emitted"}}))
		Expect(os.Remove(prepared.Env["PROFILE_DIRECTORY"])).To(Succeed())
		artifacts, err = collectGoProfiles(prepared)
		Expect(err).NotTo(HaveOccurred())
		Expect(artifacts).To(Equal([]fixtures.GoProfileArtifact{{Name: "samples", Status: "not_emitted"}}))
		_, err = prepareGoProfiles(root, map[string]fixtures.GoProfileOutput{
			"cpu": {File: "../outside.pprof", Env: "PROFILE_FILE"},
		})
		Expect(err).To(MatchError(ContainSubstring("must be relative")))
	})

	It("rejects an emitted profile through a symlinked parent", func() {
		root := GinkgoT().TempDir()
		outside := GinkgoT().TempDir()
		prepared, err := prepareGoProfiles(root, map[string]fixtures.GoProfileOutput{
			"cpu": {File: "output/cpu.pprof", Env: "PROFILE_FILE"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Remove(filepath.Join(prepared.Root, "output"))).To(Succeed())
		Expect(os.Symlink(outside, filepath.Join(prepared.Root, "output"))).To(Succeed())
		Expect(os.WriteFile(filepath.Join(outside, "cpu.pprof"), []byte("invalid"), 0o600)).To(Succeed())
		artifacts, err := collectGoProfiles(prepared)
		Expect(err).To(MatchError(ContainSubstring("outside fixture profile directory")))
		Expect(artifacts[0].Status).To(Equal("invalid"))
	})
})
