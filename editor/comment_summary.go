package editor

import (
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/flanksource/gavel/comments"
)

// RecentLimit caps the navbar's recent-comments list.
const RecentLimit = 20

type FileCommentCounts struct {
	Total  int `json:"total"`
	Unread int `json:"unread"`
}

// RecentComment is one thread's latest activity: its newest reply, or the
// comment itself when nobody has replied.
type RecentComment struct {
	FileID    string         `json:"fileId"`
	CommentID string         `json:"commentId"`
	Body      string         `json:"body"`
	Author    comments.Actor `json:"author"`
	At        time.Time      `json:"at"`
	Status    string         `json:"status"`
	Unread    bool           `json:"unread"`
}

// CommentSummary is every edited file's comment counts plus the most recent threads.
type CommentSummary struct {
	Files  map[string]FileCommentCounts `json:"files"`
	Recent []RecentComment              `json:"recent"`
}

func latestActivity(c comments.Comment) (string, comments.Actor, time.Time) {
	body, author, at := c.Body, c.Author, c.CreatedAt
	for _, message := range c.Messages {
		if message.CreatedAt.After(at) {
			body, author, at = message.Body, message.Author, message.CreatedAt
		}
	}
	return body, author, at
}

func (s *Server) commentSummary(w http.ResponseWriter, r *http.Request) {
	service, err := s.OpenComments(r.Context())
	if err != nil {
		respond(w, nil, err)
		return
	}
	summary := CommentSummary{Files: map[string]FileCommentCounts{}, Recent: []RecentComment{}}
	for _, file := range s.Store.Files() {
		if file.Kind != KindDocument {
			continue
		}
		target, err := targetOf(r.Context(), file.Path)
		if errors.As(err, new(NoProjectError)) {
			continue // no project, so no threads yet; opening the file prompts to create one
		}
		if err != nil {
			respond(w, nil, err)
			return
		}
		rows, reads, err := fileComments(r.Context(), service, target)
		if err != nil {
			respond(w, nil, err)
			return
		}
		if len(rows) == 0 {
			continue
		}
		counts := FileCommentCounts{Total: len(rows)}
		for _, c := range rows {
			unread := reads[c.ID] < c.Version
			if unread {
				counts.Unread++
			}
			body, author, at := latestActivity(c)
			summary.Recent = append(summary.Recent, RecentComment{FileID: file.ID, CommentID: c.ID, Body: body, Author: author, At: at, Status: c.Status, Unread: unread})
		}
		summary.Files[file.ID] = counts
	}
	sort.SliceStable(summary.Recent, func(i, j int) bool { return summary.Recent[i].At.After(summary.Recent[j].At) })
	if len(summary.Recent) > RecentLimit {
		summary.Recent = summary.Recent[:RecentLimit]
	}
	respond(w, summary, nil)
}

// markFileCommentsRead marks every comment on a file read at its current version:
// the file is open, so its threads are in front of the reader.
func (s *Server) markFileCommentsRead(w http.ResponseWriter, r *http.Request) {
	doc, err := s.Store.Read(r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	target, err := targetOf(r.Context(), doc.Path)
	if err != nil {
		respond(w, nil, err)
		return
	}
	service, err := s.OpenComments(r.Context())
	if err != nil {
		respond(w, nil, err)
		return
	}
	rows, reads, err := fileComments(r.Context(), service, target)
	if err != nil {
		respond(w, nil, err)
		return
	}
	marked := 0
	for _, c := range rows {
		if reads[c.ID] >= c.Version {
			continue
		}
		if _, err := service.MarkRead(r.Context(), c.ID, c.Version); err != nil {
			respond(w, nil, err)
			return
		}
		marked++
	}
	respond(w, map[string]int{"marked": marked}, nil)
}
