package reactdoctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/require"
)

func TestParseViolations(t *testing.T) {
	r := NewReactDoctor("/workspace")

	report := `{
		"schemaVersion": 1,
		"ok": true,
		"diagnostics": [
			{
				"filePath": "src/App.tsx",
				"plugin": "react-doctor",
				"rule": "no-array-index-as-key",
				"severity": "warning",
				"title": "Bad key",
				"message": "Avoid array index keys",
				"help": "Use a stable key.",
				"line": 3,
				"column": 5,
				"category": "Performance"
			}
		],
		"summary": {
			"affectedFileCount": 1,
			"totalDiagnosticCount": 1
		}
	}`

	violations, err := r.parseViolations([]byte(report))
	require.NoError(t, err)
	require.Len(t, violations, 1)

	require.Equal(t, "/workspace/src/App.tsx", violations[0].File)
	require.Equal(t, 3, violations[0].Line)
	require.Equal(t, 5, violations[0].Column)
	require.Equal(t, linterName, violations[0].Source)
	require.Equal(t, models.SeverityWarning, violations[0].Severity)
	require.Equal(t, linterName, violations[0].Rule.Package)
	require.Equal(t, "react-doctor/no-array-index-as-key", violations[0].Rule.Method)
	require.Contains(t, *violations[0].Message, "Bad key")
	require.Contains(t, *violations[0].Message, "Avoid array index keys")
	require.Contains(t, *violations[0].Message, "Use a stable key.")
	require.Equal(t, 1, r.GetFileCount())
	require.Equal(t, 1, r.GetRuleCount())
}

func TestParseViolationsIncludesProjectDiagnostics(t *testing.T) {
	r := NewReactDoctor("/workspace")

	report := `{
		"ok": true,
		"projects": [
			{
				"diagnostics": [
					{
						"filePath": "src/index.tsx",
						"plugin": "react-doctor",
						"rule": "component-size",
						"severity": "error",
						"message": "Component is too large"
					}
				]
			}
		]
	}`

	violations, err := r.parseViolations([]byte(report))
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Equal(t, "/workspace/src/index.tsx", violations[0].File)
	require.Equal(t, models.SeverityError, violations[0].Severity)
	require.Equal(t, "react-doctor/component-size", violations[0].Rule.Method)
	require.Equal(t, 1, r.GetFileCount())
	require.Equal(t, 1, r.GetRuleCount())
}

func TestParseViolationsAbsoluteFilePath(t *testing.T) {
	r := NewReactDoctor("/workspace")

	report := `{
		"diagnostics": [
			{
				"filePath": "/abs/path/App.tsx",
				"plugin": "react-doctor",
				"rule": "memoization",
				"severity": "info",
				"message": "Consider memoization"
			}
		]
	}`

	violations, err := r.parseViolations([]byte(report))
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Equal(t, "/abs/path/App.tsx", violations[0].File)
	require.Equal(t, models.SeverityInfo, violations[0].Severity)
}

func TestParseViolationsInvalidJSON(t *testing.T) {
	r := NewReactDoctor("/workspace")

	_, err := r.parseViolations([]byte(`not json`))
	require.Error(t, err)
}

func TestParseViolationsErrorReport(t *testing.T) {
	r := NewReactDoctor("/workspace")

	_, err := r.parseViolations([]byte(`{"ok":false,"error":{"message":"bad config"}}`))
	require.ErrorContains(t, err, "bad config")
}

func TestHasDefaultActivationDetectsReactPackages(t *testing.T) {
	tests := []struct {
		name    string
		pkgJSON string
		want    bool
	}{
		{
			name:    "react dependency",
			pkgJSON: `{"dependencies":{"react":"^19.0.0"}}`,
			want:    true,
		},
		{
			name:    "react dev plugin",
			pkgJSON: `{"devDependencies":{"@vitejs/plugin-react":"latest"}}`,
			want:    true,
		},
		{
			name:    "next framework",
			pkgJSON: `{"dependencies":{"next":"latest"}}`,
			want:    true,
		},
		{
			name:    "non react package",
			pkgJSON: `{"dependencies":{"vue":"latest"}}`,
			want:    false,
		},
		{
			name:    "reactDoctor config is not a react dependency",
			pkgJSON: `{"reactDoctor":{}}`,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(workDir, packageJSONFile), []byte(tt.pkgJSON), 0o644))

			require.Equal(t, tt.want, NewReactDoctor(workDir).HasDefaultActivation(workDir))
		})
	}
}

func TestExecutableCandidatesDetectPackageManager(t *testing.T) {
	t.Run("pnpm lock", func(t *testing.T) {
		repo, app := reactDoctorFixtureProject(t)
		require.NoError(t, os.WriteFile(filepath.Join(repo, "pnpm-lock.yaml"), nil, 0o644))

		require.Equal(t, []string{"pnpx", linterName}, NewReactDoctor(app).ExecutableCandidates(app))
	})

	t.Run("npm lock", func(t *testing.T) {
		repo, app := reactDoctorFixtureProject(t)
		require.NoError(t, os.WriteFile(filepath.Join(repo, "package-lock.json"), nil, 0o644))

		require.Equal(t, []string{"npx", linterName}, NewReactDoctor(app).ExecutableCandidates(app))
	})

	t.Run("npm shrinkwrap", func(t *testing.T) {
		repo, app := reactDoctorFixtureProject(t)
		require.NoError(t, os.WriteFile(filepath.Join(repo, npmShrinkwrap), nil, 0o644))

		require.Equal(t, []string{"npx", linterName}, NewReactDoctor(app).ExecutableCandidates(app))
	})

	t.Run("no lock", func(t *testing.T) {
		_, app := reactDoctorFixtureProject(t)

		require.Equal(t, []string{linterName}, NewReactDoctor(app).ExecutableCandidates(app))
	})
}

func TestDryRunCommandUsesPNPXWithLatestPackage(t *testing.T) {
	workDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "pnpm-lock.yaml"), nil, 0o644))

	r := NewReactDoctor(workDir)
	r.SetOptions(linters.RunOptions{WorkDir: workDir, ForceJSON: true})

	cmd, args := r.DryRunCommand()
	require.Equal(t, "pnpx", cmd)
	require.Equal(t, packageSpec, args[0])
	require.Contains(t, args, ".")
	require.Contains(t, args, "--json")
	require.Contains(t, args, "--json-compact")
	require.Contains(t, args, "--no-telemetry")
	require.Contains(t, args, "--no-color")
	require.Contains(t, args, "--blocking")
	require.Contains(t, args, "none")
}

func TestDryRunCommandUsesNPXExecutable(t *testing.T) {
	workDir := t.TempDir()

	r := NewReactDoctor(workDir)
	r.SetOptions(linters.RunOptions{
		WorkDir:    workDir,
		Executable: filepath.Join(t.TempDir(), "npx"),
		Files:      []string{"src/App.tsx"},
		ForceJSON:  true,
	})

	cmd, args := r.DryRunCommand()
	require.Equal(t, "npx", filepath.Base(cmd))
	require.Equal(t, packageSpec, args[0])
	require.Contains(t, args, ".")
	require.NotContains(t, args, "--changed-files-from")
}

func reactDoctorFixtureProject(t *testing.T) (string, string) {
	t.Helper()

	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))
	app := filepath.Join(repo, "apps", "web")
	require.NoError(t, os.MkdirAll(app, 0o755))
	return repo, app
}
