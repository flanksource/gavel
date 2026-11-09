package uirexport

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInputsDigestCoversEveryVisibleNonIndexedModuleFile(t *testing.T) {
	repo := inputTestModule(t, "example.com/inputs")
	writeInputTestFile(t, repo, "inputs.go", "package inputs\n")
	commitInputTestFiles(t, repo)

	base := inputTestDigest(t, repo)
	for _, name := range []string{
		"fixtures/want.golden", "notes.txt", "README.md", "page.html", "script.sh",
		"add_" + runtime.GOARCH + ".s", "fixtures/native.c", "fixtures/native.h", "fixtures/object.syso",
		"testdata/ignored.go", "vendor/example.com/dependency/ignored.go",
	} {
		t.Run(name, func(t *testing.T) {
			writeInputTestFile(t, repo, name, "first\n")
			first := inputTestDigest(t, repo)
			if first == base {
				t.Fatalf("adding %s did not change inputs digest", name)
			}
			writeInputTestFile(t, repo, name, "second\n")
			if changed := inputTestDigest(t, repo); changed == first {
				t.Fatalf("editing %s did not change inputs digest", name)
			}
			if err := os.Remove(filepath.Join(repo, filepath.FromSlash(name))); err != nil {
				t.Fatal(err)
			}
		})
	}

	writeInputTestFile(t, repo, "untracked.go", "package inputs\nfunc Untracked() {}\n")
	if got := inputTestDigest(t, repo); got != base {
		t.Fatalf("ordinary Go source belongs to the UIR index: got %s, want %s", got, base)
	}
}

func TestInputsDigestExcludesGeneratedOutputButIncludesIgnoredEmbeds(t *testing.T) {
	repo := inputTestModule(t, "example.com/embedded")
	writeInputTestFile(t, repo, ".gitignore", ".gavel/\nsecret.bin\n")
	writeInputTestFile(t, repo, "embedded.go", "package embedded\n\nimport _ \"embed\"\n\n//go:embed secret.bin\nvar secret []byte\n")
	writeInputTestFile(t, repo, "secret.bin", "first\n")
	commitInputTestFiles(t, repo)

	output := filepath.Join(repo, ".gavel", "run.json")
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		t.Fatal(err)
	}
	writeInputTestFile(t, repo, ".gavel/run.json", "first output\n")
	first, err := InputsDigest(repo, nil, output)
	if err != nil {
		t.Fatal(err)
	}
	writeInputTestFile(t, repo, ".gavel/run.json", "second output\n")
	second, err := InputsDigest(repo, nil, output)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("generated export poisoned digest: %s != %s", first, second)
	}
	writeInputTestFile(t, repo, "secret.bin", "second\n")
	third, err := InputsDigest(repo, nil, output)
	if err != nil {
		t.Fatal(err)
	}
	if third == second {
		t.Fatal("editing ignored embedded input did not change digest")
	}
}

func TestInputsDigestIncludesIgnoredSelectedAssembly(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skip("assembly fixture is amd64")
	}
	repo := inputTestModule(t, "example.com/ignoredassembly")
	writeInputTestFile(t, repo, ".gitignore", "*.s\n")
	writeInputTestFile(t, repo, "assembly.go", "package ignoredassembly\nfunc Value() int\n")
	writeInputTestFile(t, repo, "value_amd64.s", assemblyValue("amd64", 1))
	commitInputTestFiles(t, repo)
	first := inputTestDigest(t, repo)
	writeInputTestFile(t, repo, "value_amd64.s", assemblyValue("amd64", 2))
	if second := inputTestDigest(t, repo); second == first {
		t.Fatal("editing ignored build-selected assembly did not change inputs digest")
	}
}

func TestInputsDigestIncludesIgnoredBuildSelectedGo(t *testing.T) {
	repo := inputTestModule(t, "example.com/ignoredgo")
	writeInputTestFile(t, repo, ".gitignore", "zz_gen.go\nignored_test.go\n")
	writeInputTestFile(t, repo, "main.go", "package ignoredgo\n")
	writeInputTestFile(t, repo, "zz_gen.go", "package ignoredgo\nconst Value = 1\n")
	writeInputTestFile(t, repo, "ignored_test.go", "package ignoredgo\nimport \"testing\"\nfunc TestIgnored(t *testing.T) {}\n")
	commitInputTestFiles(t, repo)

	first := inputTestDigest(t, repo)
	writeInputTestFile(t, repo, "zz_gen.go", "package ignoredgo\nconst Value = 2\n")
	second := inputTestDigest(t, repo)
	if second == first {
		t.Fatal("editing ignored build-selected Go source did not change inputs digest")
	}
	writeInputTestFile(t, repo, "ignored_test.go", "package ignoredgo\nimport \"testing\"\nfunc TestIgnored(t *testing.T) { t.Log(\"changed\") }\n")
	if third := inputTestDigest(t, repo); third == second {
		t.Fatal("editing ignored build-selected Go test source did not change inputs digest")
	}
}

func TestCaptureProvenanceTracksNonGoFilesAndExcludesGeneratedOutput(t *testing.T) {
	repo := inputTestModule(t, "example.com/provenance")
	writeInputTestFile(t, repo, ".gitignore", ".gavel/\n")
	writeInputTestFile(t, repo, "provenance.go", "package provenance\n")
	commitInputTestFiles(t, repo)
	output := filepath.Join(repo, ".gavel", "run.json")
	writeInputTestFile(t, repo, ".gavel/run.json", "first\n")

	clean, err := CaptureProvenance(repo, output)
	if err != nil {
		t.Fatal(err)
	}
	if clean.Dirty || clean.GitTree != "" {
		t.Fatalf("generated output made clean provenance dirty: %+v", clean)
	}
	writeInputTestFile(t, repo, ".gavel/run.json", "second\n")
	stillClean, err := CaptureProvenance(repo, output)
	if err != nil {
		t.Fatal(err)
	}
	if stillClean != clean {
		t.Fatalf("generated output changed provenance: %+v != %+v", stillClean, clean)
	}
	writeInputTestFile(t, repo, "NOTES.txt", "untracked input\n")
	dirty, err := CaptureProvenance(repo, output)
	if err != nil {
		t.Fatal(err)
	}
	if !dirty.Dirty || dirty.GitTree == "" {
		t.Fatalf("untracked non-Go input did not make provenance dirty: %+v", dirty)
	}
}

func TestInputsDigestCoversMutableDependencyCompiledSources(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skip("assembly fixture is amd64")
	}
	parent := t.TempDir()
	dependency := filepath.Join(parent, "dependency")
	app := filepath.Join(parent, "app")
	if err := os.MkdirAll(dependency, 0o755); err != nil {
		t.Fatal(err)
	}
	writeInputTestFile(t, dependency, "go.mod", "module example.com/dependency\n\ngo 1.23\n")
	writeInputTestFile(t, dependency, "dependency.go", "package dependency\nfunc Value() int\n")
	writeInputTestFile(t, dependency, "value_"+runtime.GOARCH+".s", assemblyValue(runtime.GOARCH, 1))
	writeInputTestFile(t, dependency, "fixtures/want.golden", "first\n")
	runInputTestCommand(t, dependency, "git", "init", "-q")
	runInputTestCommand(t, dependency, "git", "config", "user.email", "test@example.com")
	runInputTestCommand(t, dependency, "git", "config", "user.name", "Test")
	commitInputTestFiles(t, dependency)

	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	writeInputTestFile(t, app, "go.mod", "module example.com/app\n\ngo 1.23\n\nrequire example.com/dependency v0.0.0\nreplace example.com/dependency => ../dependency\n")
	writeInputTestFile(t, app, "app.go", "package app\nimport \"example.com/dependency\"\nvar Value = dependency.Value()\n")
	runInputTestCommand(t, app, "git", "init", "-q")
	runInputTestCommand(t, app, "git", "config", "user.email", "test@example.com")
	runInputTestCommand(t, app, "git", "config", "user.name", "Test")
	commitInputTestFiles(t, app)

	first := inputTestDigest(t, app)
	writeInputTestFile(t, dependency, "fixtures/want.golden", "second\n")
	if second := inputTestDigest(t, app); second == first {
		t.Fatal("editing mutable dependency data did not change inputs digest")
	}
	writeInputTestFile(t, dependency, "fixtures/want.golden", "first\n")
	writeInputTestFile(t, dependency, "value_"+runtime.GOARCH+".s", assemblyValue(runtime.GOARCH, 2))
	if second := inputTestDigest(t, app); second == first {
		t.Fatal("editing mutable dependency assembly did not change inputs digest")
	}
}

func TestInputsDigestCoversMutableDependencyCgoSources(t *testing.T) {
	if output, err := exec.Command("go", "env", "CGO_ENABLED").Output(); err != nil || string(output) != "1\n" {
		t.Skip("cgo is unavailable")
	}
	parent := t.TempDir()
	dependency := filepath.Join(parent, "dependency")
	app := filepath.Join(parent, "app")
	if err := os.MkdirAll(dependency, 0o755); err != nil {
		t.Fatal(err)
	}
	writeInputTestFile(t, dependency, "go.mod", "module example.com/cdependency\n\ngo 1.23\n")
	writeInputTestFile(t, dependency, "dependency.go", "package dependency\n/* #include \"native.h\" */\nimport \"C\"\nfunc Value() int { return int(C.native_value()) }\n")
	writeInputTestFile(t, dependency, "native.h", "int native_value(void);\n")
	writeInputTestFile(t, dependency, "native.c", "int native_value(void) { return 1; }\n")
	runInputTestCommand(t, dependency, "git", "init", "-q")
	runInputTestCommand(t, dependency, "git", "config", "user.email", "test@example.com")
	runInputTestCommand(t, dependency, "git", "config", "user.name", "Test")
	commitInputTestFiles(t, dependency)

	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	writeInputTestFile(t, app, "go.mod", "module example.com/capp\n\ngo 1.23\n\nrequire example.com/cdependency v0.0.0\nreplace example.com/cdependency => ../dependency\n")
	writeInputTestFile(t, app, "app.go", "package app\nimport \"example.com/cdependency\"\nvar Value = dependency.Value()\n")
	runInputTestCommand(t, app, "git", "init", "-q")
	runInputTestCommand(t, app, "git", "config", "user.email", "test@example.com")
	runInputTestCommand(t, app, "git", "config", "user.name", "Test")
	commitInputTestFiles(t, app)

	first := inputTestDigest(t, app)
	writeInputTestFile(t, dependency, "native.c", "int native_value(void) { return 2; }\n")
	second := inputTestDigest(t, app)
	if second == first {
		t.Fatal("editing mutable dependency C source did not change inputs digest")
	}
	writeInputTestFile(t, dependency, "native.h", "int native_value(void); /* changed */\n")
	if third := inputTestDigest(t, app); third == second {
		t.Fatal("editing mutable dependency header did not change inputs digest")
	}
}

func TestInputsDigestRejectsMutableDependencyOutsideGit(t *testing.T) {
	parent := t.TempDir()
	dependency := filepath.Join(parent, "dependency")
	app := filepath.Join(parent, "app")
	writeInputTestFile(t, dependency, "go.mod", "module example.com/untrackeddependency\n\ngo 1.23\n")
	writeInputTestFile(t, dependency, "dependency.go", "package dependency\nconst Value = 1\n")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	writeInputTestFile(t, app, "go.mod", "module example.com/untrackedapp\n\ngo 1.23\n\nrequire example.com/untrackeddependency v0.0.0\nreplace example.com/untrackeddependency => ../dependency\n")
	writeInputTestFile(t, app, "app.go", "package app\nimport \"example.com/untrackeddependency\"\nvar Value = dependency.Value\n")
	runInputTestCommand(t, app, "git", "init", "-q")
	runInputTestCommand(t, app, "git", "config", "user.email", "test@example.com")
	runInputTestCommand(t, app, "git", "config", "user.name", "Test")
	commitInputTestFiles(t, app)

	_, err := InputsDigest(app, nil, "")
	if err == nil || !strings.Contains(err.Error(), "Git checkout of mutable dependency") {
		t.Fatalf("InputsDigest error = %v", err)
	}
}

func inputTestModule(t *testing.T, module string) string {
	t.Helper()
	repo := t.TempDir()
	writeInputTestFile(t, repo, "go.mod", "module "+module+"\n\ngo 1.23\n")
	runInputTestCommand(t, repo, "git", "init", "-q")
	runInputTestCommand(t, repo, "git", "config", "user.email", "test@example.com")
	runInputTestCommand(t, repo, "git", "config", "user.name", "Test")
	return repo
}

func commitInputTestFiles(t *testing.T, repo string) {
	t.Helper()
	runInputTestCommand(t, repo, "git", "add", ".")
	runInputTestCommand(t, repo, "git", "commit", "-qm", "fixture")
}

func inputTestDigest(t *testing.T, repo string) string {
	t.Helper()
	digest, err := InputsDigest(repo, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	return digest
}

func writeInputTestFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runInputTestCommand(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, output)
	}
}

func assemblyValue(architecture string, value int) string {
	if architecture != "amd64" {
		return fmt.Sprintf("// assembly fixture %d\n", value)
	}
	return fmt.Sprintf("#include \"textflag.h\"\nTEXT ·Value(SB), NOSPLIT, $0-8\n\tMOVQ $%d, ret+0(FP)\n\tRET\n", value)
}
