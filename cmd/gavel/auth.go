// Clerk login, whoami and logout commands for Gavel hosted services. They store the
// session locally and record the verified identity in the database when one is configured.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/internal/database"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

// openAuthDatabase and authHTTPClient are package variables so tests can observe
// the database options and reach a TLS test issuer.
var (
	openAuthDatabase = database.Open
	authHTTPClient   = &http.Client{Timeout: 20 * time.Second}
)

// printAuthIdentity prints the identity and records it in the database. Only login
// migrates; other callers skip recording when the identity tables don't exist yet.
func printAuthIdentity(cmd *cobra.Command, identity api.Identity, migrate bool) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 45*time.Second)
	defer cancel()
	var options []database.Option
	if migrate {
		options = append(options, database.WithMigrations())
	}
	db, err := openAuthDatabase(ctx, options...)
	if err != nil {
		return fmt.Errorf("persist authenticated identity: %w", err)
	}
	defer db.Close()
	if !db.Disabled() {
		store := api.IdentityStore{DB: db.Gorm()}
		ready, err := store.Ready(ctx)
		if err != nil {
			return fmt.Errorf("persist authenticated identity: %w", err)
		}
		if ready {
			identity, _, err = store.Sync(ctx, identity)
			if err != nil {
				return fmt.Errorf("persist authenticated identity: %w", err)
			}
		}
	}
	output, err := clicky.Format(identity, clicky.Flags.FormatOptions)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), output)
	return nil
}

func init() {
	var issuer, clientID string
	var noBrowser bool
	login := &cobra.Command{
		Use: "login", Short: "Authenticate with Clerk for Gavel hosted services",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
			defer stop()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()
			path, err := api.CredentialPath()
			if err != nil {
				return err
			}
			client, err := api.NewAuthClient(ctx, api.AuthConfig{Issuer: issuer, ClientID: clientID}, authHTTPClient)
			if err != nil {
				return err
			}
			device, err := client.BeginLogin(ctx)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Open %s and enter code: %s\n", device.VerificationURI, device.UserCode)
			if !noBrowser {
				target := device.VerificationURIComplete
				if target == "" {
					target = device.VerificationURI
				}
				if err := browser.OpenURL(target); err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "Could not open a browser. Open the URL above manually.")
				}
			}
			session, identity, err := client.CompleteLogin(ctx, device)
			if err != nil {
				return err
			}
			if err := api.SaveSession(path, session); err != nil {
				return err
			}
			return printAuthIdentity(cmd, identity, true)
		},
	}
	login.Args = cobra.NoArgs
	login.Flags().StringVar(&issuer, "issuer", os.Getenv("GAVEL_CLERK_ISSUER"), "Clerk Frontend API issuer URL (or GAVEL_CLERK_ISSUER)")
	login.Flags().StringVar(&clientID, "client-id", os.Getenv("GAVEL_CLERK_CLIENT_ID"), "Clerk public OAuth client ID (or GAVEL_CLERK_CLIENT_ID)")
	login.Flags().BoolVar(&noBrowser, "no-browser", false, "Print the verification URL without opening a browser")
	rootCmd.AddCommand(login)
	rootCmd.AddCommand(&cobra.Command{
		Use: "whoami", Short: "Show the verified identity of the logged-in Gavel user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			path, err := api.CredentialPath()
			if err != nil {
				return err
			}
			session, err := api.LoadSession(path)
			if err != nil {
				return err
			}
			client, err := api.NewAuthClient(ctx, session.Config, authHTTPClient)
			if err != nil {
				return err
			}
			updated, identity, err := client.WhoAmI(ctx, session)
			if err != nil {
				return err
			}
			if err := api.SaveSession(path, updated); err != nil {
				return err
			}
			return printAuthIdentity(cmd, identity, false)
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use: "logout", Short: "Remove locally stored Gavel credentials", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := api.CredentialPath()
			if err != nil {
				return err
			}
			if err := api.Logout(path); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out of Gavel on this machine.")
			return nil
		},
	})
}
