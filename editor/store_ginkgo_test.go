package editor

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditor(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Editor") }

var _ = Describe("File store", func() {
	DescribeTable("accepts browser templates as editable documents", func(extension string) {
		name := filepath.Join(GinkgoT().TempDir(), "page"+extension)
		Expect(os.WriteFile(name, []byte("export default () => <h1>Hello</h1>"), 0600)).To(Succeed())
		for _, input := range []string{name, filepath.Dir(name)} {
			store, err := NewStore([]string{input})
			Expect(err).NotTo(HaveOccurred())
			Expect(store.Files()).To(HaveLen(1))
			doc, err := store.Read(store.Files()[0].ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(doc.Kind).To(Equal(KindDocument))
			doc.Source = "export default () => <h1>Updated</h1>"
			_, err = store.Write(doc)
			Expect(err).NotTo(HaveOccurred())
		}
	}, Entry("TSX", ".tsx"), Entry("JSX", ".jsx"), Entry("MDX", ".mdx"))
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
	It("lists an svg as kind image and reads and writes its text source", func() {
		const svg = `<svg xmlns="http://www.w3.org/2000/svg"><rect width="4" height="4"/></svg>`
		image := filepath.Join(dir, "logo.svg")
		Expect(os.WriteFile(image, []byte(svg), 0o600)).To(Succeed())
		s, err := NewStore([]string{image})
		Expect(err).NotTo(HaveOccurred())
		file := s.Files()[0]
		doc, err := s.Read(file.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{doc.Kind, doc.Source}).To(Equal([]string{"image", svg}))
		edited := strings.Replace(svg, `width="4"`, `width="8"`, 1)
		_, err = s.Write(Document{File: file, Source: edited, Revision: doc.Revision})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.ReadFile(image)).To(Equal([]byte(edited)))
	})
	It("rejects missing and unsupported inputs", func() {
		_, err := NewStore(nil)
		Expect(err).To(HaveOccurred())
		_, err = NewStore([]string{filepath.Join(dir, "missing")})
		Expect(err).To(HaveOccurred())
	})
})
