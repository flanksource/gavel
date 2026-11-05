package launch

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLaunch(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Launch") }

type schemaFixture struct{}

func (schemaFixture) Schema(kind string) (map[string]any, error) {
	return object(map[string]any{"paths": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}, "files": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}}), nil
}

var _ = Describe("launch schema projections", func() {
	request := func(config map[string]any) []byte {
		raw, err := json.Marshal(map[string]any{"workspace": map[string]any{"cwd": "/workspace"}, "config": config})
		Expect(err).NotTo(HaveOccurred())
		return raw
	}
	It("accepts combined work at the root and rejects it at both children", func() {
		body := request(map[string]any{"test": map[string]any{}, "lint": map[string]any{}})
		_, err := Decode(body, "", schemaFixture{})
		Expect(err).NotTo(HaveOccurred())
		for _, kind := range []string{"test", "lint"} {
			_, err := Decode(body, kind, schemaFixture{})
			Expect(err).To(HaveOccurred(), kind)
		}
	})
	It("requires work and the section named by the child", func() {
		for _, kind := range []string{"", "test", "lint"} {
			_, err := Decode(request(map[string]any{}), kind, schemaFixture{})
			Expect(err).To(HaveOccurred())
		}
		_, err := Decode(request(map[string]any{"lint": map[string]any{}}), "lint", schemaFixture{})
		Expect(err).NotTo(HaveOccurred())
	})
	It("validates multiple prompt configurations without accepting unknown options", func() {
		entries := []any{
			map[string]any{"id": "design", "linter": "prompt", "options": map[string]any{"prompt": "diagram-review", "promptVersion": "v1"}},
			map[string]any{"id": "accessibility", "linter": "prompt", "options": map[string]any{"prompt": "a11y-review", "promptVersion": "v2"}},
		}
		_, err := Decode(request(map[string]any{"lint": map[string]any{"linters": entries}}), "lint", schemaFixture{})
		Expect(err).NotTo(HaveOccurred())
		entries = append(entries, map[string]any{"id": "design", "linter": "vale"})
		_, err = Decode(request(map[string]any{"lint": map[string]any{"linters": entries}}), "lint", schemaFixture{})
		Expect(err).To(MatchError(ContainSubstring("duplicate")))
	})
	DescribeTable("prompt spec overrides", func(spec map[string]any, valid bool) {
		options := map[string]any{"prompt": "spec-public-high", "promptVersion": "v1", "spec": spec}
		body := request(map[string]any{"lint": map[string]any{"linters": []any{
			map[string]any{"id": "review", "linter": "prompt", "options": options},
		}}})
		for _, kind := range []string{"", "lint"} {
			decoded, err := Decode(body, kind, schemaFixture{})
			if valid {
				Expect(err).NotTo(HaveOccurred(), kind)
				Expect(decoded.Config.Lint.Linters[0].Options).To(Equal(options))
			} else {
				Expect(err).To(HaveOccurred(), kind)
			}
		}
	},
		Entry("accepts runtime settings without an inline prompt", map[string]any{"model": "example-model", "mode": "cli", "effort": "low"}, true),
		Entry("accepts overrides without repeating the saved model", map[string]any{"effort": "low"}, true),
		Entry("accepts empty overrides", map[string]any{}, true),
		Entry("rejects unknown override fields", map[string]any{"unknown": true}, false),
		Entry("rejects invalid runtime types", map[string]any{"model": true}, false),
		Entry("validates nested override fields", map[string]any{"budget": map[string]any{"maxTokens": "many"}}, false),
	)
})
