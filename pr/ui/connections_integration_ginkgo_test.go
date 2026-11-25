package ui

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/flanksource/commons-db/connection/catalog"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("connection catalog API", func() {
	It("scopes CRUD, redacts credentials and preserves secrets during partial edits", func() {
		GinkgoT().Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "local_kms://"+base64.URLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32)))
		db := dbtest.ForGinkgo(dbtest.Options{Name: "connections_api"}).Gorm()
		Expect(db.AutoMigrate(&catalog.Record{})).To(Succeed())
		org := uuid.New()
		server := &Server{}
		server.SetOrg(org, db)
		scope := tenant.Scope{OrgID: org, Role: "org:admin"}
		call := func(method, path, body string) *httptest.ResponseRecorder {
			req := httptest.NewRequest(method, path, strings.NewReader(body))
			req = req.WithContext(tenant.WithScope(req.Context(), scope))
			mux := http.NewServeMux()
			mux.HandleFunc("GET /api/connections/storage", server.handleConnectionStorage)
			for _, pattern := range []string{"GET /api/connections", "POST /api/connections", "GET /api/connections/{reference}", "PATCH /api/connections/{reference}", "DELETE /api/connections/{reference}"} {
				mux.HandleFunc(pattern, server.handleConnections)
			}
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, req)
			return response
		}
		status := call("GET", "/api/connections/storage", "")
		Expect(status.Code).To(Equal(200))
		var storage struct {
			KEK struct {
				State string `json:"state"`
			} `json:"kek"`
		}
		Expect(json.Unmarshal(status.Body.Bytes(), &storage)).To(Succeed())
		Expect(storage.KEK.State).To(Equal("ready"))
		Expect(status.Body.String()).NotTo(ContainSubstring(base64.URLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))))
		created := call("POST", "/api/connections", `{"name":"example","type":"http","url":"https://example.com","password":"credential-value","properties":{"token":"property-secret","region":"east"}}`)
		Expect(created.Code).To(Equal(200), created.Body.String())
		Expect(created.Body.String()).NotTo(ContainSubstring("credential-value"))
		Expect(created.Body.String()).NotTo(ContainSubstring("property-secret"))
		var value struct {
			ID string `json:"id"`
		}
		Expect(json.Unmarshal(created.Body.Bytes(), &value)).To(Succeed())
		patched := call("PATCH", "/api/connections/"+value.ID, `{"name":"renamed","properties":{"region":"west"}}`)
		Expect(patched.Code).To(Equal(200), patched.Body.String())
		var raw catalog.Record
		Expect(db.First(&raw).Error).To(Succeed())
		Expect(bytes.Contains(raw.Payload, []byte("credential-value"))).To(BeFalse())
		scope.Role = "org:member"
		Expect(call("DELETE", "/api/connections/"+value.ID, "").Code).To(Equal(403))
		scope.OrgID = uuid.New()
		Expect(call("GET", "/api/connections/storage", "").Code).To(Equal(403))
		Expect(call("GET", "/api/connections/"+value.ID, "").Code).To(Equal(403))
		scope = tenant.Scope{OrgID: org, Role: "org:admin"}
		Expect(call("DELETE", "/api/connections/"+value.ID, "").Code).To(Equal(204))
	})
})
