// Tests the database-backed Authorizer against a migrated database and the fake Clerk:
// members push to their org's repositories, everyone else is refused.
package serve

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"path/filepath"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/api/clerktest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/sshkeys"
	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gossh "golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

var _ = Describe("DBAuthorizer", func() {
	var (
		db         *gorm.DB
		authorizer DBAuthorizer
		ada        uuid.UUID
		dataDir    string
		ctx        = context.Background()
	)

	orgID := func(slug string) uuid.UUID {
		var org struct{ ID uuid.UUID }
		Expect(db.Raw(`SELECT id FROM orgs WHERE slug = ?`, slug).Scan(&org).Error).To(Succeed())
		return org.ID
	}

	BeforeEach(func() {
		handle := commonsdbtest.ForGinkgo(commonsdbtest.Options{Name: "gavel_push_authorizer"})
		GinkgoT().Setenv(database.EnvDSN, handle.DSN())
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		db = opened.Gorm()

		clerk := clerktest.New(GinkgoT(), "cli", clerktest.User{ID: "user_ada", Name: "Ada"})
		clerk.Memberships = []clerktest.Membership{
			{OrgID: "org_a", OrgSlug: "alpha", OrgName: "Alpha", Role: "org:member"},
		}
		members := api.NewMembers(db, api.ClerkOrganizations{SecretKey: clerktest.SecretKey, HTTPClient: clerk.BackendClient()}, clerk.URL, 5*time.Minute)
		_, err = members.Orgs(ctx, "user_ada")
		Expect(err).NotTo(HaveOccurred())
		var user struct{ ID uuid.UUID }
		Expect(db.Raw(`SELECT id FROM users WHERE clerk_user_id = 'user_ada'`).Scan(&user).Error).To(Succeed())
		ada = user.ID
		Expect(db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, slug, name) VALUES (?, ?, 'org_g', 'gamma', 'Gamma')`, uuid.New(), clerk.URL).Error).To(Succeed())

		dataDir = GinkgoT().TempDir()
		authorizer = DBAuthorizer{DB: db, Members: members, Issuer: clerk.URL, DataDir: dataDir}
	})

	It("maps a registered key to its user", func() {
		public, _, err := ed25519.GenerateKey(rand.Reader)
		Expect(err).NotTo(HaveOccurred())
		key, err := gossh.NewPublicKey(public)
		Expect(err).NotTo(HaveOccurred())
		_, known := authorizer.UserForKey(ctx, key)
		Expect(known).To(BeFalse())

		_, err = sshkeys.Store{DB: db}.Add(ctx, ada, "laptop", string(gossh.MarshalAuthorizedKey(key)))
		Expect(err).NotTo(HaveOccurred())
		user, known := authorizer.UserForKey(ctx, ssh.PublicKey(key))
		Expect(known).To(BeTrue())
		Expect(user).To(Equal(ada))
	})

	It("lets a member push to their org's repository, the same one each time", func() {
		target, err := authorizer.AuthorizePush(ctx, ada, "alpha/app")
		Expect(err).NotTo(HaveOccurred())
		Expect(target.OrgID).To(Equal(orgID("alpha")))
		Expect(target.BareDir).To(Equal(filepath.Join(dataDir, "repos", target.OrgID.String(), target.RepoID.String()+".git")))
		var repo struct{ Name, Source string }
		Expect(db.Raw(`SELECT name, source FROM repos WHERE id = ?`, target.RepoID).Scan(&repo).Error).To(Succeed())
		Expect(repo.Name).To(Equal("app"))
		Expect(repo.Source).To(Equal("push"))

		again, err := authorizer.AuthorizePush(ctx, ada, "alpha/app")
		Expect(err).NotTo(HaveOccurred())
		Expect(again).To(Equal(target))
	})

	It("refuses a non-member, an unknown org, and a path without an org", func() {
		_, err := authorizer.AuthorizePush(ctx, ada, "gamma/app")
		Expect(err).To(MatchError(api.ErrNotMember))
		_, err = authorizer.AuthorizePush(ctx, ada, "nowhere/app")
		Expect(err).To(MatchError(api.ErrNotMember))
		_, err = authorizer.AuthorizePush(ctx, ada, "app")
		Expect(err).To(MatchError(ContainSubstring("<org>/<repo>")))
		_, err = authorizer.AuthorizePush(ctx, uuid.New(), "alpha/app")
		Expect(err).To(MatchError(ContainSubstring("unknown user")))

		var count int64
		Expect(db.Raw(`SELECT count(*) FROM repos`).Scan(&count).Error).To(Succeed())
		Expect(count).To(BeZero(), "refused pushes create no repository")
		Expect(filepath.Join(dataDir, "repos")).NotTo(BeADirectory())
	})
})
