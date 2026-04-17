package runners

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestPlaywrightDetect(t *testing.T) {
	t.Run("config file", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x"}`)
		if err := os.WriteFile(filepath.Join(tmp, "playwright.config.ts"), []byte("export default {};"), 0o644); err != nil {
			t.Fatal(err)
		}
		r := NewPlaywright(tmp)
		if got, _ := r.Detect(tmp); !got {
			t.Fatal("expected detect=true")
		}
	})
	t.Run("playwright-ct.config", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x"}`)
		if err := os.WriteFile(filepath.Join(tmp, "playwright-ct.config.ts"), []byte("export default {};"), 0o644); err != nil {
			t.Fatal(err)
		}
		r := NewPlaywright(tmp)
		if got, _ := r.Detect(tmp); !got {
			t.Fatal("expected detect=true for playwright-ct.config")
		}
	})
	t.Run("@playwright/test devDep + spec file", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x","devDependencies":{"@playwright/test":"^1.40.0"}}`)
		if err := os.WriteFile(filepath.Join(tmp, "home.spec.ts"), []byte("// playwright"), 0o644); err != nil {
			t.Fatal(err)
		}
		r := NewPlaywright(tmp)
		if got, _ := r.Detect(tmp); !got {
			t.Fatal("expected detect=true when playwright dep + spec file present")
		}
	})
	t.Run("@playwright/test devDep alone does not detect", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x","devDependencies":{"@playwright/test":"^1.40.0"}}`)
		r := NewPlaywright(tmp)
		if got, _ := r.Detect(tmp); got {
			t.Fatal("expected detect=false when dep is declared but no spec files exist")
		}
	})
	t.Run("no signals", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x"}`)
		r := NewPlaywright(tmp)
		if got, _ := r.Detect(tmp); got {
			t.Fatal("expected detect=false")
		}
	})
}

func TestPlaywrightDiscoverPackages_Monorepo(t *testing.T) {
	tmp := t.TempDir()
	writePackageJSON(t, tmp, `{"name":"root"}`)
	writePackageJSON(t, filepath.Join(tmp, "e2e/web"), `{"name":"web","devDependencies":{"@playwright/test":"^1.40.0"}}`)
	if err := os.WriteFile(filepath.Join(tmp, "e2e/web", "home.spec.ts"), []byte("// playwright"), 0o644); err != nil {
		t.Fatal(err)
	}
	writePackageJSON(t, filepath.Join(tmp, "packages/shared"), `{"name":"shared"}`)
	r := NewPlaywright(tmp)
	pkgs, err := r.DiscoverPackages(tmp, true)
	if err != nil {
		t.Fatalf("DiscoverPackages: %v", err)
	}
	sort.Strings(pkgs)
	want := []string{"./e2e/web"}
	if strings.Join(pkgs, ",") != strings.Join(want, ",") {
		t.Fatalf("packages = %v, want %v", pkgs, want)
	}
}

func TestPlaywrightBuildCommand(t *testing.T) {
	tmp := t.TempDir()
	writePackageJSON(t, tmp, `{"name":"x","devDependencies":{"@playwright/test":"^1.40.0"}}`)
	if err := os.WriteFile(filepath.Join(tmp, "home.spec.ts"), []byte("// playwright"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := NewPlaywright(tmp)
	tr, err := r.BuildCommand(".")
	if err != nil {
		t.Fatal(err)
	}
	if tr.Framework != parsers.Playwright {
		t.Errorf("framework = %q", tr.Framework)
	}
	joined := strings.Join(tr.Process.Args, " ")
	if !strings.Contains(joined, "playwright test --reporter=json") {
		t.Errorf("args missing playwright invocation: %v", tr.Process.Args)
	}
	if tr.Process.Env["PLAYWRIGHT_JSON_OUTPUT_NAME"] != tr.ReportPath {
		t.Errorf("env PLAYWRIGHT_JSON_OUTPUT_NAME=%q, want %q", tr.Process.Env["PLAYWRIGHT_JSON_OUTPUT_NAME"], tr.ReportPath)
	}
	if !filepath.IsAbs(tr.ReportPath) {
		t.Errorf("report path must be absolute: %q", tr.ReportPath)
	}
}
