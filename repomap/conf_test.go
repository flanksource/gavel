package repomap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/models"
)

// Test GetFileMap with basic file
func TestGetFileMap_BasicFile(t *testing.T) {
	// Create a temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(testFile, []byte("package main\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	conf := &ArchConf{}
	fm, err := conf.GetFileMap(testFile, "HEAD")

	if err != nil {
		t.Fatalf("GetFileMap failed: %v", err)
	}
	// GetFileMap stores the full path as-is
	if fm.Path != testFile {
		t.Errorf("Expected path '%s', got '%s'", testFile, fm.Path)
	}
	if fm.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", fm.Language)
	}
}

// Test GetFileMap with non-existent file
func TestGetFileMap_NonExistentFile(t *testing.T) {
	// GetFileMap doesn't error on non-existent files - it just returns a FileMap
	// with basic path/language info. The actual error would come from ReadFile
	// if trying to read content from a non-existent file.
	conf := &ArchConf{}
	fm, err := conf.GetFileMap("/nonexistent/file.go", "HEAD")

	if err != nil {
		t.Errorf("GetFileMap should not error, got: %v", err)
	}
	if fm.Path != "/nonexistent/file.go" {
		t.Errorf("Expected path to be set, got '%s'", fm.Path)
	}
	if fm.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", fm.Language)
	}
}

// Test GetFileMap with scope rules
func TestGetFileMap_WithScopeRules(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "handler_test.go")
	err := os.WriteFile(testFile, []byte("package main\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	conf := &ArchConf{
		Scopes: models.ScopesConfig{
			Rules: map[string][]models.PathRule{
				"test": {
					{Path: "*_test.go"},
				},
			},
		},
	}

	fm, err := conf.GetFileMap(testFile, "HEAD")
	if err != nil {
		t.Fatalf("GetFileMap failed: %v", err)
	}

	if !fm.Scopes.Contains(models.ScopeTypeTest) {
		t.Errorf("Expected scope 'test', got '%s'", fm.Scopes)
	}
}

// Test GetFileMap with technology detection
func TestGetFileMap_TechnologyDetection(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(testFile, []byte("package main\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Empty ArchConf has no tech rules, so Tech will be empty
	// To get Tech populated, we need to configure tech rules
	conf := &ArchConf{
		Tech: models.TechnologyConfig{
			Rules: models.PathRules{
				"go": {{Path: "*.go"}},
			},
		},
	}
	fm, err := conf.GetFileMap(testFile, "HEAD")

	if err != nil {
		t.Fatalf("GetFileMap failed: %v", err)
	}

	// Tech field should be populated from configured rules
	if len(fm.Tech) == 0 {
		t.Error("Expected Tech field to be populated, got empty")
	}
}

// Test GetFileMap without including history
func TestGetFileMap_NoHistory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(testFile, []byte("package main\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	conf := &ArchConf{}
	fm, err := conf.GetFileMap(testFile, "HEAD")

	if err != nil {
		t.Fatalf("GetFileMap failed: %v", err)
	}

	// Commits should be empty when includeHistory=false
	if len(fm.Commits) != 0 {
		t.Errorf("Expected no commits when includeHistory=false, got %d", len(fm.Commits))
	}
}

// Test GetFileMap with including history
func TestGetFileMap_WithHistory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(testFile, []byte("package main\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	conf := &ArchConf{}
	fm, err := conf.GetFileMap(testFile, "HEAD")

	if err != nil {
		t.Fatalf("GetFileMap failed: %v", err)
	}

	// Note: Commits will still be empty since git integration is not implemented (FIXME)
	// This test documents the expected behavior when implemented
	_ = fm.Commits
}

// Test GetFileMap with different file types
func TestGetFileMap_DifferentFileTypes(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		filename string
		expected string
	}{
		{"test.py", "python"},
		{"test.js", "javascript"},
		{"test.ts", "typescript"},
		{"test.md", "markdown"},
		{"test.go", "go"},
	}

	conf := &ArchConf{}

	for _, tc := range testCases {
		testFile := filepath.Join(tmpDir, tc.filename)
		err := os.WriteFile(testFile, []byte("test content\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tc.filename, err)
		}

		fm, err := conf.GetFileMap(testFile, "HEAD")
		if err != nil {
			t.Fatalf("GetFileMap failed for %s: %v", tc.filename, err)
		}

		if fm.Language != tc.expected {
			t.Errorf("For file %s, expected language '%s', got '%s'",
				tc.filename, tc.expected, fm.Language)
		}
	}
}

// Test LoadArchConf with valid config
func TestLoadArchConf_ValidConfig(t *testing.T) {
	// Save current directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Create a temp directory with arch.yaml
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	validConfig := `
scopes:
  rules:
    test:
      - path: "*_test.go"
    docs:
      - path: "*.md"
`
	err := os.WriteFile("arch.yaml", []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write arch.yaml: %v", err)
	}

	conf, err := LoadArchConf("arch.yaml")
	if err != nil {
		t.Fatalf("LoadArchConf failed: %v", err)
	}

	if conf == nil {
		t.Error("Expected non-nil config")
	}
	if len(conf.Scopes.Rules) == 0 {
		t.Error("Expected rules to be loaded")
	}
}

// Test LoadArchConf with scope not in allowed_scopes
func TestLoadArchConf_ScopeNotInAllowedScopes(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	invalidConfig := `
scopes:
  allowedScopes:
    - test
    - docs
  rules:
    invalid_scope:
      - path: "*.go"
`
	err := os.WriteFile("arch.yaml", []byte(invalidConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write arch.yaml: %v", err)
	}

	conf, err := LoadArchConf("arch.yaml")
	if err == nil {
		t.Errorf("Expected validation error for scope not in allowed_scopes, got nil. AllowedScopes: %v, Rules: %v", conf.Scopes.AllowedScopes, conf.Scopes.Rules)
	}
}

// Test LoadArchConf with no config file
func TestLoadArchConf_NoConfigFile(t *testing.T) {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// LoadArchConf returns error if file doesn't exist
	// Use GetConfForFile if you want nil instead of error
	_, err := LoadArchConf("arch.yaml")
	if err == nil {
		t.Error("Expected error when arch.yaml doesn't exist")
	}
}

// Test loadDefaultArchConf loads embedded defaults
func TestLoadDefaultArchConf(t *testing.T) {
	conf, err := loadDefaultArchConf()
	if err != nil {
		t.Fatalf("loadDefaultArchConf failed: %v", err)
	}
	if conf == nil {
		t.Fatal("Expected non-nil default config")
	}

	// Verify scopes are loaded
	if len(conf.Scopes.Rules) == 0 {
		t.Error("Expected default scope rules to be loaded")
	}

	// Verify tech is loaded
	if len(conf.Tech.Rules) == 0 {
		t.Error("Expected default tech rules to be loaded")
	}

	// Verify specific default rules exist
	if _, ok := conf.Scopes.Rules["test"]; !ok {
		t.Error("Expected 'test' scope in default rules")
	}
	if _, ok := conf.Tech.Rules["go"]; !ok {
		t.Error("Expected 'go' tech in default rules")
	}
}

// Test mergeArchConf with user rules taking precedence
func TestMergeArchConf_UserRulesPrecedence(t *testing.T) {
	userConf := &ArchConf{
		Scopes: models.ScopesConfig{
			Rules: models.PathRules{
				"test": {{Path: "**/*_test.go"}},
			},
		},
		Tech: models.TechnologyConfig{
			Rules: models.PathRules{
				"go": {{Path: "**/*.go"}},
			},
		},
	}

	defaultConf := &ArchConf{
		Scopes: models.ScopesConfig{
			Rules: models.PathRules{
				"test": {{Path: "**/test/**"}},
				"docs": {{Path: "**/*.md"}},
			},
		},
		Tech: models.TechnologyConfig{
			Rules: models.PathRules{
				"go":     {{Path: "**/go.mod"}},
				"python": {{Path: "**/*.py"}},
			},
		},
	}

	merged := defaultConf.Merge(userConf)

	// User rules should come first
	if len(merged.Scopes.Rules["test"]) < 1 {
		t.Error("Expected test scope rules to be present")
	}
	if merged.Scopes.Rules["test"][0].Path != "**/*_test.go" {
		t.Errorf("Expected user rule first, got %s", merged.Scopes.Rules["test"][0].Path)
	}

	// Default rules should be appended
	if _, ok := merged.Scopes.Rules["docs"]; !ok {
		t.Error("Expected docs scope from defaults")
	}
	if _, ok := merged.Tech.Rules["python"]; !ok {
		t.Error("Expected python tech from defaults")
	}
}

// Test mergeArchConf with nil user config
func TestMergeArchConf_NilUserConf(t *testing.T) {
	defaultConf := &ArchConf{
		Scopes: models.ScopesConfig{
			Rules: models.PathRules{
				"test": {{Path: "**/test/**"}},
			},
		},
	}

	merged := defaultConf.Merge(nil)

	if len(merged.Scopes.Rules) == 0 {
		t.Error("Expected default rules when user config is nil")
	}
}

// Test mergeArchConf with nil default config
func TestMergeArchConf_NilDefaultConf(t *testing.T) {
	userConf := &ArchConf{
		Scopes: models.ScopesConfig{
			Rules: models.PathRules{
				"test": {{Path: "**/*_test.go"}},
			},
		},
	}

	merged := userConf.Merge(nil)

	if len(merged.Scopes.Rules) == 0 {
		t.Error("Expected user rules when default config is nil")
	}
}

// Test GetFileMap uses defaults when no user config
func TestGetFileMap_UsesDefaults(t *testing.T) {
	// Create a temp git repo so GetFileMap can find git root
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Test Go file detection
	goFile := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(goFile, []byte("package main\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create go file: %v", err)
	}

	fm, err := GetFileMap(goFile, "HEAD")
	if err != nil {
		t.Fatalf("GetFileMap failed: %v", err)
	}

	// Should detect golang from defaults
	hasGolang := false
	for _, tech := range fm.Tech {
		if tech == models.Go {
			hasGolang = true
			break
		}
	}
	if !hasGolang {
		t.Errorf("Expected golang tech to be detected from defaults, got %v", fm.Tech)
	}

	// Test Python file detection
	pyFile := filepath.Join(tmpDir, "script.py")
	err = os.WriteFile(pyFile, []byte("print('hello')\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create python file: %v", err)
	}

	fm, err = GetFileMap(pyFile, "HEAD")
	if err != nil {
		t.Fatalf("GetFileMap failed: %v", err)
	}

	// Should detect python from defaults
	hasPython := false
	for _, tech := range fm.Tech {
		if tech == models.Python {
			hasPython = true
			break
		}
	}
	if !hasPython {
		t.Errorf("Expected python tech to be detected from defaults, got %v", fm.Tech)
	}

	// Test Dockerfile detection
	dockerfile := filepath.Join(tmpDir, "Dockerfile")
	err = os.WriteFile(dockerfile, []byte("FROM alpine\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create Dockerfile: %v", err)
	}

	fm, err = GetFileMap(dockerfile, "HEAD")
	if err != nil {
		t.Fatalf("GetFileMap failed: %v", err)
	}

	// Should detect docker tech from defaults (no deployment scope defined in defaults)
	hasDocker := false
	for _, tech := range fm.Tech {
		if tech == models.Docker {
			hasDocker = true
			break
		}
	}
	if !hasDocker {
		t.Errorf("Expected docker tech to be detected from defaults, got %v", fm.Tech)
	}
}

// Test findGitRoot
func TestFindGitRoot(t *testing.T) {
	// Create a temp directory structure with a git repo
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	subDir := filepath.Join(tmpDir, "sub", "dir")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Test finding git root from subdirectory
	root := FindGitRoot(subDir)
	if root != tmpDir {
		t.Errorf("Expected git root %s, got %s", tmpDir, root)
	}

	// Test finding git root from root itself
	root = FindGitRoot(tmpDir)
	if root != tmpDir {
		t.Errorf("Expected git root %s, got %s", tmpDir, root)
	}

	// Test with file path
	testFile := filepath.Join(subDir, "test.go")
	err = os.WriteFile(testFile, []byte("package main\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	root = FindGitRoot(testFile)
	if root != tmpDir {
		t.Errorf("Expected git root %s from file, got %s", tmpDir, root)
	}

	// Test with non-git directory
	nonGitDir := t.TempDir()
	root = FindGitRoot(nonGitDir)
	if root != "" {
		t.Errorf("Expected empty string for non-git directory, got %s", root)
	}
}

// Test GetConf
func TestGetConf(t *testing.T) {
	// Create a temp git repo
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.go")
	err = os.WriteFile(testFile, []byte("package main\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	conf, err := GetConf(testFile)
	if err != nil {
		t.Fatalf("GetConf failed: %v", err)
	}

	if conf.repoPath != tmpDir {
		t.Errorf("Expected repoPath %s, got %s", tmpDir, conf.repoPath)
	}

	if conf.RepoPath() != tmpDir {
		t.Errorf("Expected RepoPath() %s, got %s", tmpDir, conf.RepoPath())
	}
}

// Test ReadFile requires commit
func TestReadFile_RequiresCommit(t *testing.T) {
	// Create a temp git repo
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	testFile := "test.txt"
	testPath := filepath.Join(tmpDir, testFile)
	testContent := "Hello, World!"
	err = os.WriteFile(testPath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	conf, err := GetConf(testPath)
	if err != nil {
		t.Fatalf("GetConf failed: %v", err)
	}

	// ReadFile requires a non-empty commit
	_, err = conf.ReadFile(testFile, "")
	if err == nil {
		t.Error("Expected error when commit is empty")
	}

	// ReadFile with HEAD commit requires actual git repo with commits
	// In a temp dir without real git init, this will fail
	_, err = conf.ReadFile(testFile, "HEAD")
	if err == nil {
		t.Error("Expected error when reading from non-initialized git repo")
	}
}

// Test Exec wrapper
func TestExec(t *testing.T) {
	// This is a real git repository, so we can test Exec
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	conf := &ArchConf{repoPath: tmpDir}
	git := conf.Exec()

	// Test simple git command
	result, err := git("--version")
	if err != nil {
		t.Fatalf("git --version failed: %v", err)
	}
	if result.Stdout == "" {
		t.Error("Expected git --version to return output")
	}
}
