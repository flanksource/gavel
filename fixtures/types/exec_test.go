package types

import (
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
				SourceDir: "/home/user/fixtures",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveWorkDir(tt.fixture, tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}
