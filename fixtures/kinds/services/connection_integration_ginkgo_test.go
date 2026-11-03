package services_test

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons-db/dbtest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("borrowed database service", func() {
	It("exports a saved connection and leaves its server running after cleanup", func() {
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		db, cleanup, err := dbtest.Open(dbtest.Options{Name: "borrowed-service"})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(cleanup)
		step := newStep("borrow", `{"services":{"database":{"deps":"postgres","connection":{"ref":"connection://saved"},"export":{"FIXTURE_DSN":"{{.outputs.url}}"}}}}`)
		dir := filepath.Join(step.WorkDir, ".connections")
		Expect(os.Mkdir(dir, 0700)).To(Succeed())
		body := "apiVersion: mission-control.flanksource.com/v1\nkind: Connection\nmetadata: {name: saved}\nspec:\n  postgres:\n    url: {value: '" + db.DSN() + "'}\n"
		Expect(os.WriteFile(filepath.Join(dir, "saved.yaml"), []byte(body), 0600)).To(Succeed())
		outcome, err := execute(step)
		Expect(err).NotTo(HaveOccurred())
		Expect(entryOutputs(outcome, "database")).To(HaveKeyWithValue("source", "existing"))
		Expect(envValue(step, "FIXTURE_DSN")).To(Equal(db.DSN()))
		Expect(outcome.Cleanup(context.Background())).To(Succeed())
		client := databaseConnection(db.DSN())
		DeferCleanup(client.Close)
		Expect(client.Ping()).To(Succeed())
		endpoint, err := url.Parse(db.DSN())
		Expect(err).NotTo(HaveOccurred())
		endpoint.User = url.UserPassword("fixture_invalid_user", "fixture_invalid_password")
		invalidCredentials := endpoint.String()
		endpoint.Host = "127.0.0.1:0"
		for _, invalid := range []string{invalidCredentials, endpoint.String()} {
			Expect(os.WriteFile(filepath.Join(dir, "saved.yaml"), []byte(strings.Replace(body, db.DSN(), invalid, 1)), 0600)).To(Succeed())
			failed := newStep("bad", string(step.Spec.Config))
			failed.WorkDir = step.WorkDir
			_, err := execute(failed)
			Expect(err).To(MatchError(ContainSubstring("existing connection probe")))
		}
	})
})
