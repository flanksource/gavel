package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/flanksource/clicky/text"
	shared "github.com/flanksource/commons-db/cmd/query/connections"
	connectioncli "github.com/flanksource/commons-db/cmd/query/connections/cli"
	"github.com/flanksource/commons-db/connection/store"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/commons-db/query"
	catalog "github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/spf13/cobra"
)

func newConnectionsCommand() *cobra.Command {
	var command *cobra.Command
	var keeperConnection text.SensitiveString
	command = connectioncli.New(connectioncli.Options{Sources: []string{"file", "db", "server"}, Open: func(ctx context.Context, source string) (connectioncli.Backend, func() error, error) {
		close := func() error { return nil }
		if source == "server" {
			server, err := openGavelServer(ctx)
			return connectionRemote{ctx: ctx, server: server}, close, err
		}
		if source == "db" {
			db, err := database.Require(ctx, "connections")
			if err != nil {
				return nil, close, err
			}
			return catalog.Open(ctx, catalog.Options{DB: db, Scope: tenant.Local(), KeeperConnection: keeperConnection, Interactive: true})
		}
		layers, err := store.DefaultLayers("gavel")
		if err != nil {
			return nil, close, err
		}
		dir, err := getWorkingDir()
		if err != nil {
			return nil, close, err
		}
		layers[2].Dir = filepath.Join(dir, ".connections")
		files, err := store.New(layers...)
		if err != nil {
			return nil, close, err
		}
		provider, err := shared.NewStoreProvider(shared.StoreProviderOptions{Store: files, WriteLayer: store.LayerGlobal, WritableLayers: []string{store.LayerGlobal, store.LayerApp, store.LayerCWD}})
		if err != nil {
			return nil, close, err
		}
		connectionCtx := files.Attach(dbcontext.NewContext(ctx))
		service, err := shared.New(shared.Options{Virtual: provider, Context: func() dbcontext.Context { return connectionCtx }, DecodeBody: func(_ context.Context, body map[string]any) (map[string]any, error) { return body, nil }, Profiles: func(context.Context) ([]query.Profile, error) { return nil, nil }})
		return service, close, err
	}})
	command.PersistentFlags().Func("secret-keeper-connection", "SecretKeeper connection (overrides platform default)", func(value string) error { keeperConnection = text.NewSensitiveString(value); return nil })
	kek := &cobra.Command{Use: "kek", Short: "Manage the local connection encryption key"}
	kek.AddCommand(&cobra.Command{Use: "init", Short: "Initialize a missing macOS Keychain KEK after confirmation", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		keeper, _, err := catalog.Keeper(cmd.Context(), keeperConnection, true)
		if err != nil {
			return err
		}
		if err = keeper.Close(); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Connection KEK is ready")
		return nil
	}})
	command.AddCommand(kek)
	return command
}

func init() {
	rootCmd.AddCommand(newConnectionsCommand())
	fixtures.SetupConnectionResolver = func(ctx dbcontext.Context, reference string) (connection *models.Connection, err error) {
		db, err := database.Require(ctx, "connection resolution")
		if err != nil {
			return nil, err
		}
		provider, close, err := catalog.OpenStore(ctx, catalog.Options{DB: db, Scope: tenant.FromContext(ctx), Interactive: true})
		if err != nil {
			return nil, err
		}
		defer func() { err = errors.Join(err, close()) }()
		return provider.ResolveConnection(ctx, reference)
	}
}
