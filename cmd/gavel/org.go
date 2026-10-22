// `gavel org list` and `gavel org use`: the organizations the logged-in user belongs to on
// a hosted gavel, and which one CLI requests act in (sent as X-Gavel-Org).
package main

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/flanksource/gavel/api"
	"github.com/spf13/cobra"
)

// serverOrgs lists the user's organizations on the server, which needs none chosen.
func serverOrgs(ctx context.Context, server gavelServer) ([]api.MemberOrg, error) {
	var orgs []api.MemberOrg
	if err := server.call(ctx, "GET", "/api/auth/orgs", nil, &orgs); err != nil {
		return nil, fmt.Errorf("list your organizations: %w", err)
	}
	return orgs, nil
}

func orgSlugs(orgs []api.MemberOrg) string {
	slugs := make([]string, 0, len(orgs))
	for _, org := range orgs {
		slugs = append(slugs, org.Slug)
	}
	return strings.Join(slugs, ", ")
}

// pickOnlyOrg selects the user's organization when they have exactly one, and otherwise
// says how to pick one.
func pickOnlyOrg(ctx context.Context, cmd *cobra.Command, server gavelServer) error {
	orgs, err := serverOrgs(ctx, server)
	if err != nil {
		return err
	}
	switch len(orgs) {
	case 0:
		fmt.Fprintf(cmd.ErrOrStderr(), "Logged in to %s, but you belong to no organization there yet.\n", server.session.Server)
	case 1:
		server.session.Org = orgs[0].Slug
		fmt.Fprintf(cmd.ErrOrStderr(), "Logged in to %s as a member of %s.\n", server.session.Server, orgs[0].Slug)
	default:
		fmt.Fprintf(cmd.ErrOrStderr(), "Logged in to %s. Pick an organization with gavel org use <slug>: %s\n", server.session.Server, orgSlugs(orgs))
	}
	return nil
}

func init() {
	orgCmd := &cobra.Command{Use: "org", Short: "Choose the organization CLI requests act in on a hosted gavel"}
	orgCmd.AddCommand(&cobra.Command{
		Use: "list", Short: "List your organizations on the hosted gavel", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			server, err := openGavelServer(ctx)
			if err != nil {
				return err
			}
			orgs, err := serverOrgs(ctx, server)
			if err != nil {
				return err
			}
			table := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(table, "\tSLUG\tNAME\tROLE")
			for _, org := range orgs {
				current := ""
				if org.Slug == server.session.Org {
					current = "*"
				}
				fmt.Fprintf(table, "%s\t%s\t%s\t%s\n", current, org.Slug, org.Name, org.Role)
			}
			return table.Flush()
		},
	})
	orgCmd.AddCommand(&cobra.Command{
		Use: "use <slug>", Short: "Act in an organization for later CLI requests", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			server, err := openGavelServer(ctx)
			if err != nil {
				return err
			}
			orgs, err := serverOrgs(ctx, server)
			if err != nil {
				return err
			}
			for _, org := range orgs {
				if org.Slug == args[0] {
					server.session.Org = org.Slug
					path, err := api.CredentialPath()
					if err != nil {
						return err
					}
					if err := api.SaveSession(path, server.session); err != nil {
						return err
					}
					fmt.Fprintf(cmd.OutOrStdout(), "CLI requests now act in %s.\n", org.Slug)
					return nil
				}
			}
			return fmt.Errorf("you're not a member of %q; your organizations: %s", args[0], orgSlugs(orgs))
		},
	})
	rootCmd.AddCommand(orgCmd)
}
