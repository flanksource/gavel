package runcache_test

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/internal/runcache"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {
	var store *runcache.Store

	BeforeEach(func() {
		dir := GinkgoT().TempDir()
		var err error
		store, err = runcache.Open(dir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("records a successful entry and retrieves it by fingerprint", func() {
		fp := "abcdef0123456789"
		entry := runcache.Entry{
			ImportPath:    "example.com/fix/a",
			Framework:     "go test",
			ExitCode:      0,
			PassCount:     3,
			DurationNanos: 1_000_000,
		}
		Expect(store.Record(fp, false, entry)).To(Succeed())

		got, ok := store.Lookup(fp)
		Expect(ok).To(BeTrue())
		Expect(got.ImportPath).To(Equal("example.com/fix/a"))
		Expect(got.PassCount).To(Equal(3))
		Expect(got.RecordedAt).NotTo(BeZero())
	})

	It("returns a miss for an unknown fingerprint", func() {
		_, ok := store.Lookup("does-not-exist")
		Expect(ok).To(BeFalse())
	})

	It("refuses to record a failing entry", func() {
		fp := "fail"
		entry := runcache.Entry{ExitCode: 1, FailCount: 2}
		Expect(store.Record(fp, false, entry)).To(Succeed())

		_, ok := store.Lookup(fp)
		Expect(ok).To(BeFalse())
	})

	It("refuses to record when tooRecent is true", func() {
		fp := "tooRecent"
		entry := runcache.Entry{ExitCode: 0, PassCount: 1}
		Expect(store.Record(fp, true, entry)).To(Succeed())

		_, ok := store.Lookup(fp)
		Expect(ok).To(BeFalse())
	})

	It("treats a corrupted entry as a miss", func() {
		fp := "corrupt"
		Expect(store.Record(fp, false, runcache.Entry{ExitCode: 0, PassCount: 1})).To(Succeed())

		// Overwrite with garbage.
		shard := fp[:2]
		path := filepath.Join(store.Dir, shard, fp+".json")
		Expect(os.WriteFile(path, []byte("{not json"), 0o644)).To(Succeed())

		_, ok := store.Lookup(fp)
		Expect(ok).To(BeFalse())
	})

	It("shards entries by the first two fingerprint characters", func() {
		fp := "ab123456"
		Expect(store.Record(fp, false, runcache.Entry{ExitCode: 0, PassCount: 1})).To(Succeed())

		_, err := os.Stat(filepath.Join(store.Dir, "ab", fp+".json"))
		Expect(err).NotTo(HaveOccurred())
	})
})
