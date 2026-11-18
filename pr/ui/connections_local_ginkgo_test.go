// Tests the connection routes of a local dashboard: they keep the local organization's encrypted
// catalog in the database SetConnectionCatalog gives the Server.
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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("a local dashboard's connection catalog", func() {
	call := func(server *Server, method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req = req.WithContext(tenant.WithScope(req.Context(), tenant.Local()))
		mux := http.NewServeMux()
		for _, pattern := range []string{"GET /api/connections", "POST /api/connections"} {
			mux.HandleFunc(pattern, server.handleConnections)
		}
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, req)
		return response
	}

	It("keeps git connections in the local organization's catalog", func() {
		GinkgoT().Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "local_kms://"+base64.URLEncoding.EncodeToString(bytes.Repeat([]byte{2}, 32)))
		db := dbtest.ForGinkgo(dbtest.Options{Name: "connections_local"}).Gorm()
		Expect(db.AutoMigrate(&catalog.Record{})).To(Succeed())
		server := &Server{}
		server.SetConnectionCatalog(db)

		created := call(server, "POST", "/api/connections", `{"name":"acme","type":"github","password":"gh_local","properties":{"organization":"acme"}}`)
		Expect(created.Code).To(Equal(200), created.Body.String())
		Expect(created.Body.String()).NotTo(ContainSubstring("gh_local"))

		var row catalog.Record
		Expect(db.First(&row).Error).To(Succeed())
		Expect(row.OrgID).To(Equal(tenant.LocalOrgID))
		Expect(row.Type).To(Equal("github"))

		listed := call(server, "GET", "/api/connections", "")
		Expect(listed.Code).To(Equal(200), listed.Body.String())
		var inventory any
		Expect(json.Unmarshal(listed.Body.Bytes(), &inventory)).To(Succeed())
		Expect(listed.Body.String()).To(ContainSubstring(`"acme"`))
	})

	It("needs a database", func() {
		response := call(&Server{}, "GET", "/api/connections", "")
		Expect(response.Code).To(Equal(http.StatusServiceUnavailable))
		Expect(response.Body.String()).To(ContainSubstring("connections require a database"))
	})
})
