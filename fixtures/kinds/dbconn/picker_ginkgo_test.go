package dbconn

import (
	"github.com/flanksource/clicky/schemaform"
	"github.com/flanksource/commons-db/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("PickerOption", func() {
	tunnel := &models.Connection{Name: "tunnel", Type: "postgres", URL: "portforward://pg.public.example.com:5432/app"}

	DescribeTable("describes where a connection points and how it is reached, never its credentials",
		func(connection *models.Connection, description string, badges types.GomegaMatcher) {
			option := PickerOption(connection, []*models.Connection{tunnel, connection})
			Expect(option.Description).To(Equal(description))
			Expect(option.Badges).To(badges)
			Expect(option.Description).NotTo(ContainSubstring("s3cret"))
		},
		Entry("a URL with userinfo", &models.Connection{Name: "main", Type: "postgres", URL: "postgres://u:s3cret@db.example.com:5432/sales?sslmode=disable"},
			"db.example.com:5432 · sales", BeEmpty()),
		Entry("an ADO keyword string", &models.Connection{Name: "mssql", Type: "sql_server", URL: "server=sql.example.com;port=1433;user id=u;password=s3cret;database=erp"},
			"sql.example.com:1433 · erp", BeEmpty()),
		Entry("an in-cluster service", &models.Connection{Name: "svc", Type: "postgres", URL: "postgres://u@postgres.data:5432/app"},
			"postgres.data:5432 · app", Equal([]schemaform.Badge{{Label: "svc", Icon: svcIcon, Tone: "teal"}})),
		Entry("a port-forward", tunnel,
			"pg.public.example.com:5432 · app", Equal([]schemaform.Badge{{Label: "proxy", Icon: proxyIcon, Tone: "violet"}})),
		Entry("a delegation, followed to its target", &models.Connection{Name: "alias", Type: "postgres", Properties: map[string]string{"connection": "connection://tunnel"}},
			"pg.public.example.com:5432 · app", Equal([]schemaform.Badge{{Label: "via tunnel", Icon: viaIcon, Tone: "indigo"}, {Label: "proxy", Icon: proxyIcon, Tone: "violet"}})),
		Entry("a sqlite file", &models.Connection{Name: "lab", Type: "sqlite", URL: "/data/lab.db"}, "/data/lab.db", BeEmpty()),
	)

	It("labels, icons and groups the option by its type", func() {
		option := PickerOption(&models.Connection{Name: "erp", Namespace: "ops", Type: "sql_server"}, nil)
		Expect([]string{option.Value, option.Label, option.Icon, option.Group}).To(Equal([]string{"connection://ops/erp", "ops/erp (sql_server)", "sqlserver", "Sql Server"}))
	})
})
