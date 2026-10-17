// Runs push checks end to end in real containers when GAVEL_SANDBOX_RUNNER_IMAGE names a runner
// image: the dispatcher starts the job, which fetches the pushed source from the run API on this
// host, runs the check, reports its results, and exits with the check's verdict.
package ui

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPushChecksRunInTheRunnerImage(t *testing.T) {
	image := os.Getenv("GAVEL_SANDBOX_RUNNER_IMAGE")
	if image == "" {
		t.Skip("set GAVEL_SANDBOX_RUNNER_IMAGE to a runner image to run push checks in containers")
	}
	db, store, org := openOrgDB(t)
	user := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_ada')`, user).Error)
	dataDir := t.TempDir()

	// The jobs' network, and the run API on its gateway, where a job reaches serve.
	network := "gavel-sandbox-e2e-" + uuid.NewString()[:8]
	require.NoError(t, exec.Command("docker", "network", "create", network).Run())
	t.Cleanup(func() { _ = exec.Command("docker", "network", "rm", network).Run() })
	gateway, err := exec.Command("docker", "network", "inspect", "-f", "{{(index .IPAM.Config 0).Gateway}}", network).Output()
	require.NoError(t, err)
	listener, err := net.Listen("tcp", strings.TrimSpace(string(gateway))+":0")
	require.NoError(t, err)
	key := []byte(strings.Repeat("k", 32))
	api := &http.Server{Handler: (&RunsAPI{DB: db, TokenKey: key, Cache: store, Sources: &sandbox.Sources{DataDir: dataDir}}).Handler()}
	go func() { _ = api.Serve(listener) }()
	t.Cleanup(func() { _ = api.Close() })

	logs := blob.Local(t.TempDir())
	dispatcher := &sandbox.Dispatcher{
		DB: db, Image: image, Server: "http://" + listener.Addr().String(), TokenKey: key,
		Logs: func(uuid.UUID) blob.Store { return logs }, Limits: sandbox.Limits{Timeout: 5 * time.Minute},
		Executor: sandbox.NewContainer("docker", sandbox.ContainerOptions{Network: network}),
	}

	for _, check := range []struct {
		name     string
		files    map[string]string
		exitCode func(int) bool
		output   string
	}{
		{
			name: "failing",
			files: map[string]string{
				"go.mod":      "module example.com/app\n\ngo 1.26\n",
				"app_test.go": "package app\n\nimport \"testing\"\n\nfunc TestBroken(t *testing.T) { t.Fatal(\"broken on purpose\") }\n",
			},
			exitCode: func(code int) bool { return code != 0 },
			output:   "TestBroken",
		},
		{
			name: "passing",
			files: map[string]string{
				"go.mod":      "module example.com/app\n\ngo 1.26\n",
				"app_test.go": "package app\n\nimport \"testing\"\n\nfunc TestFine(t *testing.T) {}\n",
				".gavel.yaml": "ssh:\n  cmd: go test -v ./...\n",
			},
			exitCode: func(code int) bool { return code == 0 },
			output:   "--- PASS: TestFine",
		},
	} {
		t.Run(check.name, func(t *testing.T) {
			repo, sha := pushedRepoWith(t, db, org, dataDir, check.name, check.files)
			var output bytes.Buffer
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
			defer cancel()
			run, err := dispatcher.Start(ctx, tenant.Scope{OrgID: org, UserID: &user},
				sandbox.RunSpec{Kind: "push-check", RepoID: repo, Ref: "refs/heads/main", SHA: sha}, &output)
			require.NoError(t, err)
			result, err := run.Wait(ctx)
			require.NoError(t, err, output.String())
			assert.True(t, check.exitCode(result.ExitCode), "exit code %d:\n%s", result.ExitCode, output.String())
			assert.Contains(t, output.String(), check.output)
			var status string
			require.NoError(t, db.Raw(`SELECT status FROM sandbox_runs WHERE id = ?`, run.ID).Scan(&status).Error)
			assert.NotEqual(t, "running", status)
		})
	}
}

// pushedRepoWith records a pushed repository of org whose main branch holds files, and returns
// its ID and commit.
func pushedRepoWith(t *testing.T, db *gorm.DB, org uuid.UUID, dataDir, name string, files map[string]string) (uuid.UUID, string) {
	t.Helper()
	var repo struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source) VALUES (?, ?, 'push') RETURNING id`, org, name).Scan(&repo).Error)
	work := t.TempDir()
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = work
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
		return strings.TrimSpace(string(out))
	}
	git("init", "-q", "-b", "main")
	for file, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(work, file), []byte(content), 0o644))
	}
	git("add", ".")
	git("commit", "-q", "-m", "init")
	git("clone", "-q", "--bare", work, filepath.Join(dataDir, "repos", org.String(), repo.ID.String()+".git"))
	return repo.ID, git("rev-parse", "HEAD")
}
