package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Triage new API documentation", func() {
	It("serves the creation opt-in and partial-success response from the wire types", func() {
		recorder := httptest.NewRecorder()
		(&Server{}).handleOpenAPI(recorder, httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil))
		Expect(recorder.Code).To(Equal(http.StatusOK))
		var document map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &document)).To(Succeed())
		paths := document["paths"].(map[string]any)
		Expect(paths).To(HaveKey("/api/todos/new"))
		operation := paths["/api/todos/new"].(map[string]any)["post"].(map[string]any)
		content := operation["requestBody"].(map[string]any)["content"].(map[string]any)
		Expect(content).To(HaveKey("application/json"))
		Expect(content).To(HaveKey("multipart/form-data"))
		ref := content["application/json"].(map[string]any)["schema"].(map[string]any)["$ref"].(string)
		schemas := document["components"].(map[string]any)["schemas"].(map[string]any)
		request := schemas[strings.TrimPrefix(ref, "#/components/schemas/")].(map[string]any)
		Expect(request["required"]).To(ContainElement("title"))
		Expect(request["required"]).NotTo(ContainElement("triage"))
		triage := request["properties"].(map[string]any)["triage"].(map[string]any)
		Expect(triage["type"]).To(Equal("boolean"))
		response := operation["responses"].(map[string]any)["201"].(map[string]any)
		Expect(response["description"]).To(ContainSubstring("triage admission fails"))
	})
})
