package dbconn

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
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

var _ = Describe("fixture connections", func() {
	var step *fixtures.StepContext
	BeforeEach(func() {
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		step = &fixtures.StepContext{WorkDir: GinkgoT().TempDir()}
		writeConnection(step.WorkDir, "lab", "file:/data/lab.db")
		writeConnection(step.WorkDir, "archive", "file:/data/archive.db")
	})

	It("resolves named connections ahead of the catalog: inline by name, string refs through the catalog", func() {
		step.Test.Connections = &fixtures.ConnectionSet{Named: map[string]any{
			"lab":     map[string]any{"type": "postgres", "url": "postgres://u:p@db.example:5432/lab"},
			"history": "connection://archive",
		}}
		conns, err := StepConnections(context.Background(), step, Options{})
		Expect(err).NotTo(HaveOccurred())
		urls := map[string]string{}
		for _, ref := range []string{"lab", "connection://lab", "history"} {
			resolved, err := conns.Resolve(ref)
			Expect(err).NotTo(HaveOccurred())
			urls[ref] = resolved.Connection.Name + " " + resolved.Connection.URL
		}
		Expect(urls).To(Equal(map[string]string{
			"lab":              "lab postgres://u:p@db.example:5432/lab",
			"connection://lab": "lab postgres://u:p@db.example:5432/lab",
			"history":          "archive file:/data/archive.db",
		}))
		named, err := conns.NamedOfTypes([]string{models.ConnectionTypeSQLite})
		Expect(err).NotTo(HaveOccurred())
		Expect(named).To(HaveLen(1))
		Expect([]string{named[0].Name, named[0].Ref, named[0].Connection.Type}).To(Equal([]string{"history", "connection://history", "sqlite"}))
	})

	It("keys list entries by their type, the one default a step without a connection uses", func() {
		step.Test.Connections = &fixtures.ConnectionSet{Defaults: []string{"lab", "warehouse"}}
		step.Opts.Connections = func(_ dbcontext.Context, reference string) (*models.Connection, error) {
			if reference != "warehouse" {
				return nil, nil
			}
			return &models.Connection{Name: "warehouse", Type: models.ConnectionTypePostgres, URL: "postgres://db.example/w"}, nil
		}
		conns, err := StepConnections(context.Background(), step, Options{})
		Expect(err).NotTo(HaveOccurred())
		sqlite, ok := conns.DefaultFor(models.ConnectionTypeSQLite)
		Expect(ok).To(BeTrue())
		Expect(sqlite.Ref).To(Equal("connection://lab"))
		_, err = conns.SingleDefault()
		Expect(err).To(MatchError(And(ContainSubstring(`"lab" (sqlite)`), ContainSubstring(`"warehouse" (postgres)`))))
	})

	DescribeTable("fails loudly on defaults that cannot each stand for one type",
		func(defaults []string, want []string) {
			step.Test.Connections = &fixtures.ConnectionSet{Defaults: defaults}
			_, err := StepConnections(context.Background(), step, Options{})
			for _, part := range want {
				Expect(err).To(MatchError(ContainSubstring(part)))
			}
		},
		Entry("two of one type", []string{"lab", "archive"}, []string{`"lab"`, `"archive"`, "both sqlite"}),
		Entry("one that resolves nowhere", []string{"lab", "missing"}, []string{`connections[1] "missing"`, "connection not found"}),
	)

	It("has no single default when none is declared", func() {
		conns, err := StepConnections(context.Background(), step, Options{})
		Expect(err).NotTo(HaveOccurred())
		_, err = conns.SingleDefault()
		Expect(err).To(MatchError(ContainSubstring("no default connection")))
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
