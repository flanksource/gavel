package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/flanksource/commons-db/connection"
	"github.com/flanksource/commons/har"
	telltale "github.com/flanksource/telltale/sdk"
	"github.com/flanksource/telltale/sdk/protocol"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dashboard Telltale agent", func() {
	It("captures incoming dashboard HTTP and statements from Gavel's database", func() {
		agent := telltale.Loopback()
		DeferCleanup(agent.Close)
		httpProbe := armDashboardProbe(agent, "http", `{}`)
		sqlProbe := armDashboardProbe(agent, "sql", `{"connections":["self"]}`)

		const path = "/api/projects"
		const query = `SELECT "values" FROM user_filters`
		handler := dashboardHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			connection.PublishSQL(connection.Statement{Connection: connection.OwnDatabase, SQL: query})
			w.WriteHeader(http.StatusAccepted)
			w.(http.Flusher).Flush()
		}))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://gavel.test"+path, nil))
		Expect(response.Flushed).To(BeTrue())

		var exchange har.Entry
		readDashboardProbe(httpProbe, &exchange)
		Expect(exchange.Request.URL).To(Equal("http://gavel.test" + path))
		Expect(exchange.Response.Status).To(Equal(http.StatusAccepted))

		var statement connection.Statement
		readDashboardProbe(sqlProbe, &statement)
		Expect(statement.SQL).To(Equal(query))
	})
})

func armDashboardProbe(agent *telltale.Agent, name, params string) *protocol.Conn {
	stream := agent.Connect()
	DeferCleanup(stream.Close)
	Expect(stream.SetDeadline(time.Now().Add(5 * time.Second))).To(Succeed())
	wire, _, err := protocol.Authenticate(stream, "")
	Expect(err).NotTo(HaveOccurred())
	Expect(wire.Send("arm", protocol.Arm{Session: name, Probe: name, Params: json.RawMessage(params)})).To(Succeed())
	frame, _, err := wire.Read()
	Expect(err).NotTo(HaveOccurred())
	Expect(frame.Type).To(Equal("armed"))
	var armed protocol.Armed
	Expect(json.Unmarshal(frame.Body, &armed)).To(Succeed())
	Expect(armed.Error).To(BeEmpty())
	return wire
}

func readDashboardProbe(wire *protocol.Conn, record any) {
	frame, _, err := wire.Read()
	Expect(err).NotTo(HaveOccurred())
	Expect(frame.Type).To(Equal("records"))
	var records protocol.Records
	Expect(json.Unmarshal(frame.Body, &records)).To(Succeed())
	Expect(records.Items).To(HaveLen(1))
	Expect(json.Unmarshal(records.Items[0], record)).To(Succeed())
}
