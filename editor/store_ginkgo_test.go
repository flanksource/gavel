package editor

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
	"testing"
)

func TestEditor(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Editor") }

var _ = Describe("File store", func() {
	It("keeps the caller's repository path when a selected directory is a symlink", func() {
		real := GinkgoT().TempDir()
		parent := GinkgoT().TempDir()
		link := filepath.Join(parent, "linked")
		Expect(os.Symlink(real, link)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(real, "config.yaml"), []byte("name: example"), 0600)).To(Succeed())
		store, err := NewStore([]string{link})
		Expect(err).NotTo(HaveOccurred())
		Expect(store.Files()[0].Path).To(Equal(filepath.Join(link, "config.yaml")))
	})
	var dir, path string
	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		path = filepath.Join(dir, "config.yaml")
		Expect(os.WriteFile(path, []byte("# comment\nname: initial\n"), 0600)).To(Succeed())
	})
	It("deduplicates overlapping inputs and writes exact source with existing permissions", func() {
		s, err := NewStore([]string{dir, path})
		Expect(err).NotTo(HaveOccurred())
		Expect(s.Files()).To(HaveLen(1))
		doc, err := s.Read(s.Files()[0].ID)
		Expect(err).NotTo(HaveOccurred())
		doc.Source = "invalid: [\n"
		saved, err := s.Write(doc)
		Expect(err).NotTo(HaveOccurred())
		Expect(saved.Source).To(Equal(doc.Source))
		Expect(saved.Revision).NotTo(Equal(doc.Revision))
		info, err := os.Stat(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0600)))
	})
	It("rejects stale saves without changing the external edit", func() {
		s, err := NewStore([]string{path})
		Expect(err).NotTo(HaveOccurred())
		doc, err := s.Read(s.Files()[0].ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(path, []byte("external"), 0600)).To(Succeed())
		doc.Source = "draft"
		_, err = s.Write(doc)
		Expect(err).To(MatchError(ErrConflict))
		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("external"))
	})
	It("rejects a discovered path replaced by a symlink outside the root", func() {
		s, err := NewStore([]string{dir})
		Expect(err).NotTo(HaveOccurred())
		outside := filepath.Join(GinkgoT().TempDir(), "secret.yaml")
		Expect(os.WriteFile(outside, []byte("private"), 0600)).To(Succeed())
		Expect(os.Remove(path)).To(Succeed())
		Expect(os.Symlink(outside, path)).To(Succeed())
		_, err = s.Read(s.Files()[0].ID)
		Expect(err).To(HaveOccurred())
	})
	It("lists an image in a folder as kind image and refuses to read it as a document", func() {
		Expect(os.MkdirAll(filepath.Join(dir, "img"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "img", "a.png"), []byte("\x89PNG\x00"), 0o600)).To(Succeed())
		s, err := NewStore([]string{dir})
		Expect(err).NotTo(HaveOccurred())
		kinds := map[string]string{}
		for _, f := range s.Files() {
			kinds[f.Path] = f.Kind
		}
		image := filepath.Join(dir, "img", "a.png")
		Expect(kinds).To(Equal(map[string]string{path: "document", image: "image"}))
		var id string
		for _, f := range s.Files() {
			if f.Kind == "image" {
				id = f.ID
			}
		}
		_, err = s.Read(id)
		Expect(err).To(MatchError(image + " is an image; it has no editable source"))
	})
	It("rejects missing and unsupported inputs", func() {
		_, err := NewStore(nil)
		Expect(err).To(HaveOccurred())
		_, err = NewStore([]string{filepath.Join(dir, "missing")})
		Expect(err).To(HaveOccurred())
	})
})
