package ui

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/todos/types"
)

type linkedPRDetail struct {
	prDetail
	Repo string `json:"repo"`
}

func linkedPRRepository(ref string) (string, error) {
	if provider.IsAzure(ref) {
		repo, _, err := azuredevops.ParsePRURL(ref)
		if err != nil {
			return "", err
		}
		return repo.URL(), nil
	}
	repo, _, err := github.ParsePRURL(ref)
	return repo, err
}

// linkedPRClient is the client a linked PR is read through: a hosted organization's GitHub App,
// or this host's credentials.
func (s *Server) linkedPRClient(ctx context.Context, url string) (provider.Client, int, error) {
	if s.hosted {
		return s.organizationPRs(ctx, url)
	}
	return provider.ResolveURL(provider.Options{WorkDir: s.ghOpts.WorkDir, Token: s.ghOpts.Token}, url)
}

func (s *Server) handleLinkedPR(w http.ResponseWriter, r *http.Request) {
	link, err := types.PRFromURL(r.URL.Query().Get("url"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	repo, err := linkedPRRepository(link.URL)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	client, number, err := s.linkedPRClient(r.Context(), link.URL)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	status, err := client.Status(r.Context(), number, model.StatusOptions{Threads: true})
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}
	detail := linkedPRDetail{Repo: repo, prDetail: prDetail{PR: status.PR, Runs: status.Runs, Comments: append(status.PR.Comments, status.PR.ReviewThreads...)}}
	if client.Kind() == "github" {
		opts := s.ghOpts
		opts.Repo = detail.Repo
		detail.GavelResults = fetchGavelArtifacts(opts, github.FindGavelArtifacts(detail.Comments))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(detail) //nolint:errcheck
}
