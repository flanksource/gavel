package dbconn

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/clicky/schemaform"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/form"
)

func TestDBConn(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dbconn Suite")
}

func writeConnection(dir, name, url string) {
	Expect(os.MkdirAll(filepath.Join(dir, ".connections"), 0o700)).To(Succeed())
	body := "apiVersion: mission-control.flanksource.com/v1\nkind: Connection\nmetadata:\n  name: " + name +
		"\nspec:\n  type: sqlite\n  url:\n    value: " + url + "\n"
	Expect(os.WriteFile(filepath.Join(dir, ".connections", name+".yaml"), []byte(body), 0o600)).To(Succeed())
}

var _ = Describe("Resolve", func() {
	var step *fixtures.StepContext
	BeforeEach(func() {
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		step = &fixtures.StepContext{WorkDir: GinkgoT().TempDir()}
	})

	It("resolves a name from the fixture directory's .connections and answers its ref", func() {
		writeConnection(step.WorkDir, "lab", "file:/data/lab.db")
		resolved, err := Resolve(context.Background(), step, "connection://lab", Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Connection.Type).To(Equal(models.ConnectionTypeSQLite))
		Expect(resolved.Ref).To(Equal("connection://lab"))
		answered, err := resolved.Context.ConnectionResolver()(resolved.Context, "lab")
		Expect(err).NotTo(HaveOccurred())
		Expect(answered.URL).To(Equal("file:/data/lab.db"))
	})

	It("asks the host resolver before the gavel layers", func() {
		writeConnection(step.WorkDir, "lab", "file:/data/from-store.db")
		step.Opts.Connections = func(_ dbcontext.Context, reference string) (*models.Connection, error) {
			if reference != "lab" {
				return nil, nil
			}
			return &models.Connection{Name: "lab", Type: models.ConnectionTypeSQLite, URL: "file:/data/from-host.db"}, nil
		}
		resolved, err := Resolve(context.Background(), step, "lab", Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Connection.URL).To(Equal("file:/data/from-host.db"))
	})

	It("reads an inline connection spec under its name", func() {
		resolved, err := Resolve(context.Background(), step, map[string]any{
			"name": "scratch", "type": "sqlite", "url": "file:/data/scratch.db",
		}, Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{resolved.Connection.Name, resolved.Connection.URL, resolved.Ref}).
			To(Equal([]string{"scratch", "file:/data/scratch.db", "connection://scratch"}))
	})

	It("refuses an inline connection with an unknown key", func() {
		_, err := Resolve(context.Background(), step, map[string]any{"type": "sqlite", "uri": "x"}, Options{})
		Expect(err).To(MatchError(ContainSubstring(`unknown field "uri"`)))
	})

	It("fails loudly, naming the reference, when nothing defines it", func() {
		_, err := Resolve(context.Background(), step, "connection://missing", Options{})
		Expect(err).To(MatchError(ErrNotFound))
		Expect(err).To(MatchError(ContainSubstring(`"connection://missing"`)))
	})
})

var _ = Describe("connection pickers", func() {
	var workDir string
	BeforeEach(func() {
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		workDir = GinkgoT().TempDir()
		writeConnection(workDir, "lab", "file:/data/lab.db")
		writeConnection(workDir, "archive", "file:/data/archive.db")
		body := "apiVersion: mission-control.flanksource.com/v1\nkind: Connection\nmetadata:\n  name: warehouse\n" +
			"spec:\n  type: sql_server\n  url:\n    value: sqlserver://db.example\n  password:\n    value: s3cret\n"
		Expect(os.WriteFile(filepath.Join(workDir, ".connections", "warehouse.yaml"), []byte(body), 0o600)).To(Succeed())
	})

	It("lists the catalog scoped to the picker's types and search text, without secrets", func() {
		service, err := Service(context.Background(), workDir, Options{})
		Expect(err).NotTo(HaveOccurred())

		sqlite, err := Lookup(service, LookupQuery{Types: []string{"sqlite"}, Query: "LA"})
		Expect(err).NotTo(HaveOccurred())
		Expect(names(sqlite)).To(Equal([]string{"lab"}))

		all, err := Lookup(service, LookupQuery{})
		Expect(err).NotTo(HaveOccurred())
		Expect(names(all)).To(Equal([]string{"archive", "lab", "warehouse"}))
		Expect(all[2].Password).NotTo(ContainSubstring("s3cret"))
	})

	It("offers a terminal form the same connections, committing their connection:// reference", func() {
		lookup := TerminalLookup(workDir, Options{})
		result, err := lookup(context.Background(), schemaform.LookupRequest{Spec: schemaform.LookupSpec{
			Filter: form.ConnectionLookupFilter, URL: form.ConnectionLookupURL + "?types=sql_server",
		}})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Options).To(Equal([]schemaform.Option{{Value: "connection://warehouse", Label: "warehouse (sql_server)"}}))
	})

	It("fails a terminal picker loudly when no connection of its types exists", func() {
		_, err := TerminalLookup(workDir, Options{})(context.Background(), schemaform.LookupRequest{Spec: schemaform.LookupSpec{
			Filter: form.ConnectionLookupFilter, URL: form.ConnectionLookupURL + "?types=postgres",
		}})
		Expect(err).To(MatchError(ContainSubstring("no postgres connection")))
	})
})

func names(connections []*models.Connection) []string {
	out := make([]string, len(connections))
	for i, c := range connections {
		out[i] = c.Name
	}
	return out
}

var _ = Describe("Endpoint", func() {
	It("preserves keyword SQL Server credentials and parameters", func() {
		driver, endpoint, err := Endpoint(&models.Connection{Type: "sql_server", URL: "server=db.example;port=31433;user id=tenant;password=p@ss;database=example;encrypt=disable"})
		Expect(err).NotTo(HaveOccurred())
		Expect(driver).To(Equal("sqlserver"))
		Expect(endpoint.Host).To(Equal("db.example:31433"))
		Expect(endpoint.User.Username()).To(Equal("tenant"))
		password, _ := endpoint.User.Password()
		Expect(password).To(Equal("p@ss"))
		Expect(endpoint.Query().Get("database")).To(Equal("example"))
		Expect(endpoint.Query().Get("encrypt")).To(Equal("disable"))
	})
})
