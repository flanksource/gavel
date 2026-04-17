package runners

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func writePackageJSON(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(body), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
}

func TestJestDetect(t *testing.T) {
	t.Run("config file", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x"}`)
		if err := os.WriteFile(filepath.Join(tmp, "jest.config.js"), []byte("module.exports={};"), 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}
		r := NewJest(tmp)
		got, err := r.Detect(tmp)
		if err != nil || !got {
			t.Fatalf("Detect = %v, %v; want true,nil", got, err)
		}
	})
	t.Run("jest key in package.json", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x","jest":{}}`)
		r := NewJest(tmp)
		got, _ := r.Detect(tmp)
		if !got {
			t.Fatal("expected detect=true when package.json has jest key")
		}
	})
	t.Run("jest devDep + test file", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x","devDependencies":{"jest":"^29"}}`)
		if err := os.WriteFile(filepath.Join(tmp, "sum.test.js"), []byte("test('ok',()=>{});"), 0o644); err != nil {
			t.Fatal(err)
		}
		r := NewJest(tmp)
		if got, _ := r.Detect(tmp); !got {
			t.Fatal("expected detect=true when jest dep + test file present")
		}
	})
	t.Run("jest devDep alone does not detect", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x","devDependencies":{"jest":"^29"}}`)
		r := NewJest(tmp)
		if got, _ := r.Detect(tmp); got {
			t.Fatal("expected detect=false when jest is declared but no tests exist")
		}
	})
	t.Run(".jestrc bare rc file", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x"}`)
		if err := os.WriteFile(filepath.Join(tmp, ".jestrc"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		r := NewJest(tmp)
		if got, _ := r.Detect(tmp); !got {
			t.Fatal("expected detect=true for bare .jestrc file")
		}
	})
	t.Run(".jestrc.json", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x"}`)
		if err := os.WriteFile(filepath.Join(tmp, ".jestrc.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		r := NewJest(tmp)
		if got, _ := r.Detect(tmp); !got {
			t.Fatal("expected detect=true for .jestrc.json")
		}
	})
	t.Run("no config", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x"}`)
		r := NewJest(tmp)
		got, _ := r.Detect(tmp)
		if got {
			t.Fatal("expected detect=false without jest config")
		}
	})
}

func TestJestDiscoverPackages_Monorepo(t *testing.T) {
	tmp := t.TempDir()
	writePackageJSON(t, tmp, `{"name":"root","private":true,"workspaces":["apps/*"]}`)
	writePackageJSON(t, filepath.Join(tmp, "apps/web"), `{"name":"web","jest":{}}`)
	writePackageJSON(t, filepath.Join(tmp, "apps/api"), `{"name":"api"}`)
	if err := os.WriteFile(filepath.Join(tmp, "apps/api", "jest.config.js"), []byte("module.exports={};"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	writePackageJSON(t, filepath.Join(tmp, "apps/mobile"), `{"name":"mobile"}`)

	r := NewJest(tmp)
	pkgs, err := r.DiscoverPackages(tmp, true)
	if err != nil {
		t.Fatalf("DiscoverPackages: %v", err)
	}
	sort.Strings(pkgs)
	want := []string{"./apps/api", "./apps/web"}
	if strings.Join(pkgs, ",") != strings.Join(want, ",") {
		t.Fatalf("packages = %v, want %v", pkgs, want)
	}
}

func TestJestBuildCommand(t *testing.T) {
	tmp := t.TempDir()
	writePackageJSON(t, tmp, `{"name":"x","jest":{}}`)
	r := NewJest(tmp)
	tr, err := r.BuildCommand(".", "--runInBand")
	if err != nil {
		t.Fatalf("BuildCommand: %v", err)
	}
	if tr.Framework != parsers.Jest {
		t.Errorf("framework = %q", tr.Framework)
	}
	if tr.ReportPath == "" || !strings.Contains(tr.ReportPath, ".jest/") {
		t.Errorf("report path = %q", tr.ReportPath)
	}
	joined := strings.Join(tr.Process.Args, " ")
	if !strings.Contains(joined, "jest --json --outputFile=") {
		t.Errorf("args missing jest invocation: %v", tr.Process.Args)
	}
	if !strings.Contains(joined, "--runInBand") {
		t.Errorf("extra args dropped: %v", tr.Process.Args)
	}
	if !tr.Process.SucceedOnNonZero {
		t.Error("SucceedOnNonZero must be true")
	}
}

func TestDetectPackageManager(t *testing.T) {
	cases := []struct {
		name string
		file string
		cmd  string
		pre  []string
	}{
		{"pnpm", "pnpm-lock.yaml", "pnpm", []string{"exec"}},
		{"yarn", "yarn.lock", "yarn", nil},
		{"npm", "package-lock.json", "npm", []string{"exec", "--"}},
		{"bun", "bun.lockb", "bun", []string{"x"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			if err := os.WriteFile(filepath.Join(tmp, tc.file), []byte(""), 0o644); err != nil {
				t.Fatalf("write lockfile: %v", err)
			}
			cmd, pre := detectPackageManager(tmp)
			if cmd != tc.cmd {
				t.Errorf("cmd = %q, want %q", cmd, tc.cmd)
			}
			if strings.Join(pre, " ") != strings.Join(tc.pre, " ") {
				t.Errorf("pre = %v, want %v", pre, tc.pre)
			}
		})
	}
	t.Run("no lockfile", func(t *testing.T) {
		tmp := t.TempDir()
		cmd, pre := detectPackageManager(tmp)
		if cmd != "npx" || pre != nil {
			t.Errorf("cmd,pre = %q,%v; want npx,nil", cmd, pre)
		}
	})
}
