package jsonb

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJSONB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "jsonb encoding")
}

// literalEscape is the six-character text a user's output can contain without
// any NUL being involved: backslash, 'u', '0', '0', '0', '0'.
const literalEscape = `\u0000`

// stripValue marshals value, strips it, and reports both the stripped document
// and what it decodes back to.
func stripValue(value any) ([]byte, any) {
	GinkgoHelper()
	encoded, err := json.Marshal(value)
	Expect(err).NotTo(HaveOccurred())
	stripped := StripNullEscapes(encoded)
	var decoded any
	Expect(json.Unmarshal(stripped, &decoded)).To(Succeed(), "stripping produced invalid JSON: %s", stripped)
	return stripped, decoded
}

var _ = Describe("StripNullEscapes", func() {
	It("removes a NUL from a string value", func() {
		_, decoded := stripValue(map[string]string{"stdout": "before\x00after"})

		Expect(decoded).To(Equal(map[string]any{"stdout": "beforeafter"}))
	})

	It("removes a NUL from an object key", func() {
		_, decoded := stripValue(map[string]string{"a\x00b": "value"})

		Expect(decoded).To(Equal(map[string]any{"ab": "value"}))
	})

	It("preserves text that merely looks like the escape", func() {
		stripped, decoded := stripValue(map[string]string{"stdout": "log: " + literalEscape + " emitted"})

		Expect(decoded).To(Equal(map[string]any{"stdout": "log: " + literalEscape + " emitted"}))
		Expect(string(stripped)).To(ContainSubstring(`\\u0000`))
	})

	It("keeps a real backslash that immediately precedes a NUL", func() {
		_, decoded := stripValue(map[string]string{"stdout": `C:\` + "\x00" + "path"})

		Expect(decoded).To(Equal(map[string]any{"stdout": `C:\path`}))
	})

	It("strips through nested objects and arrays", func() {
		_, decoded := stripValue(map[string]any{
			"snapshots": []any{
				map[string]string{"stderr": "a\x00b"},
				map[string]any{"logs": []string{"c\x00d"}},
			},
		})

		Expect(decoded).To(Equal(map[string]any{
			"snapshots": []any{
				map[string]any{"stderr": "ab"},
				map[string]any{"logs": []any{"cd"}},
			},
		}))
	})

	It("leaves a document without the escape untouched", func() {
		encoded := []byte(`{"stdout":"clean output","count":3}`)

		Expect(string(StripNullEscapes(encoded))).To(Equal(string(encoded)))
	})

	It("does not treat a quote inside a string as the end of it", func() {
		_, decoded := stripValue(map[string]string{"stdout": `say "hi"` + "\x00" + "!"})

		Expect(decoded).To(Equal(map[string]any{"stdout": `say "hi"!`}))
	})

	It("removes several NULs in one string", func() {
		_, decoded := stripValue(map[string]string{"stdout": "\x00a\x00b\x00"})

		Expect(decoded).To(Equal(map[string]any{"stdout": "ab"}))
	})
})
