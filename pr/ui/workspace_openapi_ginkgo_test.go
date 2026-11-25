package ui

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("workspace change API schema", func() {
	It("documents the stream without publishing internal fingerprints", func() {
		document := map[string]any{"paths": map[string]any{}, "components": map[string]any{"schemas": map[string]any{}}}
		Expect(addWorkspaceOpenAPI(document)).To(Succeed())
		Expect(document["paths"]).To(HaveKey("/api/todos/changes"))
		encoded, err := json.Marshal(document)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(ContainSubstring("text/event-stream"))
		Expect(string(encoded)).NotTo(ContainSubstring("Fingerprint"))
	})
})
