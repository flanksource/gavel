package reactdoctor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/utils"
)

const (
	linterName       = "react-doctor"
	packageSpec      = "react-doctor@latest"
	packageJSONFile  = "package.json"
	npmShrinkwrap    = "npm-shrinkwrap.json"
	packageManagerPN = "pnpm"
	packageManagerNM = "npm"
)

// ReactDoctor implements the React Doctor JS/TS/React health scanner.
type ReactDoctor struct {
	linters.RunOptions
	fileCount int
	ruleCount int
}

func NewReactDoctor(workDir string) *ReactDoctor {
	return &ReactDoctor{RunOptions: linters.RunOptions{WorkDir: workDir}}
}

func (r *ReactDoctor) SetOptions(opts linters.RunOptions) { r.RunOptions = opts }

func (r *ReactDoctor) Name() string { return linterName }

func (r *ReactDoctor) ProjectRootMarkers() []string {
	return append([]string{packageJSONFile}, reactDoctorConfigFiles()...)
}

func (r *ReactDoctor) DefaultIncludes() []string {
	return []string{
		packageJSONFile,
		"**/*.js",
		"**/*.jsx",
		"**/*.ts",
		"**/*.tsx",
		"**/*.mjs",
		"**/*.cjs",
		"**/*.mts",
		"**/*.cts",
	}
}

func (r *ReactDoctor) DefaultExcludes() []string {
	return []string{
		"node_modules/**",
		"bower_components/**",
		"jspm_packages/**",
		"dist/**",
		"build/**",
		".next/**",
		"coverage/**",
		".cache/**",
	}
}

func (r *ReactDoctor) GetSupportedLanguages() []string {
	return []string{"javascript", "typescript"}
}

func (r *ReactDoctor) GetEffectiveExcludes(language string, config *models.Config) []string {
	if config == nil {
		return r.DefaultExcludes()
	}
	return config.GetAllLanguageExcludes(language, r.DefaultExcludes())
}

func (r *ReactDoctor) GetEffectiveIncludes(language string, config *models.Config) []string {
	if config == nil {
		return r.DefaultIncludes()
	}
	return config.GetAllLanguageIncludes(language, r.DefaultIncludes())
}

func (r *ReactDoctor) SupportsJSON() bool { return true }

func (r *ReactDoctor) JSONArgs() []string { return []string{"--json"} }

func (r *ReactDoctor) SupportsFix() bool { return false }

func (r *ReactDoctor) FixArgs() []string { return nil }

func (r *ReactDoctor) ValidateConfig(config *models.LinterConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	return nil
}

// HasDirectConfig reports explicit React Doctor configuration, including
// package.json#reactDoctor.
func (r *ReactDoctor) HasDirectConfig(workDir string) bool {
	for _, name := range reactDoctorConfigFiles() {
		if _, err := os.Stat(filepath.Join(workDir, name)); err == nil {
			return true
		}
	}
	return packageJSONHasReactDoctorConfig(workDir)
}

// HasDefaultActivation selects react-doctor for React projects even when no
// React Doctor config has been added yet.
func (r *ReactDoctor) HasDefaultActivation(workDir string) bool {
	return packageJSONHasReactDependency(workDir)
}

// ExecutableCandidates returns the preferred command for the package manager
// implied by the nearest lockfile, followed by the installed binary fallback.
func (r *ReactDoctor) ExecutableCandidates(workDir string) []string {
	switch detectPackageManager(workDir) {
	case packageManagerPN:
		return []string{"pnpx", linterName}
	case packageManagerNM:
		return []string{"npx", linterName}
	default:
		return []string{linterName}
	}
}

func (r *ReactDoctor) GetFileCount() int { return r.fileCount }

func (r *ReactDoctor) GetRuleCount() int { return r.ruleCount }

func (r *ReactDoctor) DryRunCommand() (string, []string) {
	cmd := r.commandName()
	return cmd, r.buildArgs()
}

func (r *ReactDoctor) Run(ctx commonsContext.Context, _ *clicky.Task) ([]models.Violation, error) {
	cmdName := r.commandName()
	args := r.buildArgs()
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = r.WorkDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = r.WrapWriter(&stdout)
	cmd.Stderr = r.WrapWriter(&stderr)

	logger.Infof("Executing: %s %s", cmdName, strings.Join(args, " "))

	runErr := cmd.Run()
	if stdout.Len() > 0 {
		violations, parseErr := r.parseViolations(stdout.Bytes())
		if parseErr != nil {
			return nil, parseErr
		}
		return violations, nil
	}
	if runErr != nil {
		return nil, fmt.Errorf("react-doctor execution failed: %w\nOutput:\n%s", runErr, stderr.String())
	}
	return []models.Violation{}, nil
}

func (r *ReactDoctor) commandName() string {
	if r.Executable != "" {
		return r.Executable
	}
	candidates := r.ExecutableCandidates(r.WorkDir)
	if len(candidates) > 0 {
		return candidates[0]
	}
	return linterName
}

func (r *ReactDoctor) commandKind() string {
	base := filepath.Base(r.commandName())
	base = strings.TrimSuffix(base, filepath.Ext(base))
	switch base {
	case "pnpx", "npx":
		return base
	default:
		return linterName
	}
}

func (r *ReactDoctor) buildArgs() []string {
	var args []string
	if kind := r.commandKind(); kind == "pnpx" || kind == "npx" {
		args = append(args, packageSpec)
	}
	if r.Config != nil {
		args = append(args, r.Config.Args...)
	}
	if !hasPathArg(args) {
		args = append(args, ".")
	}
	if r.ForceJSON && !hasFlag(args, "--json") {
		args = append(args, "--json")
	}
	if r.ForceJSON && !hasFlag(args, "--json-compact") {
		args = append(args, "--json-compact")
	}
	if !hasFlag(args, "--no-telemetry") && !hasFlag(args, "--no-score") {
		args = append(args, "--no-telemetry")
	}
	if !hasFlag(args, "--no-color") && !hasFlag(args, "--color") {
		args = append(args, "--no-color")
	}
	if !hasFlag(args, "--blocking") && !hasFlag(args, "--fail-on") {
		args = append(args, "--blocking", "none")
	}
	args = append(args, r.ExtraArgs...)
	return args
}

func (r *ReactDoctor) parseViolations(output []byte) ([]models.Violation, error) {
	trimmed := bytes.TrimSpace(output)
	if len(trimmed) == 0 {
		return []models.Violation{}, nil
	}

	var report reactDoctorReport
	if err := json.Unmarshal(trimmed, &report); err != nil {
		logger.Debugf("Failed to parse react-doctor JSON output: %v\nOutput: %s", err, string(output))
		return nil, fmt.Errorf("failed to parse react-doctor JSON output: %w", err)
	}
	if report.Error != nil && len(report.Diagnostics) == 0 && len(report.Projects) == 0 {
		return nil, fmt.Errorf("react-doctor: %s", report.Error.message())
	}

	diagnostics := append([]reactDoctorDiagnostic{}, report.Diagnostics...)
	for _, project := range report.Projects {
		diagnostics = append(diagnostics, project.Diagnostics...)
	}

	violations := make([]models.Violation, 0, len(diagnostics))
	files := make(map[string]struct{})
	rules := make(map[string]struct{})
	for _, diagnostic := range diagnostics {
		violation := diagnostic.toViolation(r.WorkDir)
		violations = append(violations, violation)
		if violation.File != "" {
			files[violation.File] = struct{}{}
		}
		if violation.Rule != nil && violation.Rule.Method != "" {
			rules[violation.Rule.Method] = struct{}{}
		}
	}

	r.fileCount = report.Summary.AffectedFileCount
	if r.fileCount == 0 {
		r.fileCount = len(files)
	}
	r.ruleCount = len(rules)

	return violations, nil
}

type reactDoctorReport struct {
	SchemaVersion int                     `json:"schemaVersion"`
	OK            bool                    `json:"ok"`
	Diagnostics   []reactDoctorDiagnostic `json:"diagnostics"`
	Projects      []reactDoctorProject    `json:"projects"`
	Summary       struct {
		AffectedFileCount    int `json:"affectedFileCount"`
		TotalDiagnosticCount int `json:"totalDiagnosticCount"`
	} `json:"summary"`
	Error *reactDoctorError `json:"error"`
}

type reactDoctorProject struct {
	Diagnostics []reactDoctorDiagnostic `json:"diagnostics"`
}

type reactDoctorError struct {
	Message string   `json:"message"`
	Name    string   `json:"name"`
	Chain   []string `json:"chain"`
}

func (e *reactDoctorError) message() string {
	if e == nil {
		return "unknown error"
	}
	if e.Message != "" {
		return e.Message
	}
	if len(e.Chain) > 0 {
		return strings.Join(e.Chain, ": ")
	}
	if e.Name != "" {
		return e.Name
	}
	return "unknown error"
}

type reactDoctorDiagnostic struct {
	FilePath string `json:"filePath"`
	Plugin   string `json:"plugin"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Help     string `json:"help"`
	URL      string `json:"url"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Category string `json:"category"`
}

func (d *reactDoctorDiagnostic) toViolation(workDir string) models.Violation {
	filename := d.FilePath
	if filename != "" && !filepath.IsAbs(filename) {
		filename = filepath.Join(workDir, filename)
	}
	rule := reactDoctorRuleName(d.Plugin, d.Rule)

	violation := models.NewViolationBuilder().
		WithFile(filename).
		WithLocation(d.Line, d.Column).
		WithCaller(filepath.Dir(filename), "unknown").
		WithCalled(linterName, rule).
		WithMessage(d.violationMessage()).
		WithSource(linterName).
		WithRuleFromLinter(linterName, rule).
		Build()
	violation.Severity = reactDoctorSeverity(d.Severity)
	return violation
}

func (d *reactDoctorDiagnostic) violationMessage() string {
	message := strings.TrimSpace(d.Message)
	title := strings.TrimSpace(d.Title)
	if message == "" {
		message = title
	} else if title != "" && !strings.Contains(message, title) {
		message = title + ": " + message
	}
	if help := strings.TrimSpace(d.Help); help != "" && !strings.Contains(message, help) {
		message += " " + help
	}
	if message == "" {
		message = "React Doctor diagnostic"
	}
	return message
}

func reactDoctorRuleName(plugin, rule string) string {
	plugin = strings.TrimSpace(plugin)
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return "unknown"
	}
	if strings.Contains(rule, "/") || plugin == "" {
		return rule
	}
	return plugin + "/" + rule
}

func reactDoctorSeverity(severity string) models.ViolationSeverity {
	switch strings.ToLower(severity) {
	case "error":
		return models.SeverityError
	case "warning", "warn":
		return models.SeverityWarning
	default:
		return models.SeverityInfo
	}
}

func reactDoctorConfigFiles() []string {
	return []string{
		"doctor.config.ts",
		"doctor.config.js",
		"doctor.config.mjs",
		"doctor.config.cjs",
		"doctor.config.json",
		"react-doctor.config.json",
	}
}

func packageJSONHasReactDoctorConfig(workDir string) bool {
	var pkg struct {
		ReactDoctor json.RawMessage `json:"reactDoctor"`
	}
	if !readPackageJSON(workDir, &pkg) {
		return false
	}
	return len(bytes.TrimSpace(pkg.ReactDoctor)) > 0 && string(bytes.TrimSpace(pkg.ReactDoctor)) != "null"
}

func packageJSONHasReactDependency(workDir string) bool {
	var pkg struct {
		Dependencies         map[string]json.RawMessage `json:"dependencies"`
		DevDependencies      map[string]json.RawMessage `json:"devDependencies"`
		PeerDependencies     map[string]json.RawMessage `json:"peerDependencies"`
		OptionalDependencies map[string]json.RawMessage `json:"optionalDependencies"`
	}
	if !readPackageJSON(workDir, &pkg) {
		return false
	}

	for _, deps := range []map[string]json.RawMessage{
		pkg.Dependencies,
		pkg.DevDependencies,
		pkg.PeerDependencies,
		pkg.OptionalDependencies,
	} {
		for name := range deps {
			if isReactPackage(name) {
				return true
			}
		}
	}
	return false
}

func readPackageJSON(workDir string, target any) bool {
	path := filepath.Join(workDir, packageJSONFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	if err := json.Unmarshal(data, target); err != nil {
		logger.V(2).Infof("react-doctor: parse %s: %v", path, err)
		return false
	}
	return true
}

func isReactPackage(name string) bool {
	switch name {
	case "react",
		"react-dom",
		"react-native",
		"next",
		"gatsby",
		"expo",
		"@remix-run/react",
		"@vitejs/plugin-react",
		"@vitejs/plugin-react-swc":
		return true
	default:
		return false
	}
}

func detectPackageManager(workDir string) string {
	cur, err := filepath.Abs(workDir)
	if err != nil {
		cur = workDir
	}
	gitRoot := utils.FindGitRoot(cur)
	for {
		if _, err := os.Stat(filepath.Join(cur, "pnpm-lock.yaml")); err == nil {
			return packageManagerPN
		}
		if _, err := os.Stat(filepath.Join(cur, packageJSONFile)); err == nil {
			if _, err := os.Stat(filepath.Join(cur, "package-lock.json")); err == nil {
				return packageManagerNM
			}
			if _, err := os.Stat(filepath.Join(cur, npmShrinkwrap)); err == nil {
				return packageManagerNM
			}
		} else {
			if _, err := os.Stat(filepath.Join(cur, "package-lock.json")); err == nil {
				return packageManagerNM
			}
			if _, err := os.Stat(filepath.Join(cur, npmShrinkwrap)); err == nil {
				return packageManagerNM
			}
		}
		if gitRoot != "" && cur == gitRoot {
			return ""
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return ""
		}
		cur = parent
	}
}

func hasPathArg(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if arg == packageSpec {
			continue
		}
		return true
	}
	return false
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag || strings.HasPrefix(arg, flag+"=") {
			return true
		}
	}
	return false
}
