package outline

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/testrunner/parsers"
)

var _ = Describe("Node framework collectors", func() {
	It("uses native discovery while keeping test bodies unexecuted", func() {
		workDir := GinkgoT().TempDir()
		jestDir := filepath.Join(workDir, "jest-app")
		playwrightDir := filepath.Join(workDir, "browser-app")
		binDir := filepath.Join(workDir, "bin")
		for _, dir := range []string{jestDir, playwrightDir, binDir} {
			Expect(os.MkdirAll(dir, 0o700)).To(Succeed())
		}

		jestFile := filepath.Join(jestDir, "sum.test.ts")
		Expect(os.WriteFile(filepath.Join(jestDir, "package.json"), []byte(`{"devDependencies":{"jest":"1.0.0"}}`), 0o600)).To(Succeed())
		Expect(os.WriteFile(jestFile, []byte(`test("adds", () => { throw new Error("must not run") })`), 0o600)).To(Succeed())

		Expect(os.WriteFile(filepath.Join(playwrightDir, "package.json"), []byte(`{"devDependencies":{"@playwright/test":"1.0.0"}}`), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(playwrightDir, "login.spec.ts"), []byte(`throw new Error("must not run")`), 0o600)).To(Succeed())

		npx := filepath.Join(binDir, "npx")
		script := `#!/bin/sh
if [ "$1" = "jest" ]; then
  printf '%s\n' 'jest warning' >&2
  printf '["%s"]' "$FAKE_JEST_FILE"
  exit 0
fi
if [ "$1" = "playwright" ]; then
  printf '%s\n' 'playwright warning' >&2
  printf '%s' '{"suites":[{"title":"login.spec.ts","file":"login.spec.ts","specs":[{"title":"logs in","file":"login.spec.ts","line":4,"tests":[{"projectName":"chromium","results":[]}]}]}],"errors":[]}'
  exit 0
fi
exit 2
`
		Expect(os.WriteFile(npx, []byte(script), 0o700)).To(Succeed())
		GinkgoT().Setenv("FAKE_JEST_FILE", jestFile)
		GinkgoT().Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

		jestEntries, err := collectJestTests(context.Background(), workDir, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(jestEntries).To(HaveLen(1))
		Expect(jestEntries[0].Framework).To(Equal(parsers.Jest))
		Expect(jestEntries[0].File).To(Equal("jest-app/sum.test.ts"))

		playwrightEntries, err := collectPlaywrightTests(context.Background(), workDir, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(playwrightEntries).To(HaveLen(1))
		Expect(playwrightEntries[0].Framework).To(Equal(parsers.Playwright))
		Expect(playwrightEntries[0].File).To(Equal("browser-app/login.spec.ts"))
	})
})
