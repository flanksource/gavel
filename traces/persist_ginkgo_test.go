// Tests that a closed run's trace store is copied, file for file, to the configured blob store
// under traces/<runID>/, and that without one it stays only in .gavel/traces.
package traces_test

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/traces"
)

var _ = Describe("persisting a run's trace store", func() {
	It("copies every file of the closed store to traces/<runID>/", func(ctx SpecContext) {
		host, err := traces.Open(GinkgoT().TempDir(), traces.RunID(time.Now()))
		Expect(err).ToNot(HaveOccurred())
		Expect(host.Close(ctx)).To(Succeed())

		want := map[string]string{}
		Expect(filepath.WalkDir(host.Dir(), func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			rel, err := filepath.Rel(host.Dir(), path)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			want["traces/"+host.RunID()+"/"+filepath.ToSlash(rel)] = string(data)
			return err
		})).To(Succeed())
		Expect(want).ToNot(BeEmpty(), "a closed store leaves files on disk")

		bucket := blob.Local(GinkgoT().TempDir())
		Expect(host.Persist(ctx, bucket)).To(Succeed())

		objects, err := bucket.List(ctx, "traces/")
		Expect(err).ToNot(HaveOccurred())
		got := map[string]string{}
		for _, object := range objects {
			reader, err := bucket.Read(ctx, object.Key)
			Expect(err).ToNot(HaveOccurred())
			data, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(reader.Close()).To(Succeed())
			got[object.Key] = string(data)
		}
		Expect(got).To(Equal(want))
		Expect(host.Dir()).To(BeADirectory(), "the local store stays for the run's snapshot")
	})

	It("does nothing without a configured store", func(ctx context.Context) {
		host := openHost(GinkgoT().TempDir())
		Expect(host.Persist(ctx, nil)).To(Succeed())
	})
})
