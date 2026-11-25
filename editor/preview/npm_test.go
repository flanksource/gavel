package preview

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPreview(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Preview packages") }

var _ = Describe("Npm extraction", func() {
	It("rejects a stale local build even when its previous generation is cached", func() {
		local := GinkgoT().TempDir()
		cache := &Cache{Root: GinkgoT().TempDir()}
		Expect(os.MkdirAll(filepath.Join(local, "dist/runtime"), 0700)).To(Succeed())
		content := []byte("export const value = 1;")
		hash := sha512.Sum384(content)
		fragment, err := json.Marshal(map[string]any{"integrity": map[string]string{"example-library@1.0.0/dist/index.js": "sha384-" + base64.StdEncoding.EncodeToString(hash[:])}})
		Expect(err).NotTo(HaveOccurred())
		for name, data := range map[string][]byte{"package.json": []byte(`{"name":"example-library","version":"1.0.0"}`), "dist/index.js": content, "dist/runtime/importmap.json": fragment, "dist/runtime/barrels.json": []byte(`{}`)} {
			Expect(os.WriteFile(filepath.Join(local, name), data, 0600)).To(Succeed())
		}
		manifest := Manifest{Name: "example-library", Version: "1.0.0"}
		_, err = cache.snapshot(local, manifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(local, "dist/index.js"), []byte("changed"), 0600)).To(Succeed())
		_, err = cache.snapshot(local, manifest)
		Expect(err).To(MatchError(ContainSubstring("stale runtime file")))
	})
	It("rejects tampered tarball bytes", func() {
		data := []byte("verified archive")
		hash := sha512.Sum512(data)
		integrity := "sha512-" + base64.StdEncoding.EncodeToString(hash[:])
		Expect(validIntegrity(data, integrity)).To(BeTrue())
		Expect(validIntegrity([]byte("changed archive"), integrity)).To(BeFalse())
	})
	It("resolves cached metadata offline without losing ESM exports or dependency declarations", func() {
		cache := &Cache{Root: GinkgoT().TempDir()}
		name := "example-library"
		path := filepath.Join(cache.Root, "metadata", fmt.Sprintf("%x.json", sha256.Sum256([]byte(name))))
		Expect(os.MkdirAll(filepath.Dir(path), 0700)).To(Succeed())
		data := []byte(`{"versions":{"1.0.0":{"name":"example-library","version":"1.0.0","type":"module","exports":{".":"./index.js"},"dependencies":{"example-dependency":"^2.0.0"}}},"dist-tags":{"latest":"1.0.0"}}`)
		Expect(os.WriteFile(path, data, 0600)).To(Succeed())
		manifest, err := cache.Resolve(context.Background(), name, "^1.0.0")
		Expect(err).NotTo(HaveOccurred())
		Expect(manifest.Version).To(Equal("1.0.0"))
		Expect(string(manifest.Exports)).To(Equal(`{".":"./index.js"}`))
		Expect(manifest.Dependencies).To(HaveKeyWithValue("example-dependency", "^2.0.0"))
	})
	DescribeTable("validates tar paths and entry types", func(name string, kind byte, valid bool) {
		var archive bytes.Buffer
		gzipWriter := gzip.NewWriter(&archive)
		writer := tar.NewWriter(gzipWriter)
		header := &tar.Header{Name: name, Typeflag: kind, Mode: 0600}
		if kind == tar.TypeReg {
			header.Size = int64(len("example"))
		}
		Expect(writer.WriteHeader(header)).To(Succeed())
		if kind == tar.TypeReg {
			_, err := writer.Write([]byte("example"))
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(writer.Close()).To(Succeed())
		Expect(gzipWriter.Close()).To(Succeed())
		root := GinkgoT().TempDir()
		err := extract(archive.Bytes(), root)
		if valid {
			Expect(err).NotTo(HaveOccurred())
			if kind == tar.TypeReg {
				Expect(os.ReadFile(filepath.Join(root, "package.json"))).To(Equal([]byte("example")))
			}
		} else {
			Expect(err).To(HaveOccurred())
		}
	}, Entry("regular package", "package/package.json", byte(tar.TypeReg), true),
		Entry("root directory", "package/", byte(tar.TypeDir), true),
		Entry("traversal", "package/../secret", byte(tar.TypeReg), false),
		Entry("absolute path", "/package/package.json", byte(tar.TypeReg), false),
		Entry("symlink", "package/link", byte(tar.TypeSymlink), false))
})
