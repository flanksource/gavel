package ui

import (
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Captain compiles every caller tool's input schema against the JSON Schema
// 2020-12 metaschema when it admits a chat turn, so one malformed property
// (comments_acknowledge's raw JSON content once published "type": "") fails
// every message the assistant is sent, not just calls to that tool.
var _ = Describe("assistant tool schemas", func() {
	It("publishes every tool input schema valid against JSON Schema 2020-12", func() {
		root, err := entityCommandRoot()
		Expect(err).NotTo(HaveOccurred())
		provider, err := newGavelChatTools(root)
		Expect(err).NotTo(HaveOccurred())
		tools, err := provider.ToolSet(GinkgoT().Context())
		Expect(err).NotTo(HaveOccurred())

		names := make([]string, 0, len(tools.Definitions))
		for _, definition := range tools.Definitions {
			names = append(names, definition.Name)
			raw, err := json.Marshal(definition.InputSchema)
			Expect(err).NotTo(HaveOccurred(), definition.Name)
			document, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
			Expect(err).NotTo(HaveOccurred(), definition.Name)
			compiler := jsonschema.NewCompiler()
			url := "mem:///gavel/chat-tools/" + definition.Name + "/input-schema.json"
			Expect(compiler.AddResource(url, document)).To(Succeed(), definition.Name)
			_, err = compiler.Compile(url)
			Expect(err).NotTo(HaveOccurred(), "%s schema: %s", definition.Name, raw)
		}
		Expect(names).To(ContainElements("comments_acknowledge", "comments_reply", "comments_create", "todo", "project"))
	})
})
