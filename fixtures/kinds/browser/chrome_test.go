package browser

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("resolveChrome", func() {
	const cft = "chrome-mac-arm64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing"

	It("uses an explicit binary and fails naming it when it is missing", func() {
		dir := GinkgoT().TempDir()
		binary := filepath.Join(dir, "my-chrome")
		writeFile(binary, "#!/bin/sh\n")
		Expect(resolveChrome(chromeSearch{Binary: binary})).To(Equal(binary))

		missing := filepath.Join(dir, "absent-chrome")
		_, err := resolveChrome(chromeSearch{Binary: missing, PlaywrightRoots: []string{dir}})
		Expect(err).To(MatchError(ContainSubstring(missing)))
	})

	It("prefers the newest Playwright chromium revision over system Chrome", func() {
		cache := GinkgoT().TempDir()
		writeFile(filepath.Join(cache, "chromium-1217", cft), "#!/bin/sh\n")
		writeFile(filepath.Join(cache, "chromium-1228", cft), "#!/bin/sh\n")
		writeFile(filepath.Join(cache, "chromium_headless_shell-1300", "chrome-mac-arm64", "chrome-headless-shell"), "#!/bin/sh\n")
		system := filepath.Join(GinkgoT().TempDir(), "Google Chrome")
		writeFile(system, "#!/bin/sh\n")

		Expect(resolveChrome(chromeSearch{PlaywrightRoots: []string{cache}, System: []string{system}})).
			To(Equal(filepath.Join(cache, "chromium-1228", cft)))
	})

	It("falls through to system Chrome and lists every location tried when none exists", func() {
		system := filepath.Join(GinkgoT().TempDir(), "Google Chrome")
		writeFile(system, "#!/bin/sh\n")
		emptyCache := GinkgoT().TempDir()
		Expect(resolveChrome(chromeSearch{PlaywrightRoots: []string{emptyCache}, System: []string{system}})).To(Equal(system))

		absent := filepath.Join(emptyCache, "no-such-chrome")
		_, err := resolveChrome(chromeSearch{PlaywrightRoots: []string{emptyCache}, System: []string{absent}})
		Expect(err).To(MatchError(And(ContainSubstring(emptyCache), ContainSubstring(absent))))
	})
})
