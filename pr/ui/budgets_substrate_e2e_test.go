// Holds an organization's runs to its budgets on a live Substrate cluster when
// GAVEL_SANDBOX_SUBSTRATE_TEST=1: its runs get environments of its size from its own template, a
// run over a concurrent budget waits for room, and a used-up budget stops running runs and refuses new ones.
package ui

import (
	"bytes"
	"context"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/gavel/budget"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/sandbox/substratetest"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var memTotal = regexp.MustCompile(`MemTotal:\s+(\d+) kB`)

func TestOrgBudgetsHoldOnSubstrate(t *testing.T) {
	cluster := substratetest.FromEnv(t)
	db, store, org := openOrgDB(t)
	dataDir := t.TempDir()
	dispatcher := substrateDispatcher(t, cluster, &RunsAPI{DB: db, Cache: store, Sources: &sandbox.Sources{DataDir: dataDir}})
	client, err := substrate.Dial(substrate.Options{EnvAPI: cluster.EnvAPI, Control: cluster.Control, ControlTokenFile: cluster.TokenFile})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	dispatcher.Templates = &sandbox.OrgTemplates{DB: db, Client: client, Atespace: "ate-env", Base: cluster.Template}
	budgets := budget.Store{DB: db}
	dispatcher.Admission = &sandbox.Admission{Budgets: budgets, RunLifetime: dispatcher.RunLifetime(), Wait: 5 * time.Minute, Recheck: 2 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	size := substrate.Size{CPUMillicores: 1000, MemoryBytes: 1 << 30}
	require.NoError(t, sandbox.SetOrgSize(ctx, db, org, &size))
	template, err := dispatcher.Templates.Resolve(ctx, org)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, client.DeleteTemplate(context.Background(), template.Name)) })
	user := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_ada')`, user).Error)
	scope := tenant.Scope{OrgID: org, UserID: &user}
	repo, sha := pushedRepoWith(t, db, org, dataDir, "sized", map[string]string{
		".gavel.yaml": "ssh:\n  cmd: grep MemTotal /proc/meminfo; sleep 15\n",
	})
	check := sandbox.RunSpec{Kind: "push-check", RepoID: repo, Ref: "refs/heads/main", SHA: sha}
	_, err = budgets.Set(ctx, org, nil, budget.Budget{Resource: budget.CPU, Kind: budget.Concurrent, Amount: 1})
	require.NoError(t, err)

	t.Run("a run over the concurrent budget waits for room, and both run at the org's size", func(t *testing.T) {
		var firstOut, secondOut bytes.Buffer
		first, err := dispatcher.Start(ctx, scope, check, &firstOut)
		require.NoError(t, err)
		second, err := dispatcher.Start(ctx, scope, check, &secondOut)
		require.NoError(t, err)
		status, message := runStatus(t, db, second.ID)
		assert.Equal(t, "queued", status)
		assert.Equal(t, "queued: this organization's runs hold 1 of its 1 concurrent core, and no run is ahead of this one", message)

		for _, run := range []struct {
			run    *sandbox.Run
			output *bytes.Buffer
		}{{first, &firstOut}, {second, &secondOut}} {
			result, err := run.run.Wait(ctx)
			require.NoError(t, err, run.output.String())
			assert.Equal(t, 0, result.ExitCode, run.output.String())
			assertSizedEnvironment(t, run.output.String())
			var recorded struct {
				Template      string
				CPUMillicores int64
				MemoryBytes   int64
			}
			require.NoError(t, db.Raw(`SELECT template, cpu_millicores, memory_bytes FROM sandbox_runs WHERE id = ?`, run.run.ID).Scan(&recorded).Error)
			assert.Equal(t, template.Name, recorded.Template)
			assert.Equal(t, size.CPUMillicores, recorded.CPUMillicores)
			assert.Equal(t, size.MemoryBytes, recorded.MemoryBytes)
		}
		assert.Contains(t, secondOut.String(), "started after waiting for room")
		var startedAfter bool
		require.NoError(t, db.Raw(`SELECT b.started_at >= a.finished_at FROM sandbox_runs a, sandbox_runs b WHERE a.id = ? AND b.id = ?`, first.ID, second.ID).Scan(&startedAfter).Error)
		assert.True(t, startedAfter, "the second run starts once the first gives back its core")
	})

	t.Run("a used-up budget stops the running run and refuses the next", func(t *testing.T) {
		long, sha := pushedRepoWith(t, db, org, dataDir, "long", map[string]string{".gavel.yaml": "ssh:\n  cmd: echo started; sleep 300\n"})
		output := &runLog{}
		running, err := dispatcher.Start(ctx, scope, sandbox.RunSpec{Kind: "push-check", RepoID: long, Ref: "refs/heads/main", SHA: sha}, output)
		require.NoError(t, err)
		require.Eventually(t, func() bool { return strings.Contains(output.String(), "started\n") }, 2*time.Minute, time.Second, "the job runs in its environment")
		require.True(t, actorFrom(t, client, template.Name), "its environment is on the cluster")
		_, err = budgets.Set(ctx, org, nil, budget.Budget{Resource: budget.CPU, Kind: budget.Cumulative, Amount: 0.005, Window: "now/d"})
		require.NoError(t, err)

		require.NoError(t, dispatcher.Enforce(ctx, org))
		waitCtx, stopWaiting := context.WithTimeout(ctx, 2*time.Minute)
		defer stopWaiting()
		_, err = running.Wait(waitCtx)
		require.ErrorIs(t, err, sandbox.ErrBudgetUsedUp, output.String())
		status, message := runStatus(t, db, running.ID)
		assert.Equal(t, "cancelled", status)
		assert.Contains(t, message, "stopped: budget used up: this organization has used ")
		assert.Contains(t, message, "core-hours cpu budget since ")
		require.Eventually(t, func() bool { return !actorFrom(t, client, template.Name) }, time.Minute, 2*time.Second, "the stopped run's environment is deleted")

		refused, err := dispatcher.Start(ctx, scope, check, nil)
		require.ErrorIs(t, err, sandbox.ErrBudgetUsedUp)
		status, message = runStatus(t, db, refused.ID)
		assert.Equal(t, "failed", status)
		assert.Contains(t, message, "budget used up: this organization has used ")
	})
}

// assertSizedEnvironment checks that output, a run's log, reports the memory of a 1 GiB
// environment: the guest's kernel keeps a little of it.
func assertSizedEnvironment(t *testing.T, output string) {
	t.Helper()
	match := memTotal.FindStringSubmatch(output)
	require.NotNil(t, match, "the run reports its memory:\n%s", output)
	kB, err := strconv.ParseInt(match[1], 10, 64)
	require.NoError(t, err)
	assert.LessOrEqual(t, kB, int64(1<<20), "at most 1 GiB")
	assert.Greater(t, kB, int64(3<<18), "more than 0.75 GiB")
}

// runLog is a run's log as it arrives, read while the run writes it.
type runLog struct {
	mu  sync.Mutex
	log bytes.Buffer
}

func (l *runLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.log.Write(p)
}

func (l *runLog) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.log.String()
}

func runStatus(t *testing.T, db *gorm.DB, run uuid.UUID) (status, message string) {
	t.Helper()
	var row struct{ Status, StatusMessage string }
	require.NoError(t, db.Raw(`SELECT status, status_message FROM sandbox_runs WHERE id = ?`, run).Scan(&row).Error)
	return row.Status, row.StatusMessage
}

// actorFrom reports whether an environment from template is still on the cluster.
func actorFrom(t *testing.T, client *substrate.Client, template string) bool {
	t.Helper()
	actors, err := client.ListActors(context.Background(), "ate-env")
	require.NoError(t, err)
	for _, actor := range actors {
		if actor.Template == template {
			return true
		}
	}
	return false
}
