package workspacefiles

import (
	git "github.com/flanksource/commons-db/git"
	"net/http"
)

type BlameFile struct {
	File
	Blame git.WorkingBlame `json:"blame"`
}

func (s *Service) blame(w http.ResponseWriter, r *http.Request) {
	file, err := s.Read(r.Context(), fileQuery(r))
	if err != nil {
		respond(w, nil, err)
		return
	}
	blame, err := git.BlameWorkingFile(r.Context(), git.WorkingBlameOptions{Directory: file.Root, Path: file.Path, Contents: []byte(file.Source)})
	if err != nil {
		respond(w, nil, err)
		return
	}
	respond(w, BlameFile{File: file, Blame: blame}, nil)
}
