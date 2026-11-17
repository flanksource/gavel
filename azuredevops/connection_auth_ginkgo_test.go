// Tests that the git connection owning an Azure DevOps repository supplies its PAT, ahead of
// AZURE_DEVOPS_EXT_PAT and the az CLI.
package azuredevops

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Azure authentication through git connections", func() {
	repo := Repository{Organization: "contoso", Project: "web", Name: "site"}
	var lookups []git.RepoURL

	finding := func(connection *models.Connection, err error) {
		lookups = nil
		previous := FindConnection
		FindConnection = func(_ context.Context, target git.RepoURL) (*models.Connection, error) {
			lookups = append(lookups, target)
			return connection, err
		}
		DeferCleanup(func() { FindConnection = previous })
		GinkgoT().Setenv("AZURE_DEVOPS_EXT_PAT", "env-pat")
	}
	basic := func(pat string) string { return "Basic " + base64.StdEncoding.EncodeToString([]byte(":"+pat)) }

	It("uses the PAT of the connection that owns the repository", func() {
		finding(&models.Connection{Name: "contoso", Type: models.ConnectionTypeAzureDevops, Password: "connection-pat"}, nil)
		header, err := New(repo).auth.authorization(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(header).To(Equal(basic("connection-pat")))
		Expect(lookups).To(HaveLen(1))
		Expect(lookups[0].Owner).To(Equal("contoso"))
		Expect(lookups[0].Project).To(Equal("web"))
		Expect(lookups[0].Name).To(Equal("site"))
	})
	It("falls back to AZURE_DEVOPS_EXT_PAT when no connection owns the repository", func() {
		finding(nil, nil)
		header, err := New(repo).auth.authorization(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(header).To(Equal(basic("env-pat")))
	})
	It("fails with the lookup's reason", func() {
		finding(nil, errors.New("the catalog's key is unavailable"))
		_, err := New(repo).auth.authorization(context.Background())
		Expect(err).To(MatchError("find the git connection for Azure DevOps repository https://dev.azure.com/contoso/web/_git/site: the catalog's key is unavailable"))
	})
	It("refuses a connection that holds no PAT", func() {
		finding(&models.Connection{Name: "contoso", Type: models.ConnectionTypeAzureDevops}, nil)
		_, err := New(repo).auth.authorization(context.Background())
		Expect(err).To(MatchError(`git connection "contoso" owns Azure DevOps repository https://dev.azure.com/contoso/web/_git/site but holds no personal access token`))
	})
})
