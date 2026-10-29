package jira

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/externalissues/syncdoc"
)

func ResolveSyncFile(options SyncOptions, site string) (SyncOptions, string, error) {
	if options.From == "" {
		return options, site, nil
	}
	if len(options.Keys) > 0 || options.Dir != "" {
		return options, site, fmt.Errorf("--from cannot be combined with issue IDs or --sync-dir")
	}
	path, err := filepath.Abs(options.From)
	if err != nil {
		return options, site, err
	}
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return options, site, err
	}
	defer root.Close()
	doc, err := syncdoc.Read(root, filepath.Base(path))
	if err != nil {
		return options, site, err
	}
	source := doc.Metadata.Source
	if doc.Metadata.Version != 1 || source.Provider != "jira" || source.Site == "" || (source.ID == "" && source.Key == "") {
		return options, site, fmt.Errorf("--from requires version 1 Jira source metadata")
	}
	stored, err := normalizeSite(source.Site)
	if err != nil {
		return options, site, err
	}
	if site != "" {
		supplied, err := normalizeSite(site)
		if err != nil {
			return options, site, err
		}
		if supplied != stored {
			return options, site, fmt.Errorf("--site does not match saved Jira source")
		}
	}
	key := source.ID
	if key == "" {
		key = source.Key
	}
	options.Keys = []string{key}
	options.Output = path
	options.CloudID = source.CloudID
	options.Attachments = doc.Metadata.Sync.Attachments
	options.Extract = doc.Metadata.Sync.Extract
	return options, stored, nil
}
