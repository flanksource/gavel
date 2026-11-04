// Runs a hosted todo's plan step end to end on a live Substrate cluster when
// GAVEL_SANDBOX_SUBSTRATE_TEST=1: the whole run executes in an environment, against the real model.
package ui

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/modelkeys"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/sandbox/substratetest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHostedTodoRunsOnSubstrate(t *testing.T) {
	cluster := substratetest.FromEnv(t)
	useModelKeyKeeper(t)
	h := serveHostedRuns(t, map[string]any{
		"envAPI": cluster.EnvAPI, "control": cluster.Control, "tokenFile": cluster.TokenFile,
		"template": cluster.Template, "image": cluster.Image, "trustBundle": substratetest.TrustBundle,
		"egress": sandbox.SubstrateAgentEgress(nil),
	}, func(d *sandbox.Dispatcher) {
		d.Keys = modelkeys.Store{DB: d.DB}
		d.CredentialURI = func(uuid.UUID, string) string { return cluster.AgentCredentialURI }
		d.Limits.Timeout = 10 * time.Minute
	})
	// The key's value stays here: the gateway injects the Secret the credential URI names.
	require.NoError(t, modelkeys.Store{DB: h.db}.Set(t.Context(), h.org, nil, "claude-code-oauth-token", "sk-ant-oat-unused-by-the-gateway"))
	t.Cleanup(func() {
		if t.Failed() {
			reportEnvironment(t, h)
		}
	})

	// A todo the agent can plan without asking: nobody answers a question here.
	runsPlanStep(t, h, "Add a CONTRIBUTING.md saying that changes go through pull requests", 10*time.Minute)
}

// reportEnvironment logs what the runs on h left: each run's environment output, and the tool
// results its transcripts hold that were errors.
func reportEnvironment(t *testing.T, h hostedRuns) {
	var runs []uuid.UUID
	require.NoError(t, h.db.Raw(`SELECT id FROM sandbox_runs WHERE org_id = ?`, h.org).Scan(&runs).Error)
	for _, run := range runs {
		if stored, err := h.logs.Read(t.Context(), "runs/"+run.String()+"/log.txt"); err == nil {
			data, _ := io.ReadAll(stored)
			_ = stored.Close()
			t.Logf("run %s's environment output:\n%s", run, data)
		}
	}
	_ = filepath.WalkDir(TranscriptsDir(h.dataDir), func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		lines := bufio.NewScanner(file)
		lines.Buffer(nil, 16<<20)
		for lines.Scan() {
			if line := lines.Text(); strings.Contains(line, `"is_error":true`) {
				t.Logf("%s: %.2000s", filepath.Base(path), line)
			}
		}
		return nil
	})
}
