package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/flanksource/gavel/comments"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ comments.ProviderStore = (*Store)(nil)

func (s *Store) Bindings(ctx context.Context, tenant, id string) ([]comments.Binding, error) {
	var rows []struct{ Payload json.RawMessage }
	if err := s.db.WithContext(ctx).Table("comment_bindings").Where("tenant_id = ? AND comment_id = ?", tenant, id).Order("provider").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]comments.Binding, 0, len(rows))
	for _, row := range rows {
		var b comments.Binding
		if err := json.Unmarshal(row.Payload, &b); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

func (s *Store) Binding(ctx context.Context, tenant, id, provider string) (comments.Binding, error) {
	var row struct{ Payload json.RawMessage }
	err := s.db.WithContext(ctx).Table("comment_bindings").Where("tenant_id = ? AND comment_id = ? AND provider = ?", tenant, id, provider).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return comments.Binding{}, comments.ErrNotFound
	}
	if err != nil {
		return comments.Binding{}, err
	}
	var b comments.Binding
	err = json.Unmarshal(row.Payload, &b)
	return b, err
}
func (s *Store) ApplyRemote(ctx context.Context, input comments.Comment, remote comments.RemoteComment, actor comments.Actor) (comments.Comment, error) {
	var out comments.Comment
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seq, err := sequence(ctx, tx, input.TenantID)
		if err != nil {
			return err
		}
		var previous struct {
			CommentID string
			Payload   json.RawMessage
		}
		q := tx.Table("comment_bindings").Where("tenant_id = ? AND provider = ? AND remote_key = ?", input.TenantID, remote.Binding.Provider, remote.Binding.RemoteKey).Take(&previous)
		if q.Error == nil {
			if input.ID != "" && input.ID != previous.CommentID {
				return comments.ErrConflict
			}
			input.ID = previous.CommentID
			var b comments.Binding
			if err = json.Unmarshal(previous.Payload, &b); err != nil {
				return err
			}
			if b.Revision == remote.Binding.Revision {
				out, err = get(ctx, tx, input.TenantID, input.ID, false)
				return err
			}
			if remote.Binding.MessageIDs == nil {
				remote.Binding.MessageIDs = map[string]string{}
			}
			for id, local := range b.MessageIDs {
				if _, ok := remote.Binding.MessageIDs[id]; !ok {
					remote.Binding.MessageIDs[id] = local
				}
			}
		} else if !errors.Is(q.Error, gorm.ErrRecordNotFound) {
			return q.Error
		}
		if input.ID == "" {
			input.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(input.TenantID+"\x00"+remote.Binding.Provider+"\x00"+remote.Binding.RemoteKey)).String()
		}
		c, err := get(ctx, tx, input.TenantID, input.ID, true)
		created := errors.Is(err, comments.ErrNotFound)
		if err != nil && !created {
			return err
		}
		if created {
			c = input
			c.Author = remote.Author
			c.CreatedAt = remote.CreatedAt
			c.Version = 0
			c.Phase = comments.Active
			c.Status = "open"
		}
		if c.Target.URI != input.Target.URI || c.Target.ScopeURI != input.Target.ScopeURI {
			return comments.ErrConflict
		}
		before := struct {
			Status     string            `json:"status"`
			Outcome    *comments.Outcome `json:"outcome,omitempty"`
			Resolution *comments.Outcome `json:"resolution,omitempty"`
		}{c.Status, c.Outcome, c.Resolution}
		var bound int64
		if err = tx.Table("comment_bindings").Where("tenant_id = ? AND comment_id = ? AND provider = ? AND remote_key <> ?", c.TenantID, c.ID, remote.Binding.Provider, remote.Binding.RemoteKey).Count(&bound).Error; err != nil {
			return err
		}
		if bound > 0 {
			return comments.ErrConflict
		}
		c.Body = remote.Body
		c.Version++
		c.UpdatedAt = time.Now().UTC()
		if remote.Binding.ThreadID != "" {
			c.WorkflowID = comments.NativePR().ID
			c.Status = "open"
			c.Phase = comments.Active
			c.Outcome = nil
			c.Resolution = nil
			if remote.Resolved {
				c.Status = "resolved"
				c.Phase = comments.Final
				c.Outcome = &comments.Outcome{Kind: "resolve", Summary: remote.ResolutionReason, Actor: comments.Actor{Kind: "system", ID: "provider:" + remote.Binding.Provider}, At: c.UpdatedAt}
				if before.Outcome != nil && before.Status == "resolved" {
					c.Outcome = before.Outcome
				}
			}
		} else if created {
			c.WorkflowID = comments.HumanReview().ID
		}
		data, err := json.Marshal(c)
		if err != nil {
			return err
		}
		if created {
			hash := sha256.Sum256([]byte(remote.Binding.RemoteKey))
			err = tx.Exec("INSERT INTO comments (tenant_id,id,scope_uri,target_uri,phase,version,request_id,request_hash,payload) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?::jsonb)", c.TenantID, c.ID, c.Target.ScopeURI, c.Target.URI, c.Phase, c.Version, "native:"+remote.Binding.Provider+":"+remote.Binding.RemoteKey, hex.EncodeToString(hash[:]), string(data)).Error
		} else {
			stored := c
			stored.Messages = nil
			stored.Attachments = nil
			data, err = json.Marshal(stored)
			if err != nil {
				return err
			}
			err = tx.Exec("UPDATE comments SET phase = ?, version = ?, payload = ?::jsonb WHERE tenant_id = ? AND id = ?", c.Phase, c.Version, string(data), c.TenantID, c.ID).Error
		}
		if err != nil {
			return err
		}
		if remote.Binding.MessageIDs == nil {
			remote.Binding.MessageIDs = map[string]string{}
		}
		for _, m := range remote.Messages {
			if _, ok := remote.Binding.MessageIDs[m.ID]; !ok {
				remote.Binding.MessageIDs[m.ID] = uuid.NewSHA1(uuid.NameSpaceOID, []byte(c.ID+":"+m.ID)).String()
			}
		}
		for _, m := range remote.Messages {
			var owner struct {
				CommentID string
				Payload   json.RawMessage
			}
			err := tx.Table("comment_messages").Select("comment_id,payload").Where("tenant_id=? AND id=?", c.TenantID, remote.Binding.MessageIDs[m.ID]).Take(&owner).Error
			if err == nil && owner.CommentID != c.ID {
				return comments.ErrConflict
			}
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			parent := ""
			if m.ParentID != "" && m.ParentID != remote.Binding.RootID {
				var ok bool
				parent, ok = remote.Binding.MessageIDs[m.ParentID]
				if !ok {
					return fmt.Errorf("native reply parent %s is missing", m.ParentID)
				}
			}
			msg := comments.Message{ID: remote.Binding.MessageIDs[m.ID], ParentID: parent, Body: m.Body, Author: m.Author, CreatedAt: m.CreatedAt}
			if owner.Payload != nil {
				var previous comments.Message
				if err = json.Unmarshal(owner.Payload, &previous); err != nil {
					return err
				}
				msg.Content = previous.Content
				msg.Author = previous.Author
				msg.CreatedAt = previous.CreatedAt
			}
			payload, err := json.Marshal(msg)
			if err != nil {
				return err
			}
			err = tx.Exec("INSERT INTO comment_messages (tenant_id,comment_id,id,sequence,payload) SELECT ?, ?, ?, COALESCE(MAX(sequence),0)+1, ?::jsonb FROM comment_messages WHERE tenant_id = ? AND comment_id = ? ON CONFLICT (tenant_id,id) DO UPDATE SET payload = EXCLUDED.payload", c.TenantID, c.ID, msg.ID, string(payload), c.TenantID, c.ID).Error
			if err != nil {
				return err
			}
		}
		data, err = json.Marshal(remote.Binding)
		if err != nil {
			return err
		}
		err = tx.Exec("INSERT INTO comment_bindings (tenant_id,comment_id,provider,remote_key,payload) VALUES (?, ?, ?, ?, ?::jsonb) ON CONFLICT (tenant_id,provider,remote_key) DO UPDATE SET payload = EXCLUDED.payload", c.TenantID, c.ID, remote.Binding.Provider, remote.Binding.RemoteKey, string(data)).Error
		if err != nil {
			return err
		}
		audit, err := json.Marshal(struct {
			Before  any              `json:"before"`
			Binding comments.Binding `json:"binding"`
		}{before, remote.Binding})
		if err != nil {
			return err
		}
		if err = appendEvent(ctx, tx, c.TenantID, seq, comments.Event{CommentID: c.ID, Kind: "provider-sync", Actor: actor, At: c.UpdatedAt, Data: audit}); err != nil {
			return err
		}
		out, err = get(ctx, tx, c.TenantID, c.ID, false)
		return err
	})
	return out, err
}
