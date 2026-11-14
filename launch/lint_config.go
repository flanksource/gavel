package launch

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"

	"github.com/flanksource/gavel/lint"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/verify"
)

type LinterChoice struct {
	Name     string   `json:"name"`
	Includes []string `json:"includes"`
}

type LintFilterRequest struct {
	PreparationID string                  `json:"preparationId"`
	Results       []*linters.LinterResult `json:"results"`
}

type LintIgnoreRequest struct {
	PreparationID string `json:"preparationId"`
	verify.LintIgnoreRule
}

func readLintRequest(w http.ResponseWriter, r *http.Request, input any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(input); err != nil {
		return &InputError{Err: err}
	}
	if decoder.Decode(new(any)) != io.EOF {
		return &InputError{Err: errors.New("expected exactly one JSON object")}
	}
	return nil
}

func (s *Service) handleLinters(w http.ResponseWriter, r *http.Request) {
	ws, err := s.options.Workspaces.Get(r.URL.Query().Get("preparationId"), requestOwner(r), false)
	if err != nil {
		writeError(w, err)
		return
	}
	registry, err := lint.DefaultRegistry(ws.Directory, ws.Directory)
	if err != nil {
		writeError(w, err)
		return
	}
	choices := []LinterChoice{}
	for _, name := range registry.List() {
		linter, _ := registry.Get(name)
		choices = append(choices, LinterChoice{Name: name, Includes: linter.DefaultIncludes()})
	}
	sort.Slice(choices, func(i, j int) bool { return choices[i].Name < choices[j].Name })
	writeJSON(w, http.StatusOK, choices)
}

func (s *Service) handleLintFilter(w http.ResponseWriter, r *http.Request) {
	var input LintFilterRequest
	if err := readLintRequest(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	ws, err := s.options.Workspaces.Get(input.PreparationID, requestOwner(r), false)
	if err != nil {
		writeError(w, err)
		return
	}
	if input.Results == nil {
		writeError(w, &InputError{Err: errors.New("results must be an array")})
		return
	}
	for _, result := range input.Results {
		if result == nil {
			writeError(w, &InputError{Err: errors.New("lint results cannot contain null")})
			return
		}
		if result.WorkDir == "" {
			result.WorkDir = ws.Directory
		}
		if _, err := s.options.Workspaces.allowed(result.WorkDir); err != nil {
			writeError(w, &InputError{Err: err})
			return
		}
	}
	if err := lint.FilterConfiguredViolations(input.Results, ws.Directory); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, input.Results)
}

func (s *Service) handleLintIgnore(w http.ResponseWriter, r *http.Request) {
	var input LintIgnoreRequest
	if err := readLintRequest(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	ws, err := s.options.Workspaces.Get(input.PreparationID, requestOwner(r), false)
	if err != nil {
		writeError(w, err)
		return
	}
	if _, err := s.options.Workspaces.allowed(ws.Directory); err != nil {
		writeError(w, &InputError{Err: err})
		return
	}
	count, err := verify.SaveLintIgnore(ws.Directory, input.LintIgnoreRule)
	if err != nil {
		writeError(w, &InputError{Err: err})
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"rule_count": count})
}
