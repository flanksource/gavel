package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/workspace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("gavel fixtures --from-step", func() {
	const fromStep = "apply"
	var workDir, file string
	var dirs workspace.Dirs

	// runAndSave runs the fixture as a terminal run started at started and
	// saves it, resumed when resume is set. It returns the saved run's id.
	runAndSave := func(started time.Time, resume *fixtures.Resume) string {
		GinkgoHelper()
		runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{file}, WorkDir: workDir, MaxWorkers: 1, Resume: resume})
		Expect(err).NotTo(HaveOccurred())
		tree, runErr := runner.Run()
		Expect(tree).NotTo(BeNil())
		opts, _ := fixtureUIRunOptions(fixtureUIRunRequest{Runner: &fixtures.RunnerOptions{Paths: []string{file}, WorkDir: workDir}})
		opts.Dirs = dirs
		Expect(saveFixtureRun(finishedRun{Opts: opts, Started: started, Err: runErr}, runner, tree)).To(Succeed())
		runs, err := snapshots.ListRuns(workDir, snapshots.ListRunsOptions{Limit: 1})
		Expect(err).NotTo(HaveOccurred())
		return runs[0].RunID
	}
	flags := func(fromRun string) fixtureResumeFlags {
		return fixtureResumeFlags{FromStep: fromStep, FromRun: fromRun, Files: []string{file}, Dirs: dirs}
	}

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		dirs = workspace.Dirs{Invocation: workDir, StateRoot: workDir}
		file = filepath.Join(workDir, "purge.fixture.yaml")
		Expect(os.WriteFile(file, []byte(`kind: Fixture
steps:
  - {id: lookup, kind: exec, script: "echo run >> lookup.log; echo east"}
  - {id: broken, kind: exec, script: "echo run >> broken.log; exit 3"}
  - {id: apply, kind: exec, script: "printf %s '{{.steps.lookup.stdout}}' >> apply.log"}
`), 0o600)).To(Succeed())
	})

	It("runs only the step and those after it, over what a failed terminal run left", func() {
		failed := runAndSave(time.Now().Add(-time.Minute), nil)

		rerun, err := flags(failed).resume()
		Expect(err).NotTo(HaveOccurred())
		runAndSave(time.Now(), rerun.Resume)

		read := func(name string) string {
			data, err := os.ReadFile(filepath.Join(workDir, name))
			Expect(err).NotTo(HaveOccurred())
			return string(data)
		}
		Expect(read("lookup.log")).To(Equal("run\n"))
		Expect(read("broken.log")).To(Equal("run\n"))
		Expect(read("apply.log")).To(Equal("east\n"))
		Expect(rerun.Resume.Prior["broken"].Failed).To(BeTrue())
	})

	DescribeTable("finds the run it resumes",
		func(fromRun func(older string) string, stdout string) {
			older := runAndSave(time.Now().Add(-time.Hour), nil)
			document, err := os.ReadFile(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(file, []byte(strings.Replace(string(document), "echo east", "echo west", 1)), 0o600)).To(Succeed())
			runAndSave(time.Now().Add(-time.Minute), nil)

			rerun, err := flags(fromRun(older)).resume()

			Expect(err).NotTo(HaveOccurred())
			Expect(rerun.Resume.Prior["lookup"].Outputs["stdout"]).To(ContainSubstring(stdout))
		},
		Entry("by its id", func(older string) string { return older }, "east"),
		Entry("by its file", func(older string) string { return filepath.Join(workDir, snapshots.Dir, older+".json") }, "east"),
		Entry("as the newest by default", func(string) string { return "" }, "west"),
	)

	DescribeTable("is refused", func(edit func(*fixtureResumeFlags), message string) {
		given := flags("")
		edit(&given)

		_, err := given.resume()

		Expect(err).To(MatchError(ContainSubstring(message)))
	},
		Entry("without a saved run", func(*fixtureResumeFlags) {}, "no saved run of"),
		Entry("with --from-run alone", func(f *fixtureResumeFlags) { f.FromStep, f.FromRun = "", "run-x" }, "--from-run needs --from-step"),
		Entry("for several files", func(f *fixtureResumeFlags) { f.Files = append(f.Files, "other.fixture.yaml") }, "one fixture file, got 2"),
		Entry("for a glob", func(f *fixtureResumeFlags) { f.Files = []string{"*.fixture.yaml"} }, "is not a file"),
		Entry("with an unknown run", func(f *fixtureResumeFlags) { f.FromRun = "run-2026-01-01T00-00-00Z" }, "--from-run"),
	)

	It("lays supplied params over the resumed run's answers", func() {
		rerun := fixtureRerun{Params: map[string]any{"db": "lab", "scale": true}}

		Expect(rerun.withParams(map[string]any{"db": "qa"})).To(Equal(map[string]any{"db": "qa", "scale": true}))
		Expect(rerun.Params).To(HaveKeyWithValue("db", "lab"))
		Expect(fixtureRerun{}.withParams(nil)).To(BeNil())
	})

	It("builds a rerun command with the run's flags, without its params, step or run", func() {
		root, cmd := &cobra.Command{Use: "gavel"}, &cobra.Command{Use: "fixtures"}
		root.AddCommand(cmd)
		for _, name := range []string{"params", "from-step", "from-run", "traces", "unset"} {
			cmd.Flags().String(name, "", "")
		}
		cmd.Flags().StringArray("param", nil, "")
		cmd.Flags().StringArray("tag", nil, "")
		Expect(cmd.ParseFlags([]string{"--param", "password=secret", "--params", "p.yaml", "--from-step", "a", "--from-run", "run-x",
			"--traces", "http_proxy,terminal", "--tag", "a b", "--tag", "c"})).To(Succeed())

		command, err := fixtureCommand(cmd)

		Expect(err).NotTo(HaveOccurred())
		Expect(command.Argv).To(Equal([]string{os.Args[0], "fixtures", "--tag=a b", "--tag=c", "--traces=http_proxy,terminal"}))
	})

	It("keeps the resume off the fixture runner, so a whole UI rerun does not resume", func() {
		resume := &fixtures.Resume{FromStep: fromStep}
		runner := &fixtures.RunnerOptions{Paths: []string{file}, WorkDir: workDir}

		opts, _ := fixtureUIRunOptions(fixtureUIRunRequest{Runner: runner, Resume: resume})

		Expect(opts.FixtureResume).To(BeIdenticalTo(resume))
		Expect(opts.FixtureRunner.Resume).To(BeNil())
	})
})
