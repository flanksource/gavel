// Tests the connection a local gavel reaches an Azure DevOps repository through: the git
// connection owning it, else AZURE_DEVOPS_EXT_PAT, else the az CLI login.
package azuredevops

import (
	"context"
	"errors"
	"time"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("A local Azure DevOps repository's connection", func() {
	var repo git.RepoURL
	BeforeEach(func() {
		var err error
		repo, err = git.ParseRepoURL("https://dev.azure.com/contoso/web/_git/site", "")
		Expect(err).NotTo(HaveOccurred())
	})
	owning := func(connection *models.Connection, err error, pat string) {
		previous := FindConnection
		FindConnection = func(context.Context, git.RepoURL) (*models.Connection, error) { return connection, err }
		DeferCleanup(func() { FindConnection = previous })
		GinkgoT().Setenv("AZURE_DEVOPS_EXT_PAT", pat)
	}

	It("is the git connection that owns the repository", func(ctx SpecContext) {
		owner := models.Connection{Name: "contoso", Type: models.ConnectionTypeAzureDevops, URL: "https://dev.azure.com/contoso", Password: "connection-pat"}
		owning(&owner, nil, "env-pat")
		connection, options, err := LocalConnection(ctx, repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(connection).To(Equal(owner))
		Expect(options.TokenSource).To(BeNil())
	})

	It("carries AZURE_DEVOPS_EXT_PAT for the repository's organization when no connection owns it", func(ctx SpecContext) {
		owning(nil, nil, "env-pat")
		connection, options, err := LocalConnection(ctx, repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(connection).To(Equal(models.Connection{Name: "AZURE_DEVOPS_EXT_PAT", Type: models.ConnectionTypeAzureDevops,
			URL: "https://dev.azure.com/contoso", Password: "env-pat"}))
		Expect(options.TokenSource).To(BeNil())
	})

	It("authenticates with the az CLI login without a PAT", func(ctx SpecContext) {
		owning(nil, nil, "")
		previous := cliLogin
		fetched := 0
		cliLogin = &authenticator{now: time.Now, fetch: func(context.Context) (accessToken, error) {
			fetched++
			return accessToken{Token: "entra", Expires: time.Now().Add(time.Hour)}, nil
		}}
		DeferCleanup(func() { cliLogin = previous })

		connection, options, err := LocalConnection(ctx, repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(connection).To(Equal(models.Connection{Name: "az CLI", Type: models.ConnectionTypeAzureDevops, URL: "https://dev.azure.com/contoso"}))
		Expect(options.TokenSource).NotTo(BeNil())
		for range 2 {
			token, err := options.TokenSource(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(token).To(Equal("entra"))
		}
		Expect(fetched).To(Equal(1), "the login's token is reused until it nears expiry")
		_, err = git.FromConnection(dbcontext.New(), connection, options)
		Expect(err).NotTo(HaveOccurred(), "a client needs no PAT with the login")
	})

	It("refuses a connection that owns the repository but holds no PAT", func(ctx SpecContext) {
		owning(&models.Connection{Name: "contoso", Type: models.ConnectionTypeAzureDevops}, nil, "env-pat")
		_, _, err := LocalConnection(ctx, repo)
		Expect(err).To(MatchError(`git connection "contoso" owns Azure DevOps repository https://dev.azure.com/contoso/web/_git/site but holds no personal access token`))
	})

	It("fails with the lookup's reason", func(ctx SpecContext) {
		owning(nil, errors.New("the catalog's key is unavailable"), "env-pat")
		_, _, err := LocalConnection(ctx, repo)
		Expect(err).To(MatchError("find the git connection for Azure DevOps repository https://dev.azure.com/contoso/web/_git/site: the catalog's key is unavailable"))
	})
})
