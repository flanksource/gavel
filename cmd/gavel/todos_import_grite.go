package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/griteexport"
	"github.com/flanksource/gavel/todos/migrategrite"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

type importGriteOptions struct {
	Input       string
	ProbeInput  string
	Workspace   string
	ArtifactDir string
	Finalize    string
	PlanRoot    string
	Frozen      bool
	Strict      bool
}

var todosImportGriteOptions importGriteOptions

var todosImportGriteCmd = &cobra.Command{
	Use:          "import-grite",
	Short:        "One-off migration of Grite issues into native PostgreSQL TODOs",
	SilenceUsage: true,
	Hidden:       true,
	Args:         cobra.NoArgs,
	RunE:         runTodosImportGrite,
}

type importGriteOutput struct {
	Mode        string                     `json:"mode"`
	ArtifactDir string                     `json:"artifactDir"`
	Report      *migrategrite.ImportReport `json:"report"`
}

type loadedGriteSource struct {
	Snapshot griteexport.Snapshot
	Raw      []byte
	Live     bool
	Result   griteexport.Result
	CursorMS int64
}

func init() {
	todosCmd.AddCommand(todosImportGriteCmd)
	todosImportGriteCmd.Flags().StringVar(&todosImportGriteOptions.Input, "input", "", "Read a Grite export JSON file instead of invoking grite")
	todosImportGriteCmd.Flags().StringVar(&todosImportGriteOptions.ProbeInput, "probe-input", "", "Second overlapping export proving an offline final delta is frozen")
	todosImportGriteCmd.Flags().StringVar(&todosImportGriteOptions.Workspace, "workspace", "", "Workspace UUID or repository key (required for file input; an absent key is created explicitly)")
	todosImportGriteCmd.Flags().StringVar(&todosImportGriteOptions.ArtifactDir, "artifact-dir", "", "Parent root for private immutable migration generations")
	todosImportGriteCmd.Flags().StringVar(&todosImportGriteOptions.Finalize, "finalize", "", "Initial artifact generation to finalize with a frozen delta")
	todosImportGriteCmd.Flags().StringVar(&todosImportGriteOptions.PlanRoot, "plan-root", "", "Root used to resolve relative legacy plan paths (default: workspace path)")
	todosImportGriteCmd.Flags().BoolVar(&todosImportGriteOptions.Frozen, "frozen", false, "Confirm Grite writers are stopped for final delta validation")
	todosImportGriteCmd.Flags().BoolVar(&todosImportGriteOptions.Strict, "strict", false, "Abort the transaction when any migration warning is produced")
}

func runTodosImportGrite(command *cobra.Command, _ []string) error {
	if err := validateImportGriteOptions(todosImportGriteOptions); err != nil {
		return err
	}
	ctx := command.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return err
	}
	db, err := database.Require(ctx, "gavel todos import-grite")
	if err != nil {
		return err
	}
	if strings.TrimSpace(todosImportGriteOptions.Finalize) != "" {
		return finalizeGriteImport(ctx, command, db, workDir, todosImportGriteOptions)
	}
	return initialGriteImport(ctx, command, db, workDir, todosImportGriteOptions)
}

func validateImportGriteOptions(options importGriteOptions) error {
	finalize := strings.TrimSpace(options.Finalize) != ""
	fileInput := strings.TrimSpace(options.Input) != ""
	probeInput := strings.TrimSpace(options.ProbeInput) != ""
	if finalize && !options.Frozen {
		return errors.New("--finalize requires --frozen after stopping all Grite writers")
	}
	if !finalize && options.Frozen {
		return errors.New("--frozen is only valid with --finalize")
	}
	if finalize && strings.TrimSpace(options.ArtifactDir) != "" {
		return errors.New("--artifact-dir cannot be combined with --finalize; final output is a sibling immutable generation")
	}
	if finalize && strings.TrimSpace(options.Workspace) != "" {
		return errors.New("--workspace cannot be combined with --finalize; workspace identity comes from the initial manifest")
	}
	if !finalize && fileInput && strings.TrimSpace(options.Workspace) == "" {
		return errors.New("initial --input requires an explicit --workspace binding")
	}
	if !finalize && probeInput {
		return errors.New("--probe-input is only valid with offline --finalize")
	}
	if finalize && fileInput && !probeInput {
		return errors.New("offline --finalize with --input also requires --probe-input")
	}
	if finalize && !fileInput && probeInput {
		return errors.New("--probe-input cannot be used when final exports are taken live")
	}
	return nil
}

func initialGriteImport(ctx context.Context, command *cobra.Command, db *gorm.DB, workDir string, options importGriteOptions) error {
	source, err := loadGriteSource(ctx, workDir, options.Input, 0)
	if err != nil {
		return err
	}
	if err := migrategrite.ValidateFullSnapshotHistory(source.Snapshot); err != nil {
		return err
	}
	document, err := migrategrite.Normalize(source.Snapshot)
	if err != nil {
		return err
	}
	workspace, err := resolveImportWorkspace(ctx, db, workDir, options.Workspace)
	if err != nil {
		return err
	}
	workspace = native.NormalizeImportWorkspace(workspace)
	if workspace.ID == uuid.Nil {
		workspace.ID = native.DeterministicImportWorkspaceID(workspace)
	}
	if workspace.ID == uuid.Nil {
		return errors.New("resolve stable native workspace ID for Grite artifacts")
	}
	artifactRoot := strings.TrimSpace(options.ArtifactDir)
	if artifactRoot == "" {
		artifactRoot, err = migrategrite.DefaultArtifactRoot()
		if err != nil {
			return err
		}
	} else {
		artifactRoot = absoluteOptionalPath(workDir, artifactRoot)
	}
	artifactDir := migrategrite.ArtifactDirectory(
		artifactRoot,
		workspaceArtifactKey(workspace.ID),
		migrategrite.ArtifactRunID(source.Snapshot, document.SourceHash),
	)
	planRoot := absoluteOptionalPath(workDir, options.PlanRoot)
	baselineDir, baselineExists, err := resolveWorkspaceInitialArtifact(filepath.Dir(artifactDir), document.SourceHash)
	if err != nil {
		return err
	}
	if baselineExists {
		artifactDir = baselineDir
		report, exists, err := existingInitialArtifact(artifactDir, source, document, workspace, workDir)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("committed initial Grite baseline disappeared during validation: %s", artifactDir)
		}
		if err := validateExistingArtifactReplay(ctx, db, source.Snapshot, report, migrategrite.ImportOptions{
			Workspace: workspace, PlanRoot: planRoot, Strict: options.Strict, DeferActivePromptRuns: true,
		}); err != nil {
			return err
		}
		return writeImportGriteOutput(command, importGriteOutput{Mode: "existing", ArtifactDir: artifactDir, Report: report})
	}
	if err := refuseExistingGriteImportWithoutBaseline(ctx, db, workspace.ID); err != nil {
		return err
	}

	pending, err := migrategrite.PrepareArtifactGeneration(artifactDir)
	if err != nil {
		return err
	}
	keepPending := false
	defer func() {
		if !keepPending {
			_ = migrategrite.DiscardArtifactGeneration(pending)
		}
	}()
	if err := migrategrite.WritePreparedSource(pending, "source-initial.json", source.Raw); err != nil {
		return err
	}
	provenance := sourceProvenance(source, workDir, workspace, false, "")
	service, err := migrategrite.NewService(db)
	if err != nil {
		return err
	}
	report, err := service.Import(ctx, source.Snapshot, migrategrite.ImportOptions{
		Workspace:             workspace,
		PlanRoot:              planRoot,
		Strict:                options.Strict,
		DeferActivePromptRuns: true,
		RequireNoPriorImport:  true,
		BeforeCommit: func(report *migrategrite.ImportReport) error {
			return migrategrite.WriteInitialArtifacts(pending, source.Raw, source.Snapshot, report, provenance)
		},
	})
	if err != nil {
		return err
	}
	// PostgreSQL is committed. Preserve the complete fsynced pending generation
	// if publication itself fails so rollback evidence is never discarded.
	keepPending = true
	if err := migrategrite.PublishArtifactGeneration(pending, artifactDir); err != nil {
		return fmt.Errorf("database committed; publish the durable pending artifact %s to %s: %w", pending, artifactDir, err)
	}
	keepPending = false
	return writeImportGriteOutput(command, importGriteOutput{Mode: "initial", ArtifactDir: artifactDir, Report: report})
}

func finalizeGriteImport(ctx context.Context, command *cobra.Command, db *gorm.DB, workDir string, options importGriteOptions) error {
	initialDir := absoluteOptionalPath(workDir, options.Finalize)
	if err := migrategrite.VerifyArtifactChecksums(initialDir); err != nil {
		return err
	}
	manifest, err := migrategrite.LoadArtifactManifest(initialDir)
	if err != nil {
		return err
	}
	if manifest.State != "initial" {
		return fmt.Errorf("--finalize requires an initial artifact generation, got %q", manifest.State)
	}
	initial, initialRaw, err := migrategrite.LoadInitialSnapshot(initialDir)
	if err != nil {
		return err
	}
	initialReport, err := migrategrite.LoadImportReport(initialDir, "initial")
	if err != nil {
		return err
	}
	if initialReport.Validation.SourceHash != manifest.InitialSourceHash ||
		initialReport.Validation.ImportFingerprint != manifest.InitialImportFingerprint {
		return errors.New("initial Grite artifact manifest and report fingerprints disagree")
	}
	initialRollback, err := migrategrite.LoadRollback(initialDir)
	if err != nil {
		return err
	}
	sourceRoot := manifest.SourceRoot
	if sourceRoot == "" {
		return errors.New("initial Grite artifact has no source root provenance")
	}
	cursor := manifest.Watermark.SinceMS()
	delta, err := loadGriteSource(ctx, sourceRoot, options.Input, cursor)
	if err != nil {
		return err
	}
	merged, err := migrategrite.MergeSnapshots(initial, delta.Snapshot)
	if err != nil {
		return err
	}
	var probe loadedGriteSource
	if delta.Live {
		probe, err = loadGriteSource(ctx, sourceRoot, "", griteexport.WatermarkFor(merged.Events).SinceMS())
	} else {
		probe, err = loadGriteSource(ctx, workDir, options.ProbeInput, griteexport.WatermarkFor(merged.Events).SinceMS())
	}
	if err != nil {
		return fmt.Errorf("load frozen Grite validation probe: %w", err)
	}
	if !delta.Live {
		if err := validateOfflineGriteProbe(workDir, options.Input, options.ProbeInput, delta, probe); err != nil {
			return err
		}
	} else if err := migrategrite.ValidateFrozenWALHeads(delta.Result.WALHead, probe.Result.WALHead); err != nil {
		return err
	}
	if err := migrategrite.ValidateFrozenProbe(merged, probe.Snapshot); err != nil {
		return err
	}
	if err := migrategrite.ValidateFullSnapshotHistory(merged); err != nil {
		return err
	}
	finalDocument, err := migrategrite.Normalize(merged)
	if err != nil {
		return err
	}
	finalDir := initialDir + "-final-" + shortHash(finalDocument.SourceHash)
	workspace, err := workspaceByID(ctx, db, manifest.WorkspaceID)
	if err != nil {
		return err
	}
	if manifest.RepoKey != "" && workspace.RepoKey != manifest.RepoKey {
		return fmt.Errorf("artifact repository %q does not match workspace repository %q", manifest.RepoKey, workspace.RepoKey)
	}
	planRoot := absoluteOptionalPath(sourceRoot, options.PlanRoot)
	if report, exists, err := existingFinalArtifact(finalDir, delta, probe, finalDocument, manifest.WorkspaceID); err != nil {
		return err
	} else if exists {
		if err := validateExistingArtifactReplay(ctx, db, merged, report, migrategrite.ImportOptions{
			Workspace: workspace, PlanRoot: planRoot, Strict: options.Strict,
		}); err != nil {
			return err
		}
		return writeImportGriteOutput(command, importGriteOutput{Mode: "final-existing", ArtifactDir: finalDir, Report: report})
	}
	pending, err := migrategrite.PrepareArtifactGeneration(finalDir)
	if err != nil {
		return err
	}
	keepPending := false
	defer func() {
		if !keepPending {
			_ = migrategrite.DiscardArtifactGeneration(pending)
		}
	}()
	if err := migrategrite.WritePreparedSource(pending, "source-initial.json", initialRaw); err != nil {
		return err
	}
	if err := migrategrite.WritePreparedSource(pending, "source-final-delta.json", delta.Raw); err != nil {
		return err
	}
	if err := migrategrite.WritePreparedSource(pending, "source-final-probe.json", probe.Raw); err != nil {
		return err
	}
	provenance := sourceProvenance(delta, sourceRoot, workspace, true, migrategrite.SHA256Bytes(probe.Raw))
	provenance.ProbeWALHead = append(json.RawMessage(nil), probe.Result.WALHead...)
	service, err := migrategrite.NewService(db)
	if err != nil {
		return err
	}
	report, err := service.Import(ctx, merged, migrategrite.ImportOptions{
		Workspace:              workspace,
		PlanRoot:               planRoot,
		Strict:                 options.Strict,
		ExpectedTargetChecksum: initialReport.Validation.TargetChecksum,
		BeforeCommit: func(report *migrategrite.ImportReport) error {
			rollback := migrategrite.MergeRollback(initialRollback, report.Rollback)
			return migrategrite.WriteFinalArtifacts(pending, initialDir, delta.Raw, delta.Snapshot, report, rollback, provenance)
		},
	})
	if err != nil {
		return err
	}
	keepPending = true
	if err := migrategrite.PublishArtifactGeneration(pending, finalDir); err != nil {
		return fmt.Errorf("database committed; publish the durable pending final artifact %s to %s: %w", pending, finalDir, err)
	}
	keepPending = false
	return writeImportGriteOutput(command, importGriteOutput{Mode: "final", ArtifactDir: finalDir, Report: report})
}

func loadGriteSource(ctx context.Context, workDir, input string, sinceMS int64) (loadedGriteSource, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		snapshot, raw, result, err := (migrategrite.LiveExporter{WorkDir: workDir}).Export(ctx, sinceMS)
		if err != nil {
			return loadedGriteSource{}, err
		}
		return loadedGriteSource{Snapshot: snapshot, Raw: raw, Live: true, Result: result, CursorMS: sinceMS}, nil
	}
	path := absoluteOptionalPath(workDir, input)
	raw, err := os.ReadFile(path)
	if err != nil {
		return loadedGriteSource{}, fmt.Errorf("read Grite import input %s: %w", path, err)
	}
	snapshot, err := griteexport.DecodeFile(raw)
	if err != nil {
		return loadedGriteSource{}, err
	}
	if snapshot.Meta.SchemaVersion != 1 {
		return loadedGriteSource{}, fmt.Errorf("unsupported Grite export schema version %d", snapshot.Meta.SchemaVersion)
	}
	if snapshot.Meta.EventCount != len(snapshot.Events) {
		return loadedGriteSource{}, fmt.Errorf("Grite export metadata reports %d events, file contains %d", snapshot.Meta.EventCount, len(snapshot.Events))
	}
	return loadedGriteSource{Snapshot: snapshot, Raw: raw, CursorMS: sinceMS}, nil
}

func existingInitialArtifact(
	dir string,
	source loadedGriteSource,
	document migrategrite.Document,
	workspace native.ImportWorkspace,
	sourceRoot string,
) (*migrategrite.ImportReport, bool, error) {
	if _, err := os.Lstat(dir); errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	if err := migrategrite.VerifyArtifactChecksums(dir); err != nil {
		return nil, false, err
	}
	manifest, err := migrategrite.LoadArtifactManifest(dir)
	if err != nil {
		return nil, false, err
	}
	if manifest.State != "initial" || manifest.WorkspaceID != workspace.ID || workspace.ID == uuid.Nil ||
		manifest.SourceMode != griteSourceMode(source) ||
		manifest.InitialSourceHash != document.SourceHash {
		return nil, false, fmt.Errorf("artifact generation %s already exists for different source content/state", dir)
	}
	if manifest.SourceRoot != filepath.Clean(sourceRoot) || (manifest.RepoKey != "" && manifest.RepoKey != workspace.RepoKey) {
		return nil, false, errors.New("existing Grite artifact source/workspace binding does not match this invocation")
	}
	report, err := migrategrite.LoadImportReport(dir, "initial")
	return report, err == nil, err
}

func resolveWorkspaceInitialArtifact(workspaceDir, sourceHash string) (string, bool, error) {
	info, err := os.Lstat(workspaceDir)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", false, fmt.Errorf("Grite workspace artifact path is not a real directory: %s", workspaceDir)
	}
	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		return "", false, err
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") && strings.Contains(entry.Name(), ".pending-") {
			pending := filepath.Join(workspaceDir, entry.Name())
			return "", false, fmt.Errorf(
				"%w: %s; publish or explicitly discard it before reusing the workspace baseline",
				migrategrite.ErrArtifactPending,
				pending,
			)
		}
	}
	var initialDirs, finalDirs []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		dir := filepath.Join(workspaceDir, entry.Name())
		entryInfo, err := os.Lstat(dir)
		if err != nil {
			return "", false, err
		}
		if entryInfo.Mode()&os.ModeSymlink != 0 {
			return "", false, fmt.Errorf("Grite artifact generation is not a real directory: %s", dir)
		}
		if !entryInfo.IsDir() {
			continue
		}
		manifest, err := migrategrite.LoadArtifactManifest(dir)
		if errors.Is(err, os.ErrNotExist) {
			return "", false, fmt.Errorf("unrecognized directory in Grite workspace artifacts: %s", dir)
		}
		if err != nil {
			return "", false, err
		}
		if err := migrategrite.VerifyArtifactChecksums(dir); err != nil {
			return "", false, fmt.Errorf("verify committed Grite artifact %s: %w", dir, err)
		}
		switch manifest.State {
		case "initial":
			initialDirs = append(initialDirs, dir)
		case "final":
			finalDirs = append(finalDirs, dir)
		default:
			return "", false, fmt.Errorf("unsupported committed Grite artifact state %q in %s", manifest.State, dir)
		}
	}
	slices.Sort(initialDirs)
	slices.Sort(finalDirs)
	if len(finalDirs) > 0 {
		return "", false, fmt.Errorf("workspace migration is already finalized; refuse another initial baseline: %s", strings.Join(finalDirs, ", "))
	}
	if len(initialDirs) > 1 {
		return "", false, fmt.Errorf("workspace has multiple committed initial Grite baselines; refuse another import: %s", strings.Join(initialDirs, ", "))
	}
	if len(initialDirs) == 0 {
		return "", false, nil
	}
	baseline := initialDirs[0]
	manifest, err := migrategrite.LoadArtifactManifest(baseline)
	if err != nil {
		return "", false, err
	}
	if manifest.InitialSourceHash != sourceHash {
		return "", false, fmt.Errorf(
			"workspace already has committed initial Grite baseline %s for a different source; refuse a second rollback baseline and run --finalize %s",
			baseline,
			baseline,
		)
	}
	return baseline, true, nil
}

func existingFinalArtifact(
	dir string,
	delta loadedGriteSource,
	probe loadedGriteSource,
	document migrategrite.Document,
	workspaceID uuid.UUID,
) (*migrategrite.ImportReport, bool, error) {
	if _, err := os.Lstat(dir); errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	if err := migrategrite.VerifyArtifactChecksums(dir); err != nil {
		return nil, false, err
	}
	manifest, err := migrategrite.LoadArtifactManifest(dir)
	if err != nil {
		return nil, false, err
	}
	mode := griteSourceMode(delta)
	if manifest.State != "final" || manifest.WorkspaceID != workspaceID || manifest.FinalSourceMode != mode ||
		manifest.FinalSourceHash != document.SourceHash {
		return nil, false, fmt.Errorf("final artifact generation %s already exists for different source content/state", dir)
	}
	if mode == "live" {
		if !probe.Live {
			return nil, false, errors.New("existing live final artifact requires a live frozen probe")
		}
		if err := migrategrite.ValidateFrozenWALHeads(manifest.FinalWALHead, delta.Result.WALHead); err != nil {
			return nil, false, fmt.Errorf("existing final artifact delta WAL binding does not match this invocation: %w", err)
		}
		if err := migrategrite.ValidateFrozenWALHeads(manifest.FinalProbeWALHead, probe.Result.WALHead); err != nil {
			return nil, false, fmt.Errorf("existing final artifact probe WAL binding does not match this invocation: %w", err)
		}
	} else if manifest.FinalRawHash != migrategrite.SHA256Bytes(delta.Raw) ||
		manifest.FinalProbeRawHash != migrategrite.SHA256Bytes(probe.Raw) {
		return nil, false, fmt.Errorf("final artifact generation %s already exists for different source content/state", dir)
	}
	report, err := migrategrite.LoadImportReport(dir, "final")
	return report, err == nil, err
}

var errExistingArtifactReplayValidated = errors.New("existing Grite artifact replay validated; roll back read-only check")

func validateExistingArtifactReplay(
	ctx context.Context,
	db *gorm.DB,
	snapshot griteexport.Snapshot,
	expected *migrategrite.ImportReport,
	options migrategrite.ImportOptions,
) error {
	if expected == nil {
		return errors.New("existing Grite artifact report is nil")
	}
	if strings.TrimSpace(expected.Validation.TargetChecksum) == "" ||
		strings.TrimSpace(expected.Validation.CaptainChecksum) == "" ||
		strings.TrimSpace(expected.Validation.ImportFingerprint) == "" {
		return errors.New("existing Grite artifact report is missing replay validation checksums")
	}
	service, err := migrategrite.NewService(db)
	if err != nil {
		return err
	}
	validated := false
	options.ExpectedTargetChecksum = expected.Validation.TargetChecksum
	options.BeforeCommit = func(actual *migrategrite.ImportReport) error {
		if err := compareExistingArtifactReplay(expected, actual); err != nil {
			return err
		}
		validated = true
		return errExistingArtifactReplayValidated
	}
	_, err = service.Import(ctx, snapshot, options)
	if validated && errors.Is(err, errExistingArtifactReplayValidated) {
		return nil
	}
	if err == nil {
		return errors.New("existing Grite artifact replay validation committed unexpectedly")
	}
	return fmt.Errorf("validate existing Grite artifact against current database: %w", err)
}

func compareExistingArtifactReplay(expected, actual *migrategrite.ImportReport) error {
	if expected == nil || actual == nil {
		return errors.New("existing Grite artifact replay report is nil")
	}
	counts := actual.Counts
	if counts.WorkspaceCreated != 0 || counts.WorkspaceUpdated != 0 ||
		counts.IssuesCreated != 0 || counts.IssuesUpdated != 0 || counts.AliasesInserted != 0 ||
		counts.EventsInserted != 0 || counts.ProjectionEventsInserted != 0 || counts.WarningsInserted != 0 ||
		counts.RelationshipsInserted != 0 || counts.RelationshipsDeleted != 0 ||
		counts.PromptRunLinksInserted != 0 || counts.PlanLinksInserted != 0 ||
		len(actual.Rollback.CreatedCaptainPlanIDs) != 0 || len(actual.Rollback.AppendedCaptainRevisionIDs) != 0 {
		return errors.New("existing Grite artifact replay would mutate the database")
	}
	if actual.WorkspaceID != expected.WorkspaceID {
		return fmt.Errorf("existing Grite artifact workspace changed: expected %s, got %s", expected.WorkspaceID, actual.WorkspaceID)
	}
	if actual.Watermark.TimestampMS != expected.Watermark.TimestampMS ||
		!slices.Equal(actual.Watermark.EventIDs, expected.Watermark.EventIDs) {
		return errors.New("existing Grite artifact watermark changed")
	}
	checks := []struct {
		name     string
		expected string
		actual   string
	}{
		{name: "source hash", expected: expected.Validation.SourceHash, actual: actual.Validation.SourceHash},
		{name: "import fingerprint", expected: expected.Validation.ImportFingerprint, actual: actual.Validation.ImportFingerprint},
		{name: "target checksum", expected: expected.Validation.TargetChecksum, actual: actual.Validation.TargetChecksum},
		{name: "Captain checksum", expected: expected.Validation.CaptainChecksum, actual: actual.Validation.CaptainChecksum},
	}
	for _, check := range checks {
		if check.actual != check.expected {
			return fmt.Errorf("existing Grite artifact %s changed: expected %s, got %s", check.name, check.expected, check.actual)
		}
	}
	return nil
}

func griteSourceMode(source loadedGriteSource) string {
	if source.Live {
		return "live"
	}
	return "file"
}

func sourceProvenance(source loadedGriteSource, root string, workspace native.ImportWorkspace, frozen bool, probeHash string) migrategrite.ArtifactProvenance {
	mode := "file"
	if source.Live {
		mode = "live"
	}
	return migrategrite.ArtifactProvenance{
		SourceMode: mode, SourceRoot: filepath.Clean(root), RepoKey: workspace.RepoKey,
		CursorMS: source.CursorMS, WALHead: source.Result.WALHead,
		FrozenValidated: frozen, ProbeRawHash: probeHash,
		BinaryVersion: version, BinaryCommit: commit,
	}
}

func validateOfflineGriteProbe(
	workDir, deltaInput, probeInput string,
	delta, probe loadedGriteSource,
) error {
	deltaPath := absoluteOptionalPath(workDir, deltaInput)
	probePath := absoluteOptionalPath(workDir, probeInput)
	if deltaPath == probePath {
		return errors.New("--probe-input must be a separately captured export, not the final delta file")
	}
	if migrategrite.SHA256Bytes(delta.Raw) == migrategrite.SHA256Bytes(probe.Raw) {
		return errors.New("--probe-input is byte-identical to the final delta; capture a second export after freezing Grite")
	}
	if delta.Snapshot.Meta.GeneratedTS <= 0 || probe.Snapshot.Meta.GeneratedTS <= delta.Snapshot.Meta.GeneratedTS {
		return fmt.Errorf(
			"offline frozen probe must have a generated_ts later than the final delta (%d <= %d)",
			probe.Snapshot.Meta.GeneratedTS,
			delta.Snapshot.Meta.GeneratedTS,
		)
	}
	return nil
}

func resolveImportWorkspace(ctx context.Context, db *gorm.DB, workDir, identity string) (native.ImportWorkspace, error) {
	repository, err := native.NewRepository(db)
	if err != nil {
		return native.ImportWorkspace{}, err
	}
	identity = strings.TrimSpace(identity)
	if identity != "" {
		if parsed, err := uuid.Parse(identity); err == nil {
			workspace, err := repository.GetWorkspace(ctx, parsed)
			if err != nil {
				return native.ImportWorkspace{}, err
			}
			return nativeWorkspace(*workspace), nil
		} else if looksLikeUUID(identity) {
			return native.ImportWorkspace{}, fmt.Errorf("invalid workspace UUID %q", identity)
		}
		normalizedIdentity := native.NormalizeImportWorkspace(native.ImportWorkspace{RepoKey: identity}).RepoKey
		workspace, err := repository.GetWorkspaceByRepoKey(ctx, normalizedIdentity)
		if err == nil {
			return nativeWorkspace(*workspace), nil
		}
		if !errors.Is(err, native.ErrNotFound) {
			return native.ImportWorkspace{}, err
		}
		return native.NormalizeImportWorkspace(native.ImportWorkspace{
			RepoKey: normalizedIdentity, RootPath: workDir, DisplayName: filepath.Base(workDir),
		}), nil
	}
	if repo, err := github.ResolveRepoFromDir(workDir); err == nil && strings.TrimSpace(repo) != "" {
		repoKey := "github.com/" + strings.ToLower(strings.TrimSpace(repo))
		if workspace, lookupErr := repository.GetWorkspaceByRepoKey(ctx, repoKey); lookupErr == nil {
			return nativeWorkspace(*workspace), nil
		} else if !errors.Is(lookupErr, native.ErrNotFound) {
			return native.ImportWorkspace{}, lookupErr
		}
		return native.ImportWorkspace{RepoKey: repoKey, RootPath: workDir, DisplayName: filepath.Base(workDir)}, nil
	}
	if workspace, err := repository.GetWorkspaceByPath(ctx, workDir); err == nil {
		return nativeWorkspace(*workspace), nil
	} else if !errors.Is(err, native.ErrNotFound) {
		return native.ImportWorkspace{}, err
	}
	return native.ImportWorkspace{RootPath: workDir, DisplayName: filepath.Base(workDir)}, nil
}

func workspaceByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (native.ImportWorkspace, error) {
	repository, err := native.NewRepository(db)
	if err != nil {
		return native.ImportWorkspace{}, err
	}
	workspace, err := repository.GetWorkspace(ctx, id)
	if err != nil {
		return native.ImportWorkspace{}, err
	}
	return nativeWorkspace(*workspace), nil
}

func nativeWorkspace(workspace native.Workspace) native.ImportWorkspace {
	return native.ImportWorkspace{
		ID: workspace.ID, RepoKey: workspace.RepoKey, RootPath: workspace.RootPath,
		DisplayName: workspace.DisplayName, CreatedAt: workspace.CreatedAt, UpdatedAt: workspace.UpdatedAt,
	}
}

func workspaceArtifactKey(id uuid.UUID) string {
	if id != uuid.Nil {
		return id.String()
	}
	return "workspace"
}

func refuseExistingGriteImportWithoutBaseline(ctx context.Context, db *gorm.DB, workspaceID uuid.UUID) error {
	if workspaceID == uuid.Nil {
		return errors.New("stable workspace ID is required for Grite import evidence preflight")
	}
	var exists bool
	if err := db.WithContext(ctx).Raw(`
		SELECT EXISTS(
			SELECT 1
			FROM todo_issue_events AS event
			JOIN todo_issues AS issue ON issue.id = event.issue_id
			WHERE issue.workspace_id = ? AND event.source = ?
			UNION ALL
			SELECT 1
			FROM todo_issue_aliases AS alias
			WHERE alias.workspace_id = ? AND alias.kind = 'grite'
		)`, workspaceID, native.DefaultImportSource, workspaceID).Scan(&exists).Error; err != nil {
		return fmt.Errorf("check existing Grite import evidence for workspace %s: %w", workspaceID, err)
	}
	if exists {
		return fmt.Errorf(
			"workspace %s already contains %s source events or Grite aliases but the selected artifact root has no baseline; refuse a second rollback baseline and recover the original migration artifacts",
			workspaceID,
			native.DefaultImportSource,
		)
	}
	return nil
}

func looksLikeUUID(value string) bool {
	compact := strings.ReplaceAll(strings.TrimSpace(value), "-", "")
	if len(compact) != 32 {
		return false
	}
	_, err := hex.DecodeString(compact)
	return err == nil
}

func shortHash(hash string) string {
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}

func writeImportGriteOutput(command *cobra.Command, output importGriteOutput) error {
	encoder := json.NewEncoder(command.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func absoluteOptionalPath(workDir, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(workDir, path)
}
