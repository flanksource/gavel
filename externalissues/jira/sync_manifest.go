package jira

import (
	"fmt"
	"os"
	"time"

	"github.com/flanksource/gavel/externalissues/syncdoc"
)

func prepareSyncMetadata(doc *syncdoc.Document, issue Issue, options SyncOptions) error {
	previous := doc.Metadata.Source
	if doc.Metadata.Version != 0 && (previous.Provider != "jira" || previous.Site != issue.Site || (previous.ID != "" && previous.ID != issue.ID) || (previous.ID == "" && previous.Key != issue.Key)) {
		return fmt.Errorf("saved source identity does not match issue %s", issue.Key)
	}
	doc.Metadata.Version = 1
	doc.Metadata.Source = syncdoc.Source{Provider: "jira", Site: issue.Site, ID: issue.ID, Key: issue.Key, URL: issue.URL, Title: issue.Title, ProjectID: issue.Project.ID, ProjectKey: issue.Project.Key, UpdatedAt: issue.Updated}
	doc.Metadata.Source.Extra = previous.Extra
	extra := doc.Metadata.Sync.Extra
	doc.Metadata.Sync = syncdoc.Settings{SyncedAt: time.Now().UTC(), Attachments: options.Attachments || options.Extract, Extract: options.Extract}
	doc.Metadata.Sync.Extra = extra
	if doc.Metadata.Attachments == nil {
		doc.Metadata.Attachments = map[string]*syncdoc.Attachment{}
	}
	for _, record := range doc.Metadata.Attachments {
		record.Present = false
	}
	for _, a := range issue.Attachments {
		record := doc.Metadata.Attachments[a.ID]
		if record == nil {
			record = &syncdoc.Attachment{}
			doc.Metadata.Attachments[a.ID] = record
		}
		record.Name = a.Name
		record.URL = a.URL
		record.MIMEType = a.MIMEType
		record.Size = a.Size
		record.CreatedAt = a.Created
		record.Present = true
	}
	return nil
}

func finishSyncMetadata(doc *syncdoc.Document, root *os.Root) {
	for _, record := range doc.Metadata.Attachments {
		for name, enrichment := range record.Enrichments {
			enrichment.Stale = !record.Present || !doc.Metadata.Sync.Attachments || record.Local == nil || enrichment.SourceSHA256 == "" || enrichment.SourceSHA256 != record.Local.SHA256
			for _, artifact := range enrichment.Artifacts {
				if _, err := syncdoc.ReadArtifact(root, artifact); err != nil {
					enrichment.Stale = true
				}
			}
			record.Enrichments[name] = enrichment
		}
	}
}
