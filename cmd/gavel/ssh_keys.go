// `gavel ssh-key add/list/rm`: the logged-in user's SSH public keys on a hosted gavel,
// which identify them to its push server.
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/flanksource/gavel/sshkeys"
	"github.com/spf13/cobra"
)

func init() {
	keyCmd := &cobra.Command{Use: "ssh-key", Short: "Manage your SSH keys for pushing to a hosted gavel"}
	var name string
	add := &cobra.Command{
		Use: "add [public-key-file]", Short: "Register an SSH public key (default ~/.ssh/id_ed25519.pub)", Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519.pub")
			if len(args) == 1 {
				path = args[0]
			}
			publicKey, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read the public key: %w", err)
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			server, err := openGavelServer(ctx)
			if err != nil {
				return err
			}
			var key sshkeys.Key
			if err := server.call(ctx, "POST", "/api/ssh-keys", map[string]string{"name": name, "publicKey": string(publicKey)}, &key); err != nil {
				return fmt.Errorf("add %s: %w", path, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added %s (%s).\n", key.Fingerprint, key.ID)
			return nil
		},
	}
	add.Flags().StringVar(&name, "name", "", "A name to recognise the key by, e.g. laptop")
	keyCmd.AddCommand(add)
	keyCmd.AddCommand(&cobra.Command{
		Use: "list", Short: "List your SSH keys", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			server, err := openGavelServer(ctx)
			if err != nil {
				return err
			}
			var keys []sshkeys.Key
			if err := server.call(ctx, "GET", "/api/ssh-keys", nil, &keys); err != nil {
				return fmt.Errorf("list SSH keys: %w", err)
			}
			table := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(table, "ID\tNAME\tFINGERPRINT\tLAST USED")
			for _, key := range keys {
				used := "never"
				if key.LastUsedAt != nil {
					used = key.LastUsedAt.Local().Format(time.RFC3339)
				}
				fmt.Fprintf(table, "%s\t%s\t%s\t%s\n", key.ID, key.Name, key.Fingerprint, used)
			}
			return table.Flush()
		},
	})
	keyCmd.AddCommand(&cobra.Command{
		Use: "rm <id>", Short: "Remove one of your SSH keys", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			server, err := openGavelServer(ctx)
			if err != nil {
				return err
			}
			if err := server.call(ctx, "DELETE", "/api/ssh-keys/"+url.PathEscape(args[0]), nil, nil); err != nil {
				return fmt.Errorf("remove SSH key %s: %w", args[0], err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed %s.\n", args[0])
			return nil
		},
	})
	rootCmd.AddCommand(keyCmd)
}
