package utils

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResolveWithin", func() {
	const (
		nested       = "prompts/grouping.prompt"
		escapingName = "../../etc/passwd"
	)
	var base string

	BeforeEach(func() {
		base = GinkgoT().TempDir()
	})

	It("resolves a relative name to an absolute path under the base", func() {
		Expect(ResolveWithin(base, nested)).To(Equal(filepath.Join(base, nested)))
	})

	It("accepts an absolute name that already lives under the base", func() {
		Expect(ResolveWithin(base, filepath.Join(base, nested))).To(Equal(filepath.Join(base, nested)))
	})

	// A workspace naming its own root — the dashboard's commonest dir param —
	// resolves to the root, not to an escape.
	It("accepts the base itself", func() {
		Expect(ResolveWithin(base, base)).To(Equal(filepath.Clean(base)))
		Expect(ResolveWithin(base, ".")).To(Equal(filepath.Clean(base)))
	})

	It("rejects a relative name that walks out of the base", func() {
		_, err := ResolveWithin(base, escapingName)

		Expect(err).To(MatchError(ErrPathEscapesBase))
		Expect(err).To(MatchError(ContainSubstring(escapingName)))
		Expect(err).To(MatchError(ContainSubstring(base)))
	})

	It("rejects an absolute name that points outside the base", func() {
		outside := filepath.Join(GinkgoT().TempDir(), nested)

		_, err := ResolveWithin(base, outside)

		Expect(err).To(MatchError(ErrPathEscapesBase))
		Expect(err).To(MatchError(ContainSubstring(outside)))
	})

	It("rejects a name that climbs above the base and back down into a sibling", func() {
		sibling := filepath.Join("..", filepath.Base(base)+"-other", nested)

		_, err := ResolveWithin(base, sibling)

		Expect(err).To(MatchError(ErrPathEscapesBase))
	})

	It("requires both a base and a name", func() {
		_, err := ResolveWithin(base, "")
		Expect(err).To(MatchError(ContainSubstring("path is required")))

		_, err = ResolveWithin("", nested)
		Expect(err).To(MatchError(ContainSubstring("base directory is required")))
	})
})

var _ = Describe("ReadFileWithin", func() {
	const nested = "prompts/grouping.prompt"
	var base, outside string

	BeforeEach(func() {
		base = GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(base, filepath.Dir(nested)), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(base, nested), []byte("inside"), 0o600)).To(Succeed())
		outside = filepath.Join(GinkgoT().TempDir(), "secret.txt")
		Expect(os.WriteFile(outside, []byte("secret"), 0o600)).To(Succeed())
	})

	It("reads a file under the base, named relative or absolute", func() {
		Expect(ReadFileWithin(base, nested)).To(Equal([]byte("inside")))
		Expect(ReadFileWithin(base, filepath.Join(base, nested))).To(Equal([]byte("inside")))
	})

	It("refuses a name that resolves outside the base", func() {
		for _, name := range []string{"../" + filepath.Base(outside), outside} {
			_, err := ReadFileWithin(base, name)
			Expect(err).To(MatchError(ErrPathEscapesBase), name)
		}
	})

	// A checked-in symlink passes the lexical check, so the read itself must refuse it.
	It("refuses a symlink under the base that leads outside it", func() {
		Expect(os.Symlink(outside, filepath.Join(base, "link.txt"))).To(Succeed())

		data, err := ReadFileWithin(base, "link.txt")

		Expect(err).To(MatchError(ContainSubstring("escapes")))
		Expect(err).To(MatchError(ContainSubstring("link.txt")))
		Expect(data).To(BeEmpty())
	})

	It("reports a missing file as not existing", func() {
		_, err := ReadFileWithin(base, "missing.prompt")
		Expect(errors.Is(err, fs.ErrNotExist)).To(BeTrue(), "%v", err)
	})
})
