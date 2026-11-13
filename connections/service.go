// Package connections wires Gavel's tenant scope to the shared encrypted catalog.
package connections

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/text"
	shared "github.com/flanksource/commons-db/cmd/query/connections"
	"github.com/flanksource/commons-db/connection/catalog"
	"github.com/flanksource/commons-db/connection/store"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/query"
	"github.com/flanksource/commons-db/secret"
	"github.com/flanksource/commons-db/secret/keychain"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"gocloud.dev/secrets"
	"golang.org/x/term"
	"gorm.io/gorm"
)

type Options struct {
	DB               *gorm.DB
	Scope            tenant.Scope
	KeeperConnection text.SensitiveString
	Interactive      bool
}

func Keeper(ctx context.Context, connection text.SensitiveString, interactive bool) (*secrets.Keeper, string, error) {
	if connection == "" {
		connection = text.SensitiveStringFromEnv("GAVEL_SECRET_KEEPER_CONNECTION")
	}
	if connection != "" {
		files, err := store.Open("gavel")
		if err != nil {
			return nil, "", err
		}
		keeper, err := secret.KeeperFromConnection(files.Attach(dbcontext.NewContext(ctx)), connection.Value())
		if err != nil {
			return nil, "", fmt.Errorf("cannot open configured SecretKeeper: check the connection type, credentials and key configuration")
		}
		id := sha256.Sum256([]byte(connection.Value()))
		return keeper, fmt.Sprintf("configured-%x", id[:16]), err
	}
	if runtime.GOOS == "darwin" {
		opts := keychain.Options{Service: "com.flanksource.gavel", Account: "connections-kek:v1"}
		if interactive && term.IsTerminal(int(os.Stdin.Fd())) {
			opts.Confirm = func(ctx context.Context) (bool, error) {
				approved, ok := clicky.PromptSelectCtx(ctx, []bool{false, true}, clicky.PromptSelectOptions[bool]{Title: "Initialize a KEK in macOS Keychain? A new key cannot decrypt existing records encrypted with a lost key."})
				return approved && ok, nil
			}
		}
		return keychain.Open(ctx, opts)
	}
	var key secret.Sensitive
	var err error
	key, err = base64.StdEncoding.DecodeString(text.SensitiveStringFromEnv("GAVEL_SECRET_KEY").Value())
	if err != nil || len(key) < 32 {
		return nil, "", fmt.Errorf("encrypted connections require a keeper connection or GAVEL_SECRET_KEY (at least 32 base64-encoded bytes)")
	}
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte("gavel-connections-v1"))
	derived := secret.Sensitive(mac.Sum(nil))
	keeper, err := secret.KeeperFromConnection(dbcontext.NewContext(ctx), "local_kms://"+base64.URLEncoding.EncodeToString([]byte(derived)))
	id := sha256.Sum256([]byte(derived))
	return keeper, fmt.Sprintf("local-%x", id[:16]), err
}

func OpenStore(ctx context.Context, options Options) (*catalog.Store, func() error, error) {
	if options.DB == nil || options.Scope.OrgID == uuid.Nil {
		return nil, nil, fmt.Errorf("connection catalog requires database and organization")
	}
	keeper, keyID, err := Keeper(ctx, options.KeeperConnection, options.Interactive)
	if err != nil {
		return nil, nil, err
	}
	provider, err := catalog.New(catalog.Options{DB: options.DB, OrgID: options.Scope.OrgID, Keeper: keeper, KeyID: keyID})
	if err != nil {
		return nil, nil, errors.Join(err, keeper.Close())
	}
	return provider, keeper.Close, nil
}

func Open(ctx context.Context, options Options) (*shared.Service, func() error, error) {
	store, close, err := OpenStore(ctx, options)
	if err != nil {
		return nil, nil, err
	}
	connectionCtx := dbcontext.NewContext(ctx).WithConnectionResolver(store.ResolveConnection)
	service, err := shared.New(shared.Options{Virtual: provider{store}, KEKStatus: func(context.Context) shared.KEKStatus {
		return shared.KEKStatus{State: "ready", Detail: "The connection encryption key is available on this host."}
	}, Context: func() dbcontext.Context { return connectionCtx }, DecodeBody: func(_ context.Context, body map[string]any) (map[string]any, error) { return body, nil }, Profiles: func(context.Context) ([]query.Profile, error) { return nil, nil }})
	if err != nil {
		return nil, nil, errors.Join(err, close())
	}
	return service, close, nil
}
