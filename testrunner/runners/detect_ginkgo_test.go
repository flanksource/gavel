package runners

import (
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Detect", func() {
	It("returns nil when no framework files exist under the working directory", func() {
		cwd := GinkgoT().TempDir()
		writeTree(cwd, map[string]string{"package.json": `{"name":"x"}`, "main.go": "package main\n"})
		for _, runner := range []Runner{NewGoTest(cwd), NewGinkgo(cwd), NewVitest(cwd), NewJest(cwd), NewPlaywright(cwd)} {
			Expect(runner.Detect(cwd, nil)).To(BeNil(), string(runner.Name()))
		}
	})

	It("looks only under the starting paths when they are given", func() {
		cwd := GinkgoT().TempDir()
		writeTree(cwd, map[string]string{"root_test.go": "package root\n", "sub/impl.go": "package sub\n"})
		Expect(NewGoTest(cwd).Detect(cwd, []string{"sub"})).To(BeNil())
		Expect(NewGoTest(cwd).Detect(cwd, nil)).NotTo(BeNil())
	})

	Describe("a test file inside a nested node package", func() {
		var cwd string
		argv := []string{"app/src/a.test.ts"}

		BeforeEach(func() {
			cwd = GinkgoT().TempDir()
			writeTree(cwd, map[string]string{
				"app/package.json":  `{"name":"app","devDependencies":{"vitest":"^1.0.0"}}`,
				"app/src/a.test.ts": "",
				"app/src/b.test.ts": "",
			})
		})

		It("is owned by the runner whose package contains it", func() {
			opts, err := NewVitest(cwd).Detect(cwd, argv)
			Expect(err).NotTo(HaveOccurred())
			Expect(opts).NotTo(BeNil())
			Expect(opts.Files).To(Equal(FileSelection{Exact: []string{"app/src/a.test.ts"}}))

			Expect(NewVitest(cwd).DiscoverPackages(*opts)).To(Equal([]string{"./app"}))
			Expect(opts.PackageFiles("./app", true)).To(Equal([]string{"src/a.test.ts"}))
		})

		It("is not detected by a runner that does not own it", func() {
			Expect(NewJest(cwd).Detect(cwd, argv)).To(BeNil())
		})

		DescribeTable("is not claimed by a node runner whose test suffixes it lacks",
			func(runner Runner, file string) {
				writeTree(cwd, map[string]string{
					"app/playwright.config.ts": "",
					"app/" + file:              "",
				})
				Expect(runner.Detect(cwd, []string{"app/" + file})).To(BeNil())
			},
			Entry("a Go spec file under a vitest package", NewVitest(cwd), "store_test.go"),
			Entry("a vitest .test.ts file under a playwright package", NewPlaywright(cwd), "src/c.test.ts"),
		)

		Describe("a starting directory below the package root", func() {
			const feature = "app/src/feature"

			BeforeEach(func() {
				writeTree(cwd, map[string]string{
					feature + "/c.test.ts":      "",
					feature + "/deep/d.test.ts": "",
					"app/src/other/readme.md":   "",
				})
			})

			It("is detected and discovered as the enclosing package", func() {
				opts, err := NewVitest(cwd).Detect(cwd, []string{feature})
				Expect(err).NotTo(HaveOccurred())
				Expect(opts).NotTo(BeNil())
				opts.Recursive = true
				Expect(NewVitest(cwd).DiscoverPackages(*opts)).To(Equal([]string{"./app"}))
			})

			It("narrows the package run to the directory", func() {
				opts, err := NewVitest(cwd).Detect(cwd, []string{feature})
				Expect(err).NotTo(HaveOccurred())
				run, err := NewVitest(cwd).BuildCommand(opts.ForPackages("./app"))
				Expect(err).NotTo(HaveOccurred())
				Expect(run.Process.Args[len(run.Process.Args)-1]).To(Equal("src/feature"))
			})

			It("runs the whole package when another starting path covers it", func() {
				opts, err := NewVitest(cwd).Detect(cwd, []string{feature, "app"})
				Expect(err).NotTo(HaveOccurred())
				run, err := NewVitest(cwd).BuildCommand(opts.ForPackages("./app"))
				Expect(err).NotTo(HaveOccurred())
				Expect(run.Process.Args).NotTo(ContainElement("src/feature"))
			})

			It("is not claimed when it holds none of the runner's test files", func() {
				Expect(NewVitest(cwd).Detect(cwd, []string{"app/src/other"})).To(BeNil())
				Expect(NewPlaywright(cwd).Detect(cwd, []string{feature})).To(BeNil())
			})
		})

		It("runs only the selected file, after the name filter, workers and bail", func() {
			opts, err := NewVitest(cwd).Detect(cwd, append(slices.Clone(argv), "--", "-t", "adds", "--maxWorkers", "2", "--bail", "--silent"))
			Expect(err).NotTo(HaveOccurred())
			run, err := NewVitest(cwd).BuildCommand(opts.ForPackages("./app"))
			Expect(err).NotTo(HaveOccurred())
			args := run.Process.Args
			vitest := slices.Index(args, "vitest")
			Expect(vitest).To(BeNumerically(">=", 0), "args %v", args)
			Expect(args[vitest+1 : vitest+3]).To(Equal([]string{"run", "--reporter=json"}))
			Expect(args[vitest+5:]).To(Equal([]string{"-t", "adds", "--maxWorkers=2", "--bail=1", "src/a.test.ts", "--silent"}))
		})
	})
})
