package migrategrite

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/flanksource/gavel/todos/griteexport"
)

type ExportCommand func(ctx context.Context, workDir, binary string, args ...string) ([]byte, error)

// LiveExporter immediately copies Grite's fixed export file after the command
// returns. The mutex prevents two imports in this process from racing that
// file; operators must still freeze external Grite writers before finalizing.
type LiveExporter struct {
	WorkDir  string
	Binary   string
	Runner   ExportCommand
	ReadFile func(string) ([]byte, error)
}

var liveExportMu sync.Mutex

func (exporter LiveExporter) Export(ctx context.Context, sinceMS int64) (griteexport.Snapshot, []byte, griteexport.Result, error) {
	liveExportMu.Lock()
	defer liveExportMu.Unlock()

	workDir := strings.TrimSpace(exporter.WorkDir)
	if workDir == "" {
		workDir = "."
	}
	binary := strings.TrimSpace(exporter.Binary)
	if binary == "" {
		path, err := exec.LookPath("grite")
		if err != nil {
			return griteexport.Snapshot{}, nil, griteexport.Result{}, fmt.Errorf("find grite for one-off import: %w", err)
		}
		binary = path
	}
	runner := exporter.Runner
	if runner == nil {
		runner = runExportCommand
	}
	if sinceMS < 0 {
		sinceMS = 0
	}
	rawResult, err := runner(ctx, workDir, binary,
		"--no-daemon", "export", "--format", "json", "--since", strconv.FormatInt(sinceMS, 10), "--json")
	if err != nil {
		return griteexport.Snapshot{}, nil, griteexport.Result{}, err
	}
	result, err := decodeExportResult(rawResult)
	if err != nil {
		return griteexport.Snapshot{}, nil, griteexport.Result{}, err
	}
	if strings.TrimSpace(result.OutputPath) == "" {
		return griteexport.Snapshot{}, nil, griteexport.Result{}, errors.New("Grite export returned no output path")
	}
	path := result.OutputPath
	if !filepath.IsAbs(path) {
		path = filepath.Join(workDir, path)
	}
	readFile := exporter.ReadFile
	if readFile == nil {
		readFile = os.ReadFile
	}
	source, err := readFile(path)
	if err != nil {
		return griteexport.Snapshot{}, nil, griteexport.Result{}, fmt.Errorf("copy Grite export %s: %w", path, err)
	}
	snapshot, err := griteexport.DecodeFile(source)
	if err != nil {
		return griteexport.Snapshot{}, nil, griteexport.Result{}, err
	}
	if snapshot.Meta.SchemaVersion != 0 && snapshot.Meta.EventCount != len(snapshot.Events) {
		return griteexport.Snapshot{}, nil, griteexport.Result{}, fmt.Errorf("Grite export metadata reports %d events, export contains %d", snapshot.Meta.EventCount, len(snapshot.Events))
	}
	if result.EventCount != len(snapshot.Events) {
		return griteexport.Snapshot{}, nil, griteexport.Result{}, fmt.Errorf("Grite command reported %d events, export contains %d", result.EventCount, len(snapshot.Events))
	}
	return snapshot, append([]byte(nil), source...), result, nil
}

func runExportCommand(ctx context.Context, workDir, binary string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, binary, args...)
	command.Dir = workDir
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("grite export: %s", message)
	}
	return stdout.Bytes(), nil
}

func decodeExportResult(raw []byte) (griteexport.Result, error) {
	var envelope struct {
		OK    bool               `json:"ok"`
		Data  griteexport.Result `json:"data"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return griteexport.Result{}, fmt.Errorf("decode Grite export result: %w", err)
	}
	if !envelope.OK {
		if envelope.Error != nil {
			return griteexport.Result{}, fmt.Errorf("Grite export %s: %s", envelope.Error.Code, envelope.Error.Message)
		}
		return griteexport.Result{}, errors.New("Grite export failed")
	}
	return envelope.Data, nil
}
