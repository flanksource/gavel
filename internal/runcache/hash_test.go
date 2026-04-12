package runcache_test

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/internal/changegraph"
	"github.com/flanksource/gavel/internal/runcache"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// buildFixtureModule mirrors the golist_test fixture: a → b → c, with a
// testdata file under c/.
func buildHashFixture() string {
	root := GinkgoT().TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, rel)
		Expect(os.MkdirAll(filepath.Dir(p), 0o755)).To(Succeed())
		Expect(os.WriteFile(p, []byte(content), 0o644)).To(Succeed())
	}
	write("go.mod", "module example.com/fix\n\ngo 1.21\n")
	write("a/a.go", `package a

import "example.com/fix/b"

func A() string { return b.B() }
`)
	write("b/b.go", `package b

import "example.com/fix/c"

func B() string { return c.C() }
`)
	write("c/c.go", `package c

func C() string { return "c" }
`)
	write("c/testdata/fixture.yaml", "value: original\n")

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = root
	out, err := tidy.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "go mod tidy: %s", out)

	return root
}

var _ = Describe("Hasher", func() {
	var (
		root  string
		graph *changegraph.Graph
	)

	BeforeEach(func() {
		root = buildHashFixture()
		var err error
		graph, err = changegraph.Load(root)
		Expect(err).NotTo(HaveOccurred())
	})

	fingerprintsOf := func() map[string]string {
		h := runcache.NewHasher(graph, nil)
		out := map[string]string{}
		for _, ip := range []string{"example.com/fix/a", "example.com/fix/b", "example.com/fix/c"} {
			fp, err := h.Effective(ip)
			Expect(err).NotTo(HaveOccurred())
			out[ip] = fp.Hex
		}
		return out
	}

	It("produces stable fingerprints across runs with no changes", func() {
		before := fingerprintsOf()

		// Reload the graph to simulate a fresh invocation, then recompute.
		var err error
		graph, err = changegraph.Load(root)
		Expect(err).NotTo(HaveOccurred())

		after := fingerprintsOf()
		Expect(after).To(Equal(before))
	})

	It("propagates changes to a leaf package up its reverse-dep chain", func() {
		before := fingerprintsOf()

		// Edit c/c.go — should change c, b, and a.
		Expect(os.WriteFile(filepath.Join(root, "c", "c.go"), []byte(`package c

func C() string { return "updated" }
`), 0o644)).To(Succeed())

		// Reload graph after edit.
		var err error
		graph, err = changegraph.Load(root)
		Expect(err).NotTo(HaveOccurred())

		after := fingerprintsOf()
		Expect(after["example.com/fix/c"]).NotTo(Equal(before["example.com/fix/c"]))
		Expect(after["example.com/fix/b"]).NotTo(Equal(before["example.com/fix/b"]))
		Expect(after["example.com/fix/a"]).NotTo(Equal(before["example.com/fix/a"]))
	})

	It("only changes the leaf fingerprint when an unrelated leaf is edited", func() {
		before := fingerprintsOf()

		// Edit a/a.go — a has no reverse deps, so only a should change.
		Expect(os.WriteFile(filepath.Join(root, "a", "a.go"), []byte(`package a

import "example.com/fix/b"

func A() string { return b.B() + "!" }
`), 0o644)).To(Succeed())

		var err error
		graph, err = changegraph.Load(root)
		Expect(err).NotTo(HaveOccurred())

		after := fingerprintsOf()
		Expect(after["example.com/fix/a"]).NotTo(Equal(before["example.com/fix/a"]))
		Expect(after["example.com/fix/b"]).To(Equal(before["example.com/fix/b"]))
		Expect(after["example.com/fix/c"]).To(Equal(before["example.com/fix/c"]))
	})

	It("changes the fingerprint when a testdata file is edited", func() {
		before := fingerprintsOf()

		Expect(os.WriteFile(
			filepath.Join(root, "c", "testdata", "fixture.yaml"),
			[]byte("value: changed\n"),
			0o644,
		)).To(Succeed())

		var err error
		graph, err = changegraph.Load(root)
		Expect(err).NotTo(HaveOccurred())

		after := fingerprintsOf()
		Expect(after["example.com/fix/c"]).NotTo(Equal(before["example.com/fix/c"]))
		Expect(after["example.com/fix/b"]).NotTo(Equal(before["example.com/fix/b"]))
		Expect(after["example.com/fix/a"]).NotTo(Equal(before["example.com/fix/a"]))
	})

	It("marks freshly-edited packages as TooRecent", func() {
		// Edit right before computing: the mtime is within ModTimeCutoff.
		Expect(os.WriteFile(filepath.Join(root, "c", "c.go"), []byte(`package c

func C() string { return "just edited" }
`), 0o644)).To(Succeed())

		var err error
		graph, err = changegraph.Load(root)
		Expect(err).NotTo(HaveOccurred())

		h := runcache.NewHasher(graph, nil)
		fp, err := h.Effective("example.com/fix/c")
		Expect(err).NotTo(HaveOccurred())
		Expect(fp.TooRecent).To(BeTrue())

		// Reverse deps should inherit the TooRecent flag via dep propagation.
		a, err := h.Effective("example.com/fix/a")
		Expect(err).NotTo(HaveOccurred())
		Expect(a.TooRecent).To(BeTrue())
	})
})
