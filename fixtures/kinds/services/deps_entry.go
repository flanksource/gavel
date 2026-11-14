package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/flanksource/deps/start"
	"github.com/flanksource/deps/start/state"

	"github.com/flanksource/gavel/fixtures"
)

// depsService is a deps catalog service started with its own StateDir under
// the step's TmpDir, so its data, logs and state never touch ~/.deps.
type depsService struct {
	borrowed *sql.DB
	catalog  string
	stateDir string
}

// startDeps starts the catalog service on a free port and blocks until
// deps/start reports it ready. Docker and helm runtimes get an ephemeral
// volume; the binary runtime keeps its data dir inside StateDir.
func startDeps(ctx context.Context, step *fixtures.StepContext, entry Entry) (*depsService, map[string]any, error) {
	if svc, outputs, err := existingConnection(ctx, step, entry); err != nil || svc != nil {
		return svc, outputs, err
	}
	port, err := freePort()
	if err != nil {
		return nil, nil, fmt.Errorf("entry %q: allocate port: %w", entry.Name, err)
	}
	svc := &depsService{catalog: entry.Deps, stateDir: filepath.Join(step.TmpDir, "deps", entry.Name)}
	instance, err := start.Start(ctx, entry.Deps, depsOptions(step, entry, svc.stateDir, port)...)
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("entry %q: deps start %s: %w", entry.Name, entry.Deps, err), svc.stopIfStarted(ctx))
	}
	outputs, err := depsOutputs(entry, instance, port)
	if err != nil {
		return nil, nil, errors.Join(err, svc.stop(ctx))
	}
	return svc, outputs, nil
}

// depsOptions are the deps/start options of an entry. A service deps runs as a
// host process starts from the step's ChildEnv — the fixture Env with its armed
// traces' exports — like every other child a fixture starts; the catalog's env
// and the entry's own env override it, and a container never sees it.
func depsOptions(step *fixtures.StepContext, entry Entry, stateDir string, port int) []start.Option {
	digest := sha256.Sum256([]byte(stateDir))
	opts := []start.Option{
		start.WithStateDir(stateDir), start.WithInstanceID(hex.EncodeToString(digest[:8])), start.WithPort(port),
		start.WithHostEnv(step.ChildEnv()),
	}
	if entry.Version != "" {
		opts = append(opts, start.WithVersion(entry.Version))
	}
	if entry.Runtime != "" {
		kind := start.RuntimeKind(entry.Runtime)
		opts = append(opts, start.WithRuntime(kind))
		if kind == start.RuntimeDocker || kind == start.RuntimeHelm {
			opts = append(opts, start.WithVolumeMode(start.VolumeEphemeral))
		}
	}
	if len(entry.Parameters) > 0 {
		opts = append(opts, start.WithParameters(entry.Parameters))
	}
	if len(entry.Env) > 0 {
		opts = append(opts, start.WithEnv(entry.Env))
	}
	if entry.Wait != nil && entry.Wait.Timeout != nil {
		opts = append(opts, start.WithWaitTimeout(entry.Wait.Timeout.Std()))
	}
	return opts
}

func depsOutputs(entry Entry, instance *start.Instance, port int) (map[string]any, error) {
	conn := instance.Connection
	endpoint, err := url.Parse(conn.URL)
	if err != nil || endpoint.Hostname() == "" {
		return nil, fmt.Errorf("entry %q: deps returned an invalid connection endpoint", entry.Name)
	}
	properties := map[string]any{}
	for k, v := range conn.Properties {
		properties[k] = v
	}
	connection := map[string]any{
		"url": conn.URL, "type": conn.Type, "username": conn.Username, "password": conn.Password, "properties": properties,
	}
	return map[string]any{
		"type": "deps", "source": "provisioned", "ready": true, "url": conn.URL, "port": port, "host": endpoint.Hostname(),
		"username": conn.Username, "password": conn.Password, "connection": connection,
		"runtime": string(instance.Runtime), "log": instance.State.LogFile, "env": stringMap(entry.Env),
	}, nil
}

func (s *depsService) observe(ctx context.Context) map[string]any {
	if s.borrowed != nil {
		err := s.borrowed.PingContext(ctx)
		if err != nil {
			return map[string]any{"ready": false, "error": err.Error()}
		}
		return map[string]any{"ready": true}
	}
	status, err := start.Status(ctx, s.catalog, start.WithStateDir(s.stateDir))
	observed := map[string]any{"ready": err == nil && status == state.StatusRunning, "status": string(status)}
	if err != nil {
		observed["error"] = err.Error()
	}
	return observed
}

func (s *depsService) stop(ctx context.Context) error {
	if s.borrowed != nil {
		return s.borrowed.Close()
	}
	if err := start.Stop(ctx, s.catalog, start.WithStateDir(s.stateDir)); err != nil {
		return fmt.Errorf("stop deps service %s: %w", s.catalog, err)
	}
	return nil
}

// stopIfStarted stops a service whose Start failed after it persisted state
// (e.g. a readiness timeout), so a failed start does not leak a process.
func (s *depsService) stopIfStarted(ctx context.Context) error {
	if _, err := start.Get(ctx, s.catalog, start.WithStateDir(s.stateDir)); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("inspect deps service %s after failed start: %w", s.catalog, err)
	}
	return s.stop(ctx)
}
