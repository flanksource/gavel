// Stores each organization's dashboard search settings in org_settings, the hosted
// counterpart of the local settings file.
package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoadOrgSearchConfig returns org's saved search settings, or the default when it has none.
func LoadOrgSearchConfig(ctx context.Context, db *gorm.DB, org uuid.UUID) (SearchConfig, error) {
	var row struct{ SearchConfig []byte }
	err := db.WithContext(ctx).Raw(`SELECT search_config FROM org_settings WHERE org_id = ?`, org).Scan(&row).Error
	if err != nil {
		return SearchConfig{}, fmt.Errorf("load the organization's search settings: %w", err)
	}
	var cfg SearchConfig
	if len(row.SearchConfig) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(row.SearchConfig, &cfg); err != nil {
		return SearchConfig{}, fmt.Errorf("decode the organization's search settings: %w", err)
	}
	return cfg, nil
}

// SaveOrgSearchConfig replaces org's search settings.
func SaveOrgSearchConfig(ctx context.Context, db *gorm.DB, org uuid.UUID, cfg SearchConfig) error {
	if db == nil {
		return errors.New("saving an organization's search settings requires the database")
	}
	encoded, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := db.WithContext(ctx).Exec(`
		INSERT INTO org_settings (org_id, search_config, updated_at) VALUES (?, ?, now())
		ON CONFLICT (org_id) DO UPDATE SET search_config = EXCLUDED.search_config, updated_at = now()`,
		org, string(encoded)).Error; err != nil {
		return fmt.Errorf("save the organization's search settings: %w", err)
	}
	return nil
}
