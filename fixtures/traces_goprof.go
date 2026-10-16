package fixtures

import (
	"context"
	"errors"
	"fmt"

	"github.com/flanksource/commons-db/tracing/goprof"

	"github.com/flanksource/gavel/testrunner/parsers"
)

// GoProfileFile is a pprof file a fixture's command wrote for its
// `goProfiles:` entry Name: Profile (one of goprof.Profiles) of process PID.
type GoProfileFile struct {
	Name    string
	Path    string
	Profile string
	PID     int
}

// IngestGoProfile reads file into a goprof trace of its own, which the
// fixture's traces carry and which downloads as pprof. Nil-safe only as an
// error: a fixture with goProfiles under --profile is always armed.
func (t *FixtureTraces) IngestGoProfile(ctx context.Context, file GoProfileFile) error {
	if t == nil {
		return fmt.Errorf("goProfiles.%s: no trace host to ingest %s into", file.Name, file.Path)
	}
	result, err := t.ingest(ctx, file)
	if err != nil {
		result.Error = err.Error()
	}
	t.profiles = append(t.profiles, result)
	if err != nil {
		return fmt.Errorf("goProfiles.%s: %w", file.Name, err)
	}
	return nil
}

// ingest reads file in one checkpoint of a goprof capture started for it: a
// goprof capture reads only the profiles its checkpoints name.
func (t *FixtureTraces) ingest(ctx context.Context, file GoProfileFile) (parsers.TraceResult, error) {
	result := parsers.TraceResult{RunID: t.host.RunID()}
	capture, err := t.host.Start(goprof.Kind, nil)
	if err != nil {
		return result, err
	}
	checkpointErr := capture.Checkpoint(ctx, goprof.CheckpointParams{
		PID: file.PID, Source: goprof.SourceFile, Profile: file.Profile, Path: file.Path, Reason: "goProfiles." + file.Name,
	})
	end, err := capture.Stop(ctx)
	if err = errors.Join(checkpointErr, err); err != nil {
		return result, err
	}
	result.TraceRef = capture.Ref(0, end)
	summary, err := t.host.Summarize(ctx, result.TraceRef)
	if err != nil {
		return result, err
	}
	if result.Summary, err = plainSummary(summary); err != nil {
		return result, err
	}
	count, ok := result.Summary["count"].(int64)
	if !ok {
		return result, fmt.Errorf("summary count is %T, not a number", result.Summary["count"])
	}
	result.Count = int(count)
	return result, nil
}

// wantsGoProfiles reports whether --profile ingests the fixture's
// `goProfiles:`, which needs it armed with a trace host.
func (r *Runner) wantsGoProfiles(fixture FixtureTest) bool {
	return r.options.Profile && len(fixture.ExecBase().GoProfiles) > 0
}
