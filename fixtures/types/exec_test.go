package types

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestResolveWorkDir(t *testing.T) {
	tests := []struct {
		name     string
		fixture  fixtures.FixtureTest
		opts     fixtures.RunOptions
		expected string
	}{
		{
			name:     "defaults to opts.WorkDir when no CWD or SourceDir",
			fixture:  fixtures.FixtureTest{},
			opts:     fixtures.RunOptions{WorkDir: "/opt/runner"},
			expected: "/opt/runner",
		},
		{
			name: "SourceDir takes precedence over opts.WorkDir",
			fixture: fixtures.FixtureTest{
				SourceDir: "/home/user/fixtures",
			},
			opts:     fixtures.RunOptions{WorkDir: "/opt/runner"},
			expected: "/home/user/fixtures",
		},
		{
			name: "file-level frontmatter CWD relative to SourceDir",
			fixture: fixtures.FixtureTest{
				SourceDir: "/home/user/fixtures",
				FrontMatter: fixtures.FrontMatter{
					ExecFixtureBase: fixtures.ExecFixtureBase{CWD: "./subdir"},
				},
			},
			opts:     fixtures.RunOptions{WorkDir: "/opt/runner"},
			expected: "/home/user/fixtures/subdir",
		},
		{
			name: "test-level CWD overrides frontmatter CWD",
			fixture: fixtures.FixtureTest{
				SourceDir:       "/home/user/fixtures",
				ExecFixtureBase: fixtures.ExecFixtureBase{CWD: "./test-specific"},
				FrontMatter: fixtures.FrontMatter{
					ExecFixtureBase: fixtures.ExecFixtureBase{CWD: "./from-frontmatter"},
				},
			},
			opts:     fixtures.RunOptions{WorkDir: "/opt/runner"},
			expected: "/home/user/fixtures/test-specific",
		},
		{
			name: "absolute CWD used directly",
			fixture: fixtures.FixtureTest{
				SourceDir:       "/home/user/fixtures",
				ExecFixtureBase: fixtures.ExecFixtureBase{CWD: "/absolute/path"},
			},
			opts:     fixtures.RunOptions{WorkDir: "/opt/runner"},
			expected: "/absolute/path",
		},
		{
			name: "dot CWD resolves to base dir",
			fixture: fixtures.FixtureTest{
				SourceDir:       "/home/user/fixtures",
				ExecFixtureBase: fixtures.ExecFixtureBase{CWD: "."},
			},
			opts:     fixtures.RunOptions{WorkDir: "/opt/runner"},
			expected: "/home/user/fixtures",
		},
		{
			name: "relative CWD without SourceDir resolves from opts.WorkDir",
			fixture: fixtures.FixtureTest{
				FrontMatter: fixtures.FrontMatter{
					ExecFixtureBase: fixtures.ExecFixtureBase{CWD: "relative/path"},
				},
			},
			opts:     fixtures.RunOptions{WorkDir: "/opt/runner"},
			expected: "/opt/runner/relative/path",
		},
		{
			name: "frontmatter CWD with absolute path in frontmatter",
			fixture: fixtures.FixtureTest{
				SourceDir: "/home/user/fixtures",
				FrontMatter: fixtures.FrontMatter{
					ExecFixtureBase: fixtures.ExecFixtureBase{CWD: "/tmp/workspace"},
				},
			},
			opts:     fixtures.RunOptions{WorkDir: "/opt/runner"},
			expected: "/tmp/workspace",
		},
		{
			name: "prepared setup relocates the run out of SourceDir",
			fixture: fixtures.FixtureTest{
				SourceDir: "/home/user/fixtures",
			},
			opts: fixtures.RunOptions{
				WorkDir: "/opt/runner",
				Setup:   &fixtures.PreparedSetup{Cwd: "/var/worktrees/abc"},
			},
			expected: "/var/worktrees/abc",
		},
		{
			name: "relative CWD resolves under the prepared setup, not SourceDir",
			fixture: fixtures.FixtureTest{
				SourceDir:       "/home/user/fixtures",
				ExecFixtureBase: fixtures.ExecFixtureBase{CWD: "./subdir"},
			},
			opts: fixtures.RunOptions{
				WorkDir: "/opt/runner",
				Setup:   &fixtures.PreparedSetup{Cwd: "/var/worktrees/abc"},
			},
			expected: "/var/worktrees/abc/subdir",
		},
		{
			name: "absolute CWD still wins over the prepared setup",
			fixture: fixtures.FixtureTest{
				SourceDir:       "/home/user/fixtures",
				ExecFixtureBase: fixtures.ExecFixtureBase{CWD: "/absolute/path"},
			},
			opts: fixtures.RunOptions{
				WorkDir: "/opt/runner",
				Setup:   &fixtures.PreparedSetup{Cwd: "/var/worktrees/abc"},
			},
			expected: "/absolute/path",
		},
		{
			name: "a setup that relocated nothing leaves SourceDir in charge",
			fixture: fixtures.FixtureTest{
				SourceDir: "/home/user/fixtures",
			},
			opts: fixtures.RunOptions{
				WorkDir: "/opt/runner",
				Setup:   &fixtures.PreparedSetup{Env: map[string]string{"TOKEN": "x"}},
			},
			expected: "/home/user/fixtures",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveWorkDir(tt.fixture, tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecFixtureTemplatesCWDBeforeResolvingWorkDir(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0755); err != nil {
		t.Fatalf("create .git: %v", err)
	}
	sourceDir := filepath.Join(root, "fixtures")
	if err := os.Mkdir(sourceDir, 0755); err != nil {
		t.Fatalf("create source dir: %v", err)
	}

	expectedExit := 0
	result := (&ExecFixture{}).Run(context.Background(), fixtures.FixtureTest{
		Name:      "templated cwd",
		SourceDir: sourceDir,
		ExecFixtureBase: fixtures.ExecFixtureBase{
			Exec: "pwd",
			CWD:  "$GIT_ROOT_DIR",
		},
		Expected: fixtures.Expectations{ExitCode: &expectedExit},
	}, fixtures.RunOptions{WorkDir: root})

	expectedPWD, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("eval symlinks: %v", err)
	}

	assert.Equal(t, root, result.CWD)
	assert.Equal(t, expectedPWD+"\n", result.Stdout)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, root, result.Test.TemplateVars["CWD"])
}

// The repository roots must follow the prepared tree. A worktree fixture whose
// GIT_ROOT_DIR still pointed at the markdown's repo would run its commands in
// the worktree while resolving paths in the tree it was isolating itself from —
// and pass.
func TestExecFixtureRerootsOntoThePreparedSetup(t *testing.T) {
	root := t.TempDir()
	mkdir(t, filepath.Join(root, ".git"))
	sourceDir := filepath.Join(root, "fixtures")
	mkdir(t, sourceDir)

	// A separate repository standing in for a prepared worktree.
	worktree := t.TempDir()
	mkdir(t, filepath.Join(worktree, ".git"))

	expectedExit := 0
	result := (&ExecFixture{}).Run(context.Background(), fixtures.FixtureTest{
		Name:      "re-rooted",
		SourceDir: sourceDir,
		ExecFixtureBase: fixtures.ExecFixtureBase{
			// printenv rather than `echo $GIT_ROOT_DIR`: args are templated
			// before the shell sees them, so a `$VAR` form would assert the
			// template variable a second time instead of the child's
			// environment.
			Exec: "sh",
			Args: []string{"-c", "printenv GIT_ROOT_DIR; printenv SETUP_DIR"},
		},
		Expected: fixtures.Expectations{ExitCode: &expectedExit},
	}, fixtures.RunOptions{
		WorkDir: root,
		Setup: &fixtures.PreparedSetup{
			Cwd:   worktree,
			Extra: map[string]any{"commit": "deadbeef"},
		},
	})

	assert.Equal(t, worktree, result.CWD, "the command did not run in the prepared tree")
	assert.Equal(t, worktree+"\n"+worktree+"\n", result.Stdout,
		"GIT_ROOT_DIR/SETUP_DIR still point at the markdown's repository")
	assert.Equal(t, worktree, result.Test.TemplateVars["GIT_ROOT_DIR"])
	assert.Equal(t, worktree, result.Test.TemplateVars["SETUP_DIR"])
	assert.Equal(t, map[string]any{"commit": "deadbeef"}, result.Test.TemplateVars["setup"])
}

// Env precedence, highest first: the fixture's own `env:`, then the setup's,
// then the auto-injected roots, then the inherited process environment.
func TestExecFixtureEnvPrecedence(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("FROM_PROCESS", "process")

	expectedExit := 0
	result := (&ExecFixture{}).Run(context.Background(), fixtures.FixtureTest{
		Name:      "env precedence",
		SourceDir: dir,
		ExecFixtureBase: fixtures.ExecFixtureBase{
			Exec: "sh",
			Args: []string{"-c", "echo $CONTESTED:$FROM_SETUP:$FROM_PROCESS"},
			Env:  map[string]any{"CONTESTED": "from-fixture"},
		},
		Expected: fixtures.Expectations{ExitCode: &expectedExit},
	}, fixtures.RunOptions{
		WorkDir: dir,
		Setup: &fixtures.PreparedSetup{
			Cwd: dir,
			Env: map[string]string{"CONTESTED": "from-setup", "FROM_SETUP": "setup"},
		},
	})

	assert.Equal(t, "from-fixture:setup:process\n", result.Stdout)
}

// A setup may also override a variable the runner would otherwise auto-inject,
// which is the seam a checkout needs to point GIT_ROOT_DIR somewhere the
// filesystem walk cannot find on its own.
func TestExecFixtureSetupOverridesAutoInjectedRoots(t *testing.T) {
	dir := t.TempDir()

	expectedExit := 0
	result := (&ExecFixture{}).Run(context.Background(), fixtures.FixtureTest{
		Name:      "setup overrides roots",
		SourceDir: dir,
		ExecFixtureBase: fixtures.ExecFixtureBase{
			Exec: "printenv",
			Args: []string{"GIT_ROOT_DIR"},
		},
		Expected: fixtures.Expectations{ExitCode: &expectedExit},
	}, fixtures.RunOptions{
		WorkDir: dir,
		Setup: &fixtures.PreparedSetup{
			Cwd: dir,
			Env: map[string]string{"GIT_ROOT_DIR": "/from/setup"},
		},
	})

	assert.Equal(t, "/from/setup\n", result.Stdout)
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
}
