package resultstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/parsers"
)

var (
	ErrNotFound = errors.New("execution not found")
	ErrConflict = errors.New("execution revision conflict")
	ErrTerminal = errors.New("execution is terminal")
)

type Metadata struct {
	ExecutionID string                `json:"executionId"`
	Workspace   string                `json:"workspace"`
	Kind        string                `json:"kind"`
	Started     time.Time             `json:"started"`
	Ended       time.Time             `json:"ended,omitempty"`
	Revision    uint64                `json:"revision"`
	Labels      map[string]string     `json:"labels,omitempty"`
	Commit      string                `json:"commit,omitempty"`
	PatchDigest string                `json:"patchDigest,omitempty"`
	Linters     []LinterConfiguration `json:"linters,omitempty"`
}

type LinterConfiguration struct {
	ID            string   `json:"id"`
	Engine        string   `json:"engine"`
	Prompt        string   `json:"prompt,omitempty"`
	PromptVersion string   `json:"promptVersion,omitempty"`
	Presets       []string `json:"presets,omitempty"`
}

type Status struct {
	State         string `json:"state"`
	Running       bool   `json:"running"`
	LintRun       bool   `json:"lint_run,omitempty"`
	StopSupported bool   `json:"stop_supported,omitempty"`
}

type Run struct {
	Metadata      Metadata                `json:"metadata"`
	Status        Status                  `json:"status"`
	Tests         []parsers.Test          `json:"tests"`
	Lint          []*linters.LinterResult `json:"lint,omitempty"`
	Error         string                  `json:"error,omitempty"`
	Admission     string                  `json:"_admission,omitempty"`
	InputHash     string                  `json:"_inputHash,omitempty"`
	Owner         string                  `json:"_owner,omitempty"`
	RemoveLint    *linters.LinterResult   `json:"-"`
	Dispatcher    json.RawMessage         `json:"_dispatcher,omitempty"`
	StopRequested bool                    `json:"_stopRequested,omitempty"`
	PreparationID string                  `json:"_preparationId,omitempty"`
}

func (r Run) Terminal() bool { return !r.Metadata.Ended.IsZero() }

func (r Run) Public() Run {
	r.Admission, r.InputHash, r.Owner, r.Dispatcher = "", "", "", nil
	r.PreparationID = ""
	if r.StopRequested && !r.Terminal() {
		r.Status.State = "stopping"
	}
	r.StopRequested = false
	results := make([]*linters.LinterResult, len(r.Lint))
	for index, result := range r.Lint {
		copy := *result
		copy.Command, copy.Args = "", nil
		results[index] = &copy
	}
	r.Lint = results
	return r
}

type ListOptions struct {
	Workspace string
	Before    string
	Limit     int
	Labels    map[string]string
	Owner     string
}

type Reader interface {
	Get(context.Context, string) (Run, error)
	List(context.Context, ListOptions) ([]Run, error)
}

type Writer interface {
	Create(context.Context, Run) (Run, bool, error)
	Update(context.Context, Run, uint64) (Run, error)
	RequestStop(context.Context, string) (Run, error)
	Import(context.Context, Run) (bool, error)
}

func (s *adapter) RequestStop(ctx context.Context, id string) (Run, error) {
	var run Run
	err := s.lock(func() error {
		var err error
		run, err = s.backend.Get(ctx, id)
		if err != nil || run.Terminal() || run.StopRequested {
			return err
		}
		run.StopRequested = true
		run.Metadata.Revision++
		return s.backend.Put(ctx, run)
	})
	return run, err
}

type TestStore interface {
	Tests(context.Context, string) ([]parsers.Test, error)
}

type LintStore interface {
	Lint(context.Context, string) ([]*linters.LinterResult, error)
}

type Store interface {
	Reader
	Writer
	TestStore
	LintStore
	Close() error
}

type Backend interface {
	Get(context.Context, string) (Run, error)
	Put(context.Context, Run) error
	All(context.Context) ([]Run, error)
	Close() error
}

type adapter struct {
	backend Backend
	lock    func(func() error) error
}

func New(backend Backend, lock func(func() error) error) Store {
	return &adapter{backend: backend, lock: lock}
}

func (s *adapter) Get(ctx context.Context, id string) (Run, error) {
	return s.backend.Get(ctx, id)
}

func (s *adapter) List(ctx context.Context, options ListOptions) ([]Run, error) {
	runs, err := s.backend.All(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(runs, func(i, j int) bool {
		if runs[i].Metadata.Started.Equal(runs[j].Metadata.Started) {
			return runs[i].Metadata.ExecutionID > runs[j].Metadata.ExecutionID
		}
		return runs[i].Metadata.Started.After(runs[j].Metadata.Started)
	})
	result := make([]Run, 0)
	include := options.Before == ""
	for _, run := range runs {
		if options.Owner != "" && run.Owner != options.Owner {
			continue
		}
		if !include {
			include = run.Metadata.ExecutionID == options.Before
			continue
		}
		if options.Workspace != "" && run.Metadata.Workspace != options.Workspace {
			continue
		}
		matches := true
		for key, value := range options.Labels {
			if run.Metadata.Labels[key] != value {
				matches = false
			}
		}
		if !matches {
			continue
		}
		result = append(result, run)
		if options.Limit > 0 && len(result) >= options.Limit {
			break
		}
	}
	if options.Before != "" && !include {
		return nil, fmt.Errorf("%w: unknown history cursor", ErrNotFound)
	}
	return result, nil
}

func (s *adapter) Create(ctx context.Context, run Run) (Run, bool, error) {
	created := false
	err := s.lock(func() error {
		all, err := s.backend.All(ctx)
		if err != nil {
			return err
		}
		for _, previous := range all {
			if previous.Metadata.ExecutionID == run.Metadata.ExecutionID && previous.Admission != run.Admission {
				return ErrConflict
			}
			if previous.Admission != run.Admission || previous.Metadata.Workspace != run.Metadata.Workspace {
				continue
			}
			if previous.InputHash != run.InputHash {
				return ErrConflict
			}
			run = previous
			return nil
		}
		if run.Admission == "" || run.InputHash == "" || run.Metadata.ExecutionID == "" || run.Metadata.Workspace == "" {
			return errors.New("execution admission requires identity, workspace, key and input hash")
		}
		run.Metadata.Revision = 1
		if err := s.backend.Put(ctx, run); err != nil {
			return err
		}
		created = true
		return nil
	})
	return run, created, err
}

func (s *adapter) Update(ctx context.Context, run Run, revision uint64) (Run, error) {
	err := s.lock(func() error {
		previous, err := s.backend.Get(ctx, run.Metadata.ExecutionID)
		if err != nil {
			return err
		}
		if previous.Terminal() {
			return ErrTerminal
		}
		if previous.Metadata.Revision != revision || previous.Admission != run.Admission || previous.InputHash != run.InputHash || previous.Metadata.Workspace != run.Metadata.Workspace || previous.Owner != run.Owner {
			return ErrConflict
		}
		run.Metadata.Revision = revision + 1
		return s.backend.Put(ctx, run)
	})
	return run, err
}

func (s *adapter) Tests(ctx context.Context, id string) ([]parsers.Test, error) {
	run, err := s.Get(ctx, id)
	return run.Tests, err
}

func (s *adapter) Lint(ctx context.Context, id string) ([]*linters.LinterResult, error) {
	run, err := s.Get(ctx, id)
	return run.Lint, err
}

func (s *adapter) Close() error { return s.backend.Close() }
