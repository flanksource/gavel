package snapshots

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/commons/properties"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Local snapshot cache", func() {
	It("compresses legacy results without changing their age or invalidating pointers", func() {
		ctx := context.Background()
		repo := initRepo(GinkgoTB())
		properties.Set(compressionDefaultProperty, "none")
		DeferCleanup(properties.Unset, compressionDefaultProperty)
		snapshot := &testui.Snapshot{Error: strings.Repeat("failed assertion\n", 5_000)}
		path, err := Save(repo, snapshot)
		Expect(err).NotTo(HaveOccurred())
		age := time.Now().Truncate(time.Second).Add(-12 * 24 * time.Hour)
		Expect(os.Chtimes(path, age, age)).To(Succeed())
		before, err := CacheInfo(repo)
		Expect(err).NotTo(HaveOccurred())

		properties.Set(compressionDefaultProperty, "gzip")
		Expect(CompressCache(ctx, repo)).To(Equal(1))
		Expect(path).NotTo(BeAnExistingFile())
		Expect(ResolveLast(repo)).To(Equal(path + ".gz"))
		for _, name := range []string{PointerLast, branchPointerName(repo)} {
			pointer, err := LoadPointer(repo, name)
			Expect(err).NotTo(HaveOccurred())
			Expect(filepath.Join(repo, pointer.Path)).To(Equal(path + ".gz"))
		}
		Expect(LoadRunFile(path)).To(Equal(snapshot))
		info, err := os.Stat(path + ".gz")
		Expect(err).NotTo(HaveOccurred())
		Expect(info.ModTime()).To(BeTemporally("==", age))
		after, err := CacheInfo(repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(after.Bytes).To(BeNumerically("<", before.Bytes))
		Expect(CompressCache(ctx, repo)).To(Equal(0))
	})

	It("combines age and size thresholds and keeps live runs and pointed-to results", func() {
		ctx := context.Background()
		repo := initRepo(GinkgoTB())
		properties.Set(compressionDefaultProperty, "none")
		DeferCleanup(properties.Unset, compressionDefaultProperty)
		now := time.Now().Truncate(time.Second)
		older := func(days int) time.Time { return now.Add(-time.Duration(days) * 24 * time.Hour) }
		large := &testui.Snapshot{Error: strings.Repeat("x", 10_000_000)}
		oldSmall := sampleSnapshot()
		live := &testui.Snapshot{Status: testui.SnapshotStatus{Running: true}}
		paths := make([]string, 0, 5)
		for _, item := range []struct {
			snapshot *testui.Snapshot
			age      int
		}{{oldSmall, 31}, {large, 11}, {oldSmall, 11}, {large, 2}, {live, 31}} {
			path, err := SavePerRun(repo, item.snapshot, older(item.age), "")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chtimes(path, older(item.age), older(item.age))).To(Succeed())
			paths = append(paths, path)
		}
		last, err := Save(repo, oldSmall)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chtimes(last, older(40), older(40))).To(Succeed())
		unrelated := filepath.Join(repo, Dir, "baseline.json")
		Expect(writeJSON(unrelated, oldSmall)).To(Succeed())
		Expect(os.Chtimes(unrelated, older(40), older(40))).To(Succeed())

		rules, err := ParseCacheAges([]string{"30d", "10MB:10d"})
		Expect(err).NotTo(HaveOccurred())
		Expect(CleanupCache(ctx, repo, rules, now)).To(Equal(2))
		for _, path := range paths[:2] {
			Expect(path).NotTo(BeAnExistingFile())
		}
		for _, path := range append(paths[2:], last, unrelated) {
			Expect(path).To(BeAnExistingFile())
		}
		Expect(CleanupCache(ctx, repo, rules, now)).To(Equal(0))
	})
})
