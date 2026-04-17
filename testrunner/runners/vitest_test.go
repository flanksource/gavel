package runners

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestVitestDetect(t *testing.T) {
	t.Run("vitest.config", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x"}`)
		if err := os.WriteFile(filepath.Join(tmp, "vitest.config.ts"), []byte("export default {};"), 0o644); err != nil {
			t.Fatal(err)
		}
		r := NewVitest(tmp)
		if got, _ := r.Detect(tmp); !got {
			t.Fatal("expected detect=true")
		}
	})
	t.Run("vitest in devDependencies", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x","devDependencies":{"vitest":"^1.0.0"}}`)
		r := NewVitest(tmp)
		if got, _ := r.Detect(tmp); !got {
			t.Fatal("expected detect=true")
		}
	})
	t.Run("bare vite.config without vitest dep", func(t *testing.T) {
		tmp := t.TempDir()
		writePackageJSON(t, tmp, `{"name":"x"}`)
		if err := os.WriteFile(filepath.Join(tmp, "vite.config.ts"), []byte("export default {};"), 0o644); err != nil {
			t.Fatal(err)
		}
		r := NewVitest(tmp)
		if got, _ := r.Detect(tmp); got {
			t.Fatal("expected detect=false for bare vite config without vitest dep")
		}
	})
}

func TestVitestDiscoverPackages_Monorepo(t *testing.T) {
	tmp := t.TempDir()
	writePackageJSON(t, tmp, `{"name":"root"}`)
	writePackageJSON(t, filepath.Join(tmp, "apps/web"), `{"name":"web","devDependencies":{"vitest":"^1.0.0"}}`)
	writePackageJSON(t, filepath.Join(tmp, "apps/api"), `{"name":"api"}`)
	if err := os.WriteFile(filepath.Join(tmp, "apps/api", "vitest.config.ts"), []byte("export default {};"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := NewVitest(tmp)
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

func TestVitestBuildCommand(t *testing.T) {
	tmp := t.TempDir()
	writePackageJSON(t, tmp, `{"name":"x","devDependencies":{"vitest":"^1.0.0"}}`)
	r := NewVitest(tmp)
	tr, err := r.BuildCommand(".")
	if err != nil {
		t.Fatal(err)
	}
	if tr.Framework != parsers.Vitest {
		t.Errorf("framework = %q", tr.Framework)
	}
	joined := strings.Join(tr.Process.Args, " ")
	if !strings.Contains(joined, "vitest run --reporter=json --outputFile=") {
		t.Errorf("args missing vitest run invocation: %v", tr.Process.Args)
	}
}
