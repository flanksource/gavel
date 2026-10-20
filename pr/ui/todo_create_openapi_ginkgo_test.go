package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create TODO API documentation", func() {
	var paths, schemas map[string]any
	BeforeEach(func() {
		recorder := httptest.NewRecorder()
		(&Server{}).handleOpenAPI(recorder, httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil))
		Expect(recorder.Code).To(Equal(http.StatusOK))
		var document map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &document)).To(Succeed())
		paths = document["paths"].(map[string]any)
		schemas = document["components"].(map[string]any)["schemas"].(map[string]any)
	})
	schemaFor := func(content map[string]any) map[string]any {
		ref := content["application/json"].(map[string]any)["schema"].(map[string]any)["$ref"].(string)
		return schemas[strings.TrimPrefix(ref, "#/components/schemas/")].(map[string]any)
	}

	It("serves the optional step and the partial-success run response from the wire types", func() {
		operation := paths["/api/todos/new"].(map[string]any)["post"].(map[string]any)
		content := operation["requestBody"].(map[string]any)["content"].(map[string]any)
		Expect(content).To(HaveKey("multipart/form-data"))
		request := schemaFor(content)
		Expect(request["required"]).To(ContainElement("title"))
		Expect(request["required"]).NotTo(ContainElement("step"))
		Expect(request["properties"].(map[string]any)["step"]).To(HaveKeyWithValue("type", "string"))
		created := operation["responses"].(map[string]any)["201"].(map[string]any)
		Expect(schemaFor(created["content"].(map[string]any))["properties"]).To(HaveKey("run"))
	})

	It("serves the new-todo steps listing", func() {
		operation := paths["/api/todos/new/steps"].(map[string]any)["get"].(map[string]any)
		ok := operation["responses"].(map[string]any)["200"].(map[string]any)
		Expect(schemaFor(ok["content"].(map[string]any))["properties"]).To(HaveKey("steps"))
	})
})
