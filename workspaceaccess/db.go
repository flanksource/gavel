package workspaceaccess

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/internal/database"
	"github.com/gofrs/flock"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type durable struct{}

func identity() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	dir := filepath.Join(home, ".config", "gavel")
	if err = os.MkdirAll(dir, 0700); err != nil {
		return "", "", err
	}
	lock := flock.New(filepath.Join(dir, "workspace-machine.lock"))
	if err = lock.Lock(); err != nil {
		return "", "", err
	}
	defer lock.Unlock()
	path := filepath.Join(dir, "workspace-machine-id")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		data = []byte(uuid.NewString())
		err = os.WriteFile(path, data, 0600)
	}
	if err != nil {
		return "", "", err
	}
	id, err := uuid.Parse(strings.TrimSpace(string(data)))
	if err != nil {
		return "", "", fmt.Errorf("invalid workspace machine identity: %w", err)
	}
	account, err := user.Current()
	if err != nil {
		return "", "", err
	}
	if account.Uid == "" {
		return "", "", errors.New("OS user identity is empty")
	}
	return id.String(), account.Uid, nil
}
func scoped(ctx context.Context) (*gorm.DB, error) {
	machine, user, err := identity()
	if err != nil {
		return nil, err
	}
	db, err := database.Require(ctx, "workspace access")
	if err != nil {
		return nil, err
	}
	return db.WithContext(ctx).Model(&grant{}).Where("machine_id = ? AND user_id = ?", machine, user), nil
}
func (durable) List(ctx context.Context) ([]Root, error) {
	db, err := scoped(ctx)
	if err != nil {
		return nil, err
	}
	var grants []grant
	if err = db.Order("path").Find(&grants).Error; err != nil {
		return nil, err
	}
	roots := make([]Root, 0, len(grants))
	for _, g := range grants {
		roots = append(roots, Root{ID: g.ID, Path: g.Path, Source: "saved"})
	}
	return roots, nil
}
func (durable) Grant(ctx context.Context, path string) (Root, error) {
	machine, user, err := identity()
	if err != nil {
		return Root{}, err
	}
	db, err := database.Require(ctx, "workspace access")
	if err != nil {
		return Root{}, err
	}
	g := grant{ID: uuid.NewString(), MachineID: machine, UserID: user, Path: path}
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&g).Error; err != nil {
			return err
		}
		return tx.Exec("SELECT pg_notify('gavel_workspace_access_change', ?)", machine).Error
	})
	return Root{ID: g.ID, Path: path, Source: "saved", Available: true}, err
}
func (durable) Revoke(ctx context.Context, id string) error {
	db, err := scoped(ctx)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		r := tx.Where("id = ?", id).Delete(&grant{})
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected == 0 {
			return errors.New("saved grant does not exist")
		}
		return tx.Exec("SELECT pg_notify('gavel_workspace_access_change', '')").Error
	})
}
