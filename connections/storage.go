package connections

import (
	"context"
	"errors"
	"fmt"

	"github.com/flanksource/clicky/text"
	shared "github.com/flanksource/commons-db/cmd/query/connections"
	"github.com/flanksource/commons-db/connection/catalog"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/commons-db/secret/keychain"
)

func Status(ctx context.Context, connection text.SensitiveString) shared.KEKStatus {
	keeper, _, err := Keeper(ctx, connection, false)
	if errors.Is(err, keychain.ErrMissing) {
		return shared.KEKStatus{State: "missing", Detail: "Initialize the key on this host with gavel connections kek init."}
	}
	if err != nil {
		return shared.KEKStatus{State: "unavailable", Detail: "The configured encryption key could not be opened. Check Keychain access or SecretKeeper configuration."}
	}
	if err := keeper.Close(); err != nil {
		return shared.KEKStatus{State: "unavailable", Detail: "The encryption key provider could not be closed cleanly."}
	}
	return shared.KEKStatus{State: "ready", Detail: "The connection encryption key is available on this host."}
}

type provider struct{ *catalog.Store }

func (p provider) StorageLayers() []shared.StorageLayer {
	return []shared.StorageLayer{{Name: "db", Label: "Encrypted database", Writable: true}}
}
func (p provider) DefaultLayer() string { return "db" }
func (p provider) SaveConnection(ctx context.Context, c *models.Connection) (*models.Connection, error) {
	if c.Layer != "" && c.Layer != "db" {
		return nil, fmt.Errorf("connection storage layer %q is unavailable", c.Layer)
	}
	return p.Store.SaveConnection(ctx, c)
}
