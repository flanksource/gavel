package verify

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type decodeEntry struct {
	Count int `json:"count" yaml:"count"`
}

type decodeDocument struct {
	Name    string                 `json:"name" yaml:"name"`
	Entries []decodeEntry          `json:"entries" yaml:"entries"`
	Named   map[string]decodeEntry `json:"named" yaml:"named"`
	Free    map[string]any         `json:"free" yaml:"free"`
	Spec    api.Spec               `json:"spec" yaml:"spec"`
}

type decodeCustom struct {
	decodeEntry `json:",inline" yaml:",inline"`
}

func (d *decodeCustom) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &d.decodeEntry)
}

func (decodeCustom) DecodeFields() any { return decodeEntry{} }

var _ = Describe("unknown field decoding", func() {
	var warnings []string
	var options DecodeOptions
	BeforeEach(func() {
		warnings = nil
		options = DecodeOptions{Source: "project.yaml", Warn: func(message string) { warnings = append(warnings, message) }}
	})

	DescribeTable("reports all unknown fields without discarding known fields",
		func(decode func([]byte, any, DecodeOptions) error, document string) {
			var target decodeDocument
			Expect(decode([]byte(document), &target, options)).To(Succeed())
			Expect(target.Name).To(Equal("example"))
			Expect(target.Entries).To(Equal([]decodeEntry{{Count: 3}}))
			Expect(warnings).To(Equal([]string{
				"project.yaml: unknown field entries[0].future",
				"project.yaml: unknown field named.release.extra",
				"project.yaml: unknown field obsolete",
			}))
		},
		Entry("JSON", DecodeJSON, `{"name":"example","entries":[{"count":3,"future":null}],"named":{"release":{"extra":true}},"free":{"arbitrary":{"nested":true}},"obsolete":1}`),
		Entry("YAML", DecodeYAML, "name: example\nentries:\n  - count: 3\n    future: null\nnamed:\n  release:\n    extra: true\nfree:\n  arbitrary:\n    nested: true\nobsolete: 1\n"),
	)

	It("can reject unknown fields with the same source and paths", func() {
		options.Policy = UnknownFieldsError
		var target decodeDocument
		err := DecodeJSON([]byte(`{"entries":[{"misspelled":1}],"unknown":true}`), &target, options)
		Expect(err).To(MatchError("project.yaml: unknown field entries[0].misspelled\nproject.yaml: unknown field unknown"))
		Expect(warnings).To(BeEmpty())
	})

	DescribeTable("keeps syntax and type errors fatal",
		func(decode func([]byte, any, DecodeOptions) error, document string) {
			var target decodeDocument
			Expect(decode([]byte(document), &target, options)).NotTo(Succeed())
			Expect(warnings).To(BeEmpty())
		},
		Entry("JSON type", DecodeJSON, `{"entries":[{"count":"many"}],"ignored":true}`),
		Entry("JSON syntax", DecodeJSON, `{"name":`),
		Entry("JSON trailing document", DecodeJSON, `{} {}`),
		Entry("YAML type", DecodeYAML, "entries:\n  - count: many\nignored: true\n"),
		Entry("YAML syntax", DecodeYAML, "name: ["),
		Entry("YAML duplicate keys", DecodeYAML, "name: first\nname: second\n"),
		Entry("YAML trailing document", DecodeYAML, "name: first\n---\nname: second\n"),
	)

	It("uses an explicit custom wire shape to inspect nested fields", func() {
		var target struct {
			Custom decodeCustom `json:"custom"`
		}
		Expect(DecodeJSON([]byte(`{"custom":{"count":3,"extra":true}}`), &target, options)).To(Succeed())
		Expect(target.Custom.Count).To(Equal(3))
		Expect(warnings).To(Equal([]string{"project.yaml: unknown field custom.extra"}))
	})

	It("inspects embedded spec fields using their flattened JSON names", func() {
		var target decodeDocument
		Expect(DecodeJSON([]byte(`{"spec":{"model":"cli:sonnet","permissions":{"typo":true}}}`), &target, options)).To(Succeed())
		Expect(target.Spec.Name).To(Equal("cli:sonnet"))
		Expect(warnings).To(Equal([]string{"project.yaml: unknown field spec.permissions.typo"}))
	})

	It("inspects YAML merges at each typed destination without duplicate warnings", func() {
		var target decodeDocument
		document := "entries:\n  - &entry\n    count: 3\n    typo: true\n  - <<: *entry\nnamed:\n  release: *entry\n"
		Expect(DecodeYAML([]byte(document), &target, options)).To(Succeed())
		Expect(target.Entries).To(Equal([]decodeEntry{{Count: 3}, {Count: 3}}))
		Expect(warnings).To(Equal([]string{
			"project.yaml: unknown field entries[0].typo",
			"project.yaml: unknown field entries[1].typo",
			"project.yaml: unknown field named.release.typo",
		}))
	})

	It("rejects an invalid policy before decoding", func() {
		options.Policy = UnknownFieldPolicy(99)
		var target decodeDocument
		Expect(DecodeJSON([]byte(`{}`), &target, options)).To(MatchError(ContainSubstring("invalid unknown field policy")))
	})

	It("respects dynamic keys in Captain's custom MCP decoder", func() {
		var target decodeDocument
		Expect(DecodeJSON([]byte(`{"spec":{"permissions":{"mcp":{"example":"enabled"}}}}`), &target, options)).To(Succeed())
		Expect(warnings).To(BeEmpty())
		Expect(target.Spec.Permissions.MCP.Modes).To(HaveKey("example"))
	})

	It("preserves strict sandbox decoder failures", func() {
		var target decodeDocument
		err := DecodeJSON([]byte(`{"spec":{"sandbox":{"mode":"native","typo":true}}}`), &target, options)
		Expect(err).To(MatchError(ContainSubstring(`unknown field "typo"`)))
		Expect(warnings).To(BeEmpty())
	})
})
