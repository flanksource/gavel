package jira

import (
	"context"
	"github.com/flanksource/gavel/externalissues"
	"github.com/flanksource/gavel/externalissues/syncdoc"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("sync document metadata", func() {
	It("retains enrichment results while marking removed, changed and unverified inputs stale", func() {
		root, err := os.OpenRoot(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(root.Close)
		for _, scenario := range []struct {
			present bool
			hash    string
			input   string
			stale   bool
		}{
			{true, "original", "original", false}, {false, "original", "original", true}, {true, "changed", "original", true}, {true, "original", "", true},
		} {
			doc := &syncdoc.Document{Metadata: syncdoc.Metadata{Sync: syncdoc.Settings{Attachments: true}, Attachments: map[string]*syncdoc.Attachment{"asset": {Present: scenario.present, Local: &syncdoc.Artifact{SHA256: scenario.hash}, Enrichments: map[string]syncdoc.Enrichment{"ocr": {SourceSHA256: scenario.input, Extra: map[string]yaml.Node{"processor": {Kind: yaml.ScalarNode, Tag: "!!str", Value: "retained"}}}}}}}}
			finishSyncMetadata(doc, root)
			Expect(doc.Metadata.Attachments["asset"].Enrichments["ocr"].Stale).To(Equal(scenario.stale))
			Expect(doc.Metadata.Attachments["asset"].Enrichments["ocr"].Extra["processor"].Value).To(Equal("retained"))
		}
	})
	It("records a source identity and reloads saved settings", func() {
		dir := GinkgoT().TempDir()
		file := filepath.Join(dir, "issue.md")
		Expect(os.WriteFile(file, []byte("---\ngavel:\n  version: 1\n  source:\n    provider: jira\n    site: https://example.test\n    id: '10001'\n    key: EXAMPLE-1\n  sync:\n    attachments: true\n    extract: true\n---\nBody\n"), 0600)).To(Succeed())
		options, site, err := ResolveSyncFile(SyncOptions{From: file}, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(site).To(Equal("https://example.test"))
		Expect(options.Keys).To(Equal([]string{"10001"}))
		Expect(options.Attachments && options.Extract).To(BeTrue())
		Expect(options.Output).To(Equal(file))
		_, _, err = ResolveSyncFile(SyncOptions{From: file}, "https://other.test")
		Expect(err).To(HaveOccurred())
	})
	It("rejects missing or unterminated metadata before network access", func() {
		for _, body := range []string{"Body", "---\ngavel: {}\n", "---\ngavel:\n  version: 2\n---\n"} {
			file := filepath.Join(GinkgoT().TempDir(), "issue.md")
			Expect(os.WriteFile(file, []byte(body), 0600)).To(Succeed())
			_, _, err := ResolveSyncFile(SyncOptions{From: file}, "")
			Expect(err).To(HaveOccurred())
		}
	})
	It("reuses extraction only when all artifact and pipeline hashes verify", func() {
		dir := GinkgoT().TempDir()
		root, err := os.OpenRoot(dir)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(root.Close)
		image := documentTestImage()
		imagePath := "attachments/cas/" + syncdoc.Hash(image) + ".png"
		Expect(root.MkdirAll("attachments/cas", 0700)).To(Succeed())
		Expect(root.WriteFile(imagePath, image, 0600)).To(Succeed())
		body := "Cached text\n\n![Image](" + imagePath + ")"
		cached, err := rebaseExtraction(body, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(cached).To(ContainSubstring("../cas/"))
		Expect(root.MkdirAll("attachments/extracted", 0700)).To(Succeed())
		path := "attachments/extracted/" + syncdoc.Hash([]byte(cached)) + ".md"
		Expect(root.WriteFile(path, []byte(cached), 0600)).To(Succeed())
		record := &syncdoc.Attachment{Local: &syncdoc.Artifact{SHA256: "source"}, Extraction: &syncdoc.Extraction{SourceSHA256: "source", PipelineFingerprint: extractionFingerprint("docx"), Markdown: syncdoc.Artifact{Path: path, SHA256: syncdoc.Hash([]byte(cached))}, Images: []syncdoc.Artifact{{Path: imagePath, SHA256: syncdoc.Hash(image)}}}}
		options := syncExtractionOptions{Attachment: externalissues.Attachment{Name: "report.docx"}, Record: record, Root: root, Stage: GinkgoT().TempDir(), Path: filepath.Join(dir, "missing.docx"), Result: &SyncResult{Path: filepath.Join(dir, "issue.md")}}
		output, err := syncExtraction(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(output.Markdown).To(ContainSubstring(imagePath))
		Expect(root.Remove(imagePath)).To(Succeed())
		_, err = syncExtraction(context.Background(), options)
		Expect(err).To(HaveOccurred())
		Expect(root.WriteFile(imagePath, image, 0600)).To(Succeed())
		options.Force = true
		_, err = syncExtraction(context.Background(), options)
		Expect(err).To(HaveOccurred())
		options.Force = false
		record.Extraction.PipelineFingerprint = "old"
		_, err = syncExtraction(context.Background(), options)
		Expect(err).To(HaveOccurred())
	})
})
