package launch

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/flanksource/gavel/resultstore"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
)

type ServiceOptions struct {
	Workspaces *Workspaces
	Store      resultstore.Store
	Schemas    SchemaProvider
	Executor   Executor
}

type activeExecution struct {
	mu     sync.Mutex
	run    resultstore.Run
	cancel context.CancelFunc
	done   chan struct{}
	fault  error
}

type Service struct {
	options ServiceOptions
	secret  []byte
	mu      sync.Mutex
	active  map[string]*activeExecution
}

func NewService(options ServiceOptions) (*Service, error) {
	if options.Workspaces == nil || options.Store == nil || options.Schemas == nil || options.Executor == nil {
		return nil, errors.New("launch requires workspace, store, schema and executor adapters")
	}
	secret, err := loadSecret(options.Workspaces.policy.Root)
	if err != nil {
		return nil, err
	}
	return &Service{options: options, secret: secret, active: map[string]*activeExecution{}}, nil
}

func loadSecret(root string) (_ []byte, result error) {
	dir := filepath.Join(root, ".gavel")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(filepath.Join(dir, ".launch-key.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return nil, err
	}
	defer func() { result = errors.Join(result, syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)) }()
	path := filepath.Join(dir, ".launch-key")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		key, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if len(key) != 32 {
			return nil, errors.New("invalid persisted launch key")
		}
		return key, nil
	}
	if err != nil {
		return nil, err
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		file.Close()
		return nil, err
	}
	_, writeErr := file.Write(key)
	if err := errors.Join(writeErr, file.Sync(), file.Close()); err != nil {
		return nil, err
	}
	directory, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	if err := errors.Join(directory.Sync(), directory.Close()); err != nil {
		return nil, err
	}
	return key, nil
}

func (s *Service) digest(data []byte) string {
	hash := hmac.New(sha256.New, s.secret)
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *Service) Start(ctx context.Context, request Request, key, owner string) (resultstore.Run, error) {
	if key == "" {
		return resultstore.Run{}, &InputError{Err: errors.New("Idempotency-Key is required")}
	}
	input, err := json.Marshal(request)
	if err != nil {
		return resultstore.Run{}, err
	}
	scope, err := s.options.Workspaces.Scope(request, owner)
	if err != nil {
		return resultstore.Run{}, &InputError{Err: err}
	}
	admission := s.digest([]byte(owner + "\x00" + scope + "\x00" + key))
	inputHash := s.digest(input)
	s.mu.Lock()
	defer s.mu.Unlock()
	runs, err := s.options.Store.List(ctx, resultstore.ListOptions{})
	if err != nil {
		return resultstore.Run{}, err
	}
	for _, run := range runs {
		if run.Admission != admission {
			continue
		}
		if run.InputHash != inputHash {
			return resultstore.Run{}, resultstore.ErrConflict
		}
		return run.Public(), nil
	}
	var ws PreparedWorkspace
	if request.Workspace != nil {
		ws, err = s.options.Workspaces.Prepare(ctx, *request.Workspace, owner)
		if err != nil {
			return resultstore.Run{}, &InputError{Err: err}
		}
		request.PreparationID = ws.ID
	}
	ws, err = s.options.Workspaces.Get(request.PreparationID, owner, true)
	if err != nil {
		return resultstore.Run{}, &InputError{Err: err}
	}
	execution, err := s.options.Executor.Prepare(ctx, ws, request.Config)
	if err != nil {
		return resultstore.Run{}, errors.Join(&InputError{Err: err}, s.options.Workspaces.Release(ws.ID))
	}
	kind := "test"
	if request.Config.Lint != nil {
		kind = "lint"
		if request.Config.Test != nil {
			kind = "test+lint"
		}
	}
	dispatcher, err := native.LocalOwner()
	if err != nil {
		return resultstore.Run{}, errors.Join(err, s.options.Workspaces.Release(ws.ID))
	}
	now := time.Now().UTC()
	dispatcher.HeartbeatAt = &now
	dispatcherJSON, err := json.Marshal(dispatcher)
	if err != nil {
		return resultstore.Run{}, errors.Join(err, s.options.Workspaces.Release(ws.ID))
	}
	run := resultstore.Run{
		Metadata: resultstore.Metadata{ExecutionID: uuid.NewString(), Workspace: scope, Kind: kind, Started: now, Labels: request.Labels, Commit: ws.Commit, PatchDigest: ws.PatchDigest},
		Status:   resultstore.Status{State: "queued", Running: true, StopSupported: true, LintRun: request.Config.Lint != nil}, Admission: admission, InputHash: inputHash,
		Owner: s.digest([]byte(owner)), Dispatcher: dispatcherJSON,
		PreparationID: ws.ID,
	}
	run.Metadata.Linters = linterProvenance(request.Config.Lint)
	run, created, err := s.options.Store.Create(ctx, run)
	if err != nil || !created {
		return run.Public(), errors.Join(err, s.options.Workspaces.Release(ws.ID))
	}
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	active := &activeExecution{run: run, cancel: cancel, done: make(chan struct{})}
	s.active[run.Metadata.ExecutionID] = active
	go s.execute(runCtx, active, ws, execution)
	return run.Public(), nil
}

func (s *Service) execute(ctx context.Context, active *activeExecution, ws PreparedWorkspace, execution Execution) {
	defer close(active.done)
	defer active.cancel()
	publish := func(update resultstore.Run) error {
		active.mu.Lock()
		defer active.mu.Unlock()
		if active.fault != nil {
			return active.fault
		}
		if update.Tests != nil {
			active.run.Tests = update.Tests
		}
		if update.RemoveLint != nil {
			remove := update.RemoveLint
			kept := active.run.Lint[:0]
			for _, item := range active.run.Lint {
				if item.ConfigurationID != remove.ConfigurationID || item.Linter != remove.Linter || item.WorkDir != remove.WorkDir {
					kept = append(kept, item)
				}
			}
			active.run.Lint = kept
		}
		for _, incoming := range update.Lint {
			found := false
			for index, previous := range active.run.Lint {
				if previous.ConfigurationID == incoming.ConfigurationID && previous.Linter == incoming.Linter && previous.WorkDir == incoming.WorkDir {
					active.run.Lint[index], found = incoming, true
					break
				}
			}
			if !found {
				active.run.Lint = append(active.run.Lint, incoming)
			}
		}
		active.run.Status.State = "running"
		var dispatcher native.RunOwner
		if err := json.Unmarshal(active.run.Dispatcher, &dispatcher); err != nil {
			return err
		}
		now := time.Now().UTC()
		dispatcher.HeartbeatAt = &now
		data, marshalErr := json.Marshal(dispatcher)
		if marshalErr != nil {
			return marshalErr
		}
		active.run.Dispatcher = data
		updated, err := s.persistActive(context.WithoutCancel(ctx), active)
		if err != nil {
			active.fault = fmt.Errorf("persist execution: %w", err)
			active.cancel()
			return active.fault
		}
		active.run = updated
		return nil
	}
	err := publish(resultstore.Run{})
	if err == nil {
		watchCtx, stopWatch := context.WithCancel(ctx)
		watched := make(chan struct{})
		go func() { defer close(watched); s.watch(watchCtx, active, publish) }()
		err = execution.Run(ctx, publish)
		stopWatch()
		<-watched
	}
	active.mu.Lock()
	artifactErr := retainFindingSources(ws.Directory, active.run.Lint)
	active.mu.Unlock()
	err = errors.Join(err, artifactErr, s.options.Workspaces.Release(ws.ID))
	active.mu.Lock()
	defer active.mu.Unlock()
	active.run.Status = resultstore.Status{State: "completed", LintRun: active.run.Status.LintRun}
	if err != nil {
		active.run.Status.State, active.run.Error = "failed", err.Error()
	}
	if ctx.Err() != nil && active.fault == nil {
		active.run.Status.State = "cancelled"
	}
	active.run.Metadata.Ended = time.Now().UTC()
	updated, writeErr := s.persistActive(context.Background(), active)
	if writeErr != nil {
		active.fault = errors.Join(active.fault, fmt.Errorf("persist completion: %w", writeErr))
		return
	}
	active.run = updated
}

func (s *Service) Get(ctx context.Context, id string) (resultstore.Run, error) {
	s.mu.Lock()
	active := s.active[id]
	s.mu.Unlock()
	if active != nil {
		active.mu.Lock()
		fault := active.fault
		active.mu.Unlock()
		if fault != nil {
			return resultstore.Run{}, fault
		}
	}
	run, err := s.options.Store.Get(ctx, id)
	if err == nil && !run.Terminal() && active == nil {
		var owner native.RunOwner
		if err := json.Unmarshal(run.Dispatcher, &owner); err != nil {
			return resultstore.Run{}, err
		}
		if alive, reason := owner.Alive(time.Now()); !alive {
			run.Status.State, run.Status.Running, run.Status.StopSupported = "interrupted", false, false
			run.Error, run.Metadata.Ended = reason, time.Now().UTC()
			run, err = s.options.Store.Update(ctx, run, run.Metadata.Revision)
		}
	}
	return run.Public(), err
}

func (s *Service) owned(ctx context.Context, id, owner string) error {
	run, err := s.options.Store.Get(ctx, id)
	if err != nil {
		return err
	}
	if run.Owner != s.digest([]byte(owner)) {
		return resultstore.ErrNotFound
	}
	return nil
}

func (s *Service) Stop(ctx context.Context, id string) error {
	if _, err := s.options.Store.RequestStop(ctx, id); err != nil {
		return err
	}
	s.mu.Lock()
	active := s.active[id]
	s.mu.Unlock()
	if active != nil {
		active.cancel()
		return nil
	}
	run, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if run.Terminal() {
		return nil
	}
	return nil
}

func (s *Service) Close() error {
	s.mu.Lock()
	active := make([]*activeExecution, 0, len(s.active))
	for _, run := range s.active {
		run.cancel()
		active = append(active, run)
	}
	s.mu.Unlock()
	for _, run := range active {
		<-run.done
	}
	return errors.Join(s.options.Workspaces.Close(), s.options.Store.Close())
}
