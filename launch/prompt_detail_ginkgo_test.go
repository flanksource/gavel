package launch

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons-db/shell"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("public prompt form", func() {
	It("redacts nested and referenced private defaults while preserving input controls", func() {
		const secret = "private-form-canary"
		detail := captaincli.PromptDetail{
			Content: secret,
			InputSchema: map[string]any{"type": "object", "$defs": map[string]any{"secret": map[string]any{"type": "string", "writeOnly": true}}, "properties": map[string]any{
				"name":     map[string]any{"type": "string", "default": "visible"},
				"token":    map[string]any{"$ref": "#/$defs/secret", "default": secret},
				"combined": map[string]any{"allOf": []any{map[string]any{"type": "string"}, map[string]any{"writeOnly": true}}, "default": secret},
				"nested":   map[string]any{"type": "object", "properties": map[string]any{"password": map[string]any{"type": "string", "writeOnly": true, "default": secret}}},
			}},
			InputDefault: map[string]any{"name": "visible", "token": secret, "combined": secret, "nested": map[string]any{"password": secret}},
			Run:          captaincli.PromptRenderRequest{Variables: map[string]any{"token": secret}, Spec: &api.Spec{Setup: &shell.Setup{Env: []string{"TOKEN=" + secret}}}},
		}
		form, err := promptFormDetail(detail)
		Expect(err).NotTo(HaveOccurred())
		data, err := json.Marshal(form)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(ContainSubstring(secret))
		Expect(string(data)).To(ContainSubstring("writeOnly"))
		Expect(string(data)).To(ContainSubstring("visible"))
	})
})
