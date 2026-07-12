package migrategrite

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/gavel/todos/griteexport"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
)

const artifactSchemaVersion = 1

var (
	ErrArtifactCommitted = errors.New("Grite migration artifact generation is already committed")
	ErrArtifactPending   = errors.New("Grite migration artifact generation has a preserved pending sibling")
)

type ArtifactProvenance struct {
	SourceMode      string          `json:"sourceMode"`
	SourceRoot      string          `json:"sourceRoot"`
	RepoKey         string          `json:"repoKey,omitempty"`
	CursorMS        int64           `json:"cursorMs"`
	WALHead         json.RawMessage `json:"walHead,omitempty"`
	ProbeWALHead    json.RawMessage `json:"probeWalHead,omitempty"`
	ProbeRawHash    string          `json:"probeRawHash,omitempty"`
	FrozenValidated bool            `json:"frozenValidated"`
	BinaryVersion   string          `json:"binaryVersion,omitempty"`
	BinaryCommit    string          `json:"binaryCommit,omitempty"`
}

type ArtifactManifest struct {
	SchemaVersion            int                   `json:"schemaVersion"`
	State                    string                `json:"state"`
	Frozen                   bool                  `json:"frozen"`
	WorkspaceID              uuid.UUID             `json:"workspaceId"`
	SourceMode               string                `json:"sourceMode"`
	SourceRoot               string                `json:"sourceRoot"`
	RepoKey                  string                `json:"repoKey,omitempty"`
	InitialCursorMS          int64                 `json:"initialCursorMs"`
	FinalCursorMS            int64                 `json:"finalCursorMs,omitempty"`
	InitialRawHash           string                `json:"initialRawHash"`
	FinalRawHash             string                `json:"finalRawHash,omitempty"`
	FinalProbeRawHash        string                `json:"finalProbeRawHash,omitempty"`
	FinalSourceMode          string                `json:"finalSourceMode,omitempty"`
	InitialSourceHash        string                `json:"initialSourceHash"`
	FinalSourceHash          string                `json:"finalSourceHash,omitempty"`
	InitialImportFingerprint string                `json:"initialImportFingerprint"`
	FinalImportFingerprint   string                `json:"finalImportFingerprint,omitempty"`
	Watermark                griteexport.Watermark `json:"watermark"`
	GeneratedAt              time.Time             `json:"generatedAt"`
	WALHead                  json.RawMessage       `json:"walHead,omitempty"`
	FinalWALHead             json.RawMessage       `json:"finalWalHead,omitempty"`
	FinalProbeWALHead        json.RawMessage       `json:"finalProbeWalHead,omitempty"`
	BinaryVersion            string                `json:"binaryVersion,omitempty"`
	BinaryCommit             string                `json:"binaryCommit,omitempty"`
	Warnings                 []Warning             `json:"warnings,omitempty"`
}

// ArtifactRunID is stable for one semantic initial source. The workspace is
// already the parent directory, so the full normalized source hash prevents a
// later export timestamp from establishing a second rollback baseline.
func ArtifactRunID(_ griteexport.Snapshot, sourceHash string) string {
	hash := strings.TrimSpace(sourceHash)
	if hash == "" {
		hash = "unknown"
	}
	return "initial-" + safeArtifactComponent(hash)
}

func SHA256Bytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func DefaultArtifactRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory for Grite artifacts: %w", err)
	}
	return filepath.Join(home, ".gavel", "migrations", "grite"), nil
}

func ArtifactDirectory(root, workspaceKey, runID string) string {
	workspaceKey = safeArtifactComponent(workspaceKey)
	if workspaceKey == "" {
		workspaceKey = "workspace"
	}
	return filepath.Join(root, workspaceKey, safeArtifactComponent(runID))
}

// PrepareArtifactGeneration creates a private sibling directory. Callers write
// and fsync the complete artifact there before committing PostgreSQL, then use
// PublishArtifactGeneration after commit. The target must not already exist.
func PrepareArtifactGeneration(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", errors.New("Grite artifact target is empty")
	}
	target, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if _, err := os.Lstat(target); err == nil {
		return "", fmt.Errorf("%w: %s", ErrArtifactCommitted, target)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	parent := filepath.Dir(target)
	if err := ensureArtifactParent(parent); err != nil {
		return "", err
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") && strings.Contains(entry.Name(), ".pending-") {
			pending := filepath.Join(parent, entry.Name())
			return "", pendingArtifactError(pending, target)
		}
	}
	// A deterministic workspace reservation makes the check-and-create atomic:
	// concurrent first imports cannot both enter their database transactions.
	pending := filepath.Join(parent, ".workspace.pending-generation")
	if err := os.Mkdir(pending, 0o700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return "", pendingArtifactError(pending, target)
		}
		return "", fmt.Errorf("reserve pending Grite artifact generation: %w", err)
	}
	if err := os.Chmod(pending, 0o700); err != nil {
		_ = os.RemoveAll(pending)
		return "", err
	}
	if err := syncDirectory(pending); err != nil {
		_ = os.Remove(pending)
		_ = syncDirectory(parent)
		return "", fmt.Errorf("sync pending Grite artifact reservation: %w", err)
	}
	if err := syncDirectory(parent); err != nil {
		_ = os.Remove(pending)
		_ = syncDirectory(parent)
		return "", fmt.Errorf("persist pending Grite artifact reservation: %w", err)
	}
	return pending, nil
}

func pendingArtifactError(pending, target string) error {
	return fmt.Errorf(
		"%w: %s; publish or explicitly discard it before retrying %s",
		ErrArtifactPending,
		pending,
		target,
	)
}

func PublishArtifactGeneration(pending, target string) error {
	if err := VerifyArtifactChecksums(pending); err != nil {
		return fmt.Errorf("verify pending Grite artifact generation: %w", err)
	}
	if _, err := os.Lstat(target); err == nil {
		return fmt.Errorf("%w: %s", ErrArtifactCommitted, target)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := syncDirectory(pending); err != nil {
		return err
	}
	if err := os.Rename(pending, target); err != nil {
		return fmt.Errorf("publish Grite artifact generation: %w", err)
	}
	return syncDirectory(filepath.Dir(target))
}

func DiscardArtifactGeneration(pending string) error {
	if strings.TrimSpace(pending) == "" {
		return nil
	}
	parent := filepath.Dir(pending)
	if err := os.RemoveAll(pending); err != nil {
		return err
	}
	if err := syncDirectory(parent); err != nil {
		return fmt.Errorf("persist discarded Grite artifact generation: %w", err)
	}
	return nil
}

func WritePreparedSource(dir, name string, source []byte) error {
	if name != "source-initial.json" && name != "source-final-delta.json" && name != "source-final-probe.json" {
		return fmt.Errorf("unsupported prepared source filename %q", name)
	}
	if err := ensureArtifactDir(dir); err != nil {
		return err
	}
	if err := writeImmutableArtifact(filepath.Join(dir, name), source); err != nil {
		return err
	}
	return syncDirectory(dir)
}

func WriteInitialArtifacts(
	dir string,
	source []byte,
	snapshot griteexport.Snapshot,
	report *ImportReport,
	provenance ...ArtifactProvenance,
) error {
	if report == nil {
		return errors.New("initial Grite artifact report is nil")
	}
	if err := ensureUncommittedGeneration(dir, []string{"source-initial.json"}); err != nil {
		return err
	}
	var sourceInfo ArtifactProvenance
	if len(provenance) > 0 {
		sourceInfo = provenance[0]
	}
	if err := validateArtifactProvenance(sourceInfo, false); err != nil {
		return err
	}
	if err := writeImmutableArtifact(filepath.Join(dir, "source-initial.json"), source); err != nil {
		return err
	}
	if err := writeJSONArtifact(filepath.Join(dir, "validation-initial.json"), report.Validation); err != nil {
		return err
	}
	if err := writeJSONArtifact(filepath.Join(dir, "report-initial.json"), report); err != nil {
		return err
	}
	if err := writeJSONArtifact(filepath.Join(dir, "rollback.json"), report.Rollback); err != nil {
		return err
	}
	if err := writeRollbackMarkdown(filepath.Join(dir, "rollback.md"), report); err != nil {
		return err
	}
	manifest := ArtifactManifest{
		SchemaVersion:            artifactSchemaVersion,
		State:                    "initial",
		WorkspaceID:              report.WorkspaceID,
		SourceMode:               sourceInfo.SourceMode,
		SourceRoot:               filepath.Clean(sourceInfo.SourceRoot),
		RepoKey:                  sourceInfo.RepoKey,
		InitialCursorMS:          sourceInfo.CursorMS,
		InitialRawHash:           SHA256Bytes(source),
		InitialSourceHash:        report.Validation.SourceHash,
		InitialImportFingerprint: report.Validation.ImportFingerprint,
		Watermark:                report.Watermark,
		GeneratedAt:              time.UnixMilli(snapshot.Meta.GeneratedTS).UTC(),
		WALHead:                  cloneRaw(sourceInfo.WALHead),
		BinaryVersion:            sourceInfo.BinaryVersion,
		BinaryCommit:             sourceInfo.BinaryCommit,
		Warnings:                 report.Warnings,
	}
	if sourceInfo.SourceRoot == "" {
		manifest.SourceRoot = ""
	}
	if err := writeJSONArtifact(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		return err
	}
	if err := writeArtifactChecksums(dir); err != nil {
		return err
	}
	return VerifyArtifactChecksums(dir)
}

// WriteFinalArtifacts builds a complete immutable final generation in dir. It
// copies the verified initial evidence instead of modifying the initial run.
func WriteFinalArtifacts(
	dir, initialDir string,
	source []byte,
	snapshot griteexport.Snapshot,
	report *ImportReport,
	rollback Rollback,
	provenance ...ArtifactProvenance,
) error {
	if report == nil {
		return errors.New("final Grite artifact report is nil")
	}
	sourceInfo := sourceInfoFor(provenance)
	if err := validateArtifactProvenance(sourceInfo, true); err != nil {
		return err
	}
	if err := VerifyArtifactChecksums(initialDir); err != nil {
		return fmt.Errorf("verify initial Grite artifacts: %w", err)
	}
	initialManifest, err := LoadArtifactManifest(initialDir)
	if err != nil {
		return err
	}
	if initialManifest.State != "initial" {
		return fmt.Errorf("finalization requires an initial artifact, got state %q", initialManifest.State)
	}
	if initialManifest.WorkspaceID != report.WorkspaceID {
		return fmt.Errorf("final artifact workspace %s does not match initial workspace %s", report.WorkspaceID, initialManifest.WorkspaceID)
	}
	if err := ensureUncommittedGeneration(dir, []string{"source-initial.json", "source-final-delta.json", "source-final-probe.json"}); err != nil {
		return err
	}
	initialSource, err := os.ReadFile(filepath.Join(initialDir, "source-initial.json"))
	if err != nil {
		return err
	}
	if err := writeImmutableArtifact(filepath.Join(dir, "source-initial.json"), initialSource); err != nil {
		return err
	}
	if err := writeImmutableArtifact(filepath.Join(dir, "source-final-delta.json"), source); err != nil {
		return err
	}
	for _, name := range []string{"validation-initial.json", "report-initial.json"} {
		data, err := os.ReadFile(filepath.Join(initialDir, name))
		if err != nil {
			return err
		}
		if err := writeImmutableArtifact(filepath.Join(dir, name), data); err != nil {
			return err
		}
	}
	if err := writeJSONArtifact(filepath.Join(dir, "validation-final.json"), report.Validation); err != nil {
		return err
	}
	if err := writeJSONArtifact(filepath.Join(dir, "report-final.json"), report); err != nil {
		return err
	}
	if err := writeJSONArtifact(filepath.Join(dir, "rollback.json"), rollback); err != nil {
		return err
	}
	if err := writeRollbackMarkdown(filepath.Join(dir, "rollback.md"), report); err != nil {
		return err
	}
	probe, err := os.ReadFile(filepath.Join(dir, "source-final-probe.json"))
	if err != nil {
		return fmt.Errorf("read frozen Grite probe artifact: %w", err)
	}
	probeHash := SHA256Bytes(probe)
	if sourceInfo.ProbeRawHash != "" && sourceInfo.ProbeRawHash != probeHash {
		return errors.New("frozen Grite probe hash does not match prepared probe")
	}
	sourceInfo.ProbeRawHash = probeHash
	manifest := initialManifest
	manifest.State = "final"
	manifest.Frozen = sourceInfo.FrozenValidated
	manifest.FinalCursorMS = sourceInfo.CursorMS
	manifest.FinalRawHash = SHA256Bytes(source)
	manifest.FinalProbeRawHash = sourceInfo.ProbeRawHash
	manifest.FinalSourceMode = sourceInfo.SourceMode
	manifest.FinalSourceHash = report.Validation.SourceHash
	manifest.FinalImportFingerprint = report.Validation.ImportFingerprint
	manifest.Watermark = report.Watermark
	manifest.GeneratedAt = time.UnixMilli(snapshot.Meta.GeneratedTS).UTC()
	manifest.FinalWALHead = cloneRaw(sourceInfo.WALHead)
	manifest.FinalProbeWALHead = cloneRaw(sourceInfo.ProbeWALHead)
	manifest.Warnings = report.Warnings
	if err := writeJSONArtifact(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		return err
	}
	if err := writeArtifactChecksums(dir); err != nil {
		return err
	}
	return VerifyArtifactChecksums(dir)
}

func VerifyArtifactChecksums(dir string) error {
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("Grite artifact generation is not a real directory: %s", dir)
	}
	manifest, err := loadArtifactManifestUnchecked(dir)
	if err != nil {
		return err
	}
	allowed := artifactFiles(manifest.State)
	if allowed == nil {
		return fmt.Errorf("unsupported Grite artifact state %q", manifest.State)
	}
	checksumData, err := os.ReadFile(filepath.Join(dir, "checksums.sha256"))
	if err != nil {
		return fmt.Errorf("read Grite artifact checksums: %w", err)
	}
	expected := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(checksumData)), "\n") {
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 || len(parts[0]) != 64 {
			return fmt.Errorf("invalid Grite artifact checksum line %q", line)
		}
		name := strings.TrimSpace(parts[1])
		if !allowed[name] || name == "checksums.sha256" {
			return fmt.Errorf("unexpected Grite artifact checksum entry %q", name)
		}
		if _, exists := expected[name]; exists {
			return fmt.Errorf("duplicate Grite artifact checksum entry %q", name)
		}
		expected[name] = parts[0]
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, entry := range entries {
		name := entry.Name()
		if !allowed[name] {
			return fmt.Errorf("unexpected file in Grite artifact generation: %s", name)
		}
		info, err := os.Lstat(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("Grite artifact %s is not a regular file", name)
		}
		seen[name] = true
		if name == "checksums.sha256" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if expected[name] != SHA256Bytes(data) {
			return fmt.Errorf("Grite artifact checksum mismatch for %s", name)
		}
	}
	for name := range allowed {
		if !seen[name] {
			return fmt.Errorf("Grite artifact is missing %s", name)
		}
	}
	if manifest.InitialRawHash != "" {
		data, err := os.ReadFile(filepath.Join(dir, "source-initial.json"))
		if err != nil {
			return err
		}
		if SHA256Bytes(data) != manifest.InitialRawHash {
			return errors.New("initial Grite raw source hash does not match manifest")
		}
	}
	if manifest.State == "final" && manifest.FinalRawHash != "" {
		data, err := os.ReadFile(filepath.Join(dir, "source-final-delta.json"))
		if err != nil {
			return err
		}
		if SHA256Bytes(data) != manifest.FinalRawHash {
			return errors.New("final Grite raw source hash does not match manifest")
		}
	}
	return verifyArtifactContent(dir, manifest)
}

func verifyArtifactContent(dir string, manifest ArtifactManifest) error {
	initialData, err := os.ReadFile(filepath.Join(dir, "source-initial.json"))
	if err != nil {
		return err
	}
	initial, err := griteexport.DecodeFile(initialData)
	if err != nil {
		return fmt.Errorf("decode initial Grite artifact source: %w", err)
	}
	initialDocument, err := Normalize(initial)
	if err != nil {
		return fmt.Errorf("normalize initial Grite artifact source: %w", err)
	}
	if initialDocument.SourceHash != manifest.InitialSourceHash {
		return errors.New("initial Grite source hash does not match manifest")
	}
	initialReport, err := LoadImportReport(dir, "initial")
	if err != nil {
		return fmt.Errorf("load initial Grite artifact report: %w", err)
	}
	if initialReport.WorkspaceID != manifest.WorkspaceID ||
		initialReport.Validation.SourceHash != manifest.InitialSourceHash ||
		initialReport.Validation.ImportFingerprint != manifest.InitialImportFingerprint {
		return errors.New("initial Grite artifact report does not match manifest")
	}
	if manifest.State != "final" {
		return nil
	}
	if !manifest.Frozen {
		return errors.New("final Grite artifact is not marked frozen")
	}
	if manifest.FinalSourceMode == "live" {
		if err := ValidateFrozenWALHeads(manifest.FinalWALHead, manifest.FinalProbeWALHead); err != nil {
			return fmt.Errorf("revalidate final live Grite WAL head: %w", err)
		}
	}
	deltaData, err := os.ReadFile(filepath.Join(dir, "source-final-delta.json"))
	if err != nil {
		return err
	}
	delta, err := griteexport.DecodeFile(deltaData)
	if err != nil {
		return fmt.Errorf("decode final Grite delta artifact: %w", err)
	}
	merged, err := MergeSnapshots(initial, delta)
	if err != nil {
		return fmt.Errorf("merge final Grite artifact sources: %w", err)
	}
	finalDocument, err := Normalize(merged)
	if err != nil {
		return fmt.Errorf("normalize final Grite artifact source: %w", err)
	}
	if finalDocument.SourceHash != manifest.FinalSourceHash {
		return errors.New("final Grite source hash does not match manifest")
	}
	finalReport, err := LoadImportReport(dir, "final")
	if err != nil {
		return fmt.Errorf("load final Grite artifact report: %w", err)
	}
	if finalReport.WorkspaceID != manifest.WorkspaceID ||
		finalReport.Validation.SourceHash != manifest.FinalSourceHash ||
		finalReport.Validation.ImportFingerprint != manifest.FinalImportFingerprint {
		return errors.New("final Grite artifact report does not match manifest")
	}
	probeData, err := os.ReadFile(filepath.Join(dir, "source-final-probe.json"))
	if err != nil {
		return err
	}
	if manifest.FinalProbeRawHash == "" || SHA256Bytes(probeData) != manifest.FinalProbeRawHash {
		return errors.New("final Grite probe hash does not match manifest")
	}
	probe, err := griteexport.DecodeFile(probeData)
	if err != nil {
		return fmt.Errorf("decode final Grite probe artifact: %w", err)
	}
	if err := ValidateFrozenProbe(merged, probe); err != nil {
		return fmt.Errorf("revalidate final frozen Grite probe: %w", err)
	}
	return nil
}

func LoadArtifactManifest(dir string) (ArtifactManifest, error) {
	return loadArtifactManifestUnchecked(dir)
}

func loadArtifactManifestUnchecked(dir string) (ArtifactManifest, error) {
	var manifest ArtifactManifest
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return manifest, fmt.Errorf("read Grite import manifest: %w", err)
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return manifest, fmt.Errorf("decode Grite import manifest: %w", err)
	}
	if manifest.SchemaVersion != artifactSchemaVersion {
		return manifest, fmt.Errorf("unsupported Grite artifact schema version %d", manifest.SchemaVersion)
	}
	return manifest, nil
}

func LoadInitialSnapshot(dir string) (griteexport.Snapshot, []byte, error) {
	data, err := os.ReadFile(filepath.Join(dir, "source-initial.json"))
	if err != nil {
		return griteexport.Snapshot{}, nil, fmt.Errorf("read initial Grite source: %w", err)
	}
	snapshot, err := griteexport.DecodeFile(data)
	if err != nil {
		return griteexport.Snapshot{}, nil, err
	}
	return snapshot, data, nil
}

func LoadRollback(dir string) (Rollback, error) {
	var rollback Rollback
	data, err := os.ReadFile(filepath.Join(dir, "rollback.json"))
	if err != nil {
		return rollback, fmt.Errorf("read Grite rollback artifact: %w", err)
	}
	if err := json.Unmarshal(data, &rollback); err != nil {
		return rollback, fmt.Errorf("decode Grite rollback artifact: %w", err)
	}
	return rollback, nil
}

func LoadImportReport(dir, state string) (*ImportReport, error) {
	if state != "initial" && state != "final" {
		return nil, fmt.Errorf("unsupported Grite report state %q", state)
	}
	data, err := os.ReadFile(filepath.Join(dir, "report-"+state+".json"))
	if err != nil {
		return nil, err
	}
	var report ImportReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

// MergeRollback keeps the earliest before-image and unions all inserted keys.
// It is used when a frozen final delta adds source events after the initial run.
func MergeRollback(initial, delta Rollback) Rollback {
	merged := initial
	if !merged.Native.WorkspaceCreated && merged.Native.WorkspaceBefore == nil {
		merged.Native.WorkspaceBefore = delta.Native.WorkspaceBefore
		merged.Native.WorkspacePathsBefore = delta.Native.WorkspacePathsBefore
	}
	merged.Native.WorkspaceCreated = merged.Native.WorkspaceCreated || delta.Native.WorkspaceCreated
	merged.Native.CreatedIssueIDs = uniqueUUIDs(append(merged.Native.CreatedIssueIDs, delta.Native.CreatedIssueIDs...))
	createdIssues := make(map[uuid.UUID]bool, len(merged.Native.CreatedIssueIDs))
	for _, id := range merged.Native.CreatedIssueIDs {
		createdIssues[id] = true
	}
	merged.Native.CreatedAliases = uniqueAliases(append(merged.Native.CreatedAliases, delta.Native.CreatedAliases...))
	merged.Native.InsertedEvents = uniqueEventKeys(append(merged.Native.InsertedEvents, delta.Native.InsertedEvents...))
	initialInsertedRelationshipKeys := make(map[string]bool, len(initial.Native.InsertedRelationships))
	for _, relationship := range initial.Native.InsertedRelationships {
		initialInsertedRelationshipKeys[nativeRelationshipKey(relationship)] = true
	}
	cancelledRelationships := map[string]bool{}
	for _, relationship := range delta.Native.DeletedRelationships {
		key := nativeRelationshipKey(relationship)
		if initialInsertedRelationshipKeys[key] {
			cancelledRelationships[key] = true
		}
	}
	merged.Native.InsertedRelationships = uniqueRelationships(append(merged.Native.InsertedRelationships, delta.Native.InsertedRelationships...))
	merged.Native.DeletedRelationships = uniqueRelationships(append(merged.Native.DeletedRelationships, delta.Native.DeletedRelationships...))
	if len(cancelledRelationships) > 0 {
		merged.Native.InsertedRelationships = filterRelationships(merged.Native.InsertedRelationships, cancelledRelationships)
		merged.Native.DeletedRelationships = filterRelationships(merged.Native.DeletedRelationships, cancelledRelationships)
	}
	merged.Native.InsertedPromptRunLinks = uniquePromptLinks(append(merged.Native.InsertedPromptRunLinks, delta.Native.InsertedPromptRunLinks...))
	merged.Native.InsertedPlanLinks = uniquePlanLinks(append(merged.Native.InsertedPlanLinks, delta.Native.InsertedPlanLinks...))
	preimages := make(map[uuid.UUID]native.ImportIssuePreimage)
	for _, preimage := range append(initial.Native.IssuePreimages, delta.Native.IssuePreimages...) {
		if createdIssues[preimage.Issue.ID] {
			continue
		}
		if _, exists := preimages[preimage.Issue.ID]; !exists {
			preimages[preimage.Issue.ID] = preimage
		}
	}
	merged.Native.IssuePreimages = merged.Native.IssuePreimages[:0]
	for _, preimage := range preimages {
		merged.Native.IssuePreimages = append(merged.Native.IssuePreimages, preimage)
	}
	sort.Slice(merged.Native.IssuePreimages, func(i, j int) bool {
		return merged.Native.IssuePreimages[i].Issue.ID.String() < merged.Native.IssuePreimages[j].Issue.ID.String()
	})
	merged.CreatedCaptainPlanIDs = uniqueUUIDs(append(merged.CreatedCaptainPlanIDs, delta.CreatedCaptainPlanIDs...))
	merged.AppendedCaptainRevisionIDs = uniqueUUIDs(append(merged.AppendedCaptainRevisionIDs, delta.AppendedCaptainRevisionIDs...))
	return merged
}

func ensureArtifactParent(parent string) error {
	parent, err := filepath.Abs(parent)
	if err != nil {
		return err
	}
	var missing []string
	current := parent
	for {
		info, err := os.Lstat(current)
		if err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("Grite artifact parent is not a real directory: %s", current)
			}
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		missing = append(missing, current)
		next := filepath.Dir(current)
		if next == current {
			return fmt.Errorf("find existing parent for Grite artifact directory %s", parent)
		}
		current = next
	}
	for i := len(missing) - 1; i >= 0; i-- {
		dir := missing[i]
		mkdirErr := os.Mkdir(dir, 0o700)
		if mkdirErr != nil && !errors.Is(mkdirErr, os.ErrExist) {
			return fmt.Errorf("create Grite artifact parent %s: %w", dir, mkdirErr)
		}
		info, err := os.Lstat(dir)
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("Grite artifact parent is not a real directory: %s", dir)
		}
		if mkdirErr == nil {
			if err := os.Chmod(dir, 0o700); err != nil {
				return err
			}
		}
		if err := syncDirectory(dir); err != nil {
			return fmt.Errorf("sync new Grite artifact parent %s: %w", dir, err)
		}
		if err := syncDirectory(filepath.Dir(dir)); err != nil {
			return fmt.Errorf("persist new Grite artifact parent %s: %w", dir, err)
		}
	}
	return nil
}

func ensureArtifactDir(dir string) error {
	info, err := os.Lstat(dir)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create Grite artifact directory: %w", err)
		}
		return os.Chmod(dir, 0o700)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("Grite artifact path is not a real directory: %s", dir)
	}
	return nil
}

func ensureUncommittedGeneration(dir string, allowedExisting []string) error {
	if err := ensureArtifactDir(dir); err != nil {
		return err
	}
	if _, err := os.Lstat(filepath.Join(dir, "manifest.json")); err == nil {
		return fmt.Errorf("%w: %s", ErrArtifactCommitted, dir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	allowed := map[string]bool{}
	for _, name := range allowedExisting {
		allowed[name] = true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !allowed[entry.Name()] {
			return fmt.Errorf("pending Grite artifact contains unexpected entry %s", entry.Name())
		}
		info, err := os.Lstat(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("pending Grite artifact entry %s is not regular", entry.Name())
		}
	}
	return nil
}

func writeJSONArtifact(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode Grite artifact %s: %w", filepath.Base(path), err)
	}
	data = append(data, '\n')
	return atomicWriteArtifact(path, data)
}

func writeImmutableArtifact(path string, data []byte) error {
	existing, err := os.ReadFile(path)
	if err == nil {
		if string(existing) != string(data) {
			return fmt.Errorf("immutable Grite artifact %s already contains different data", filepath.Base(path))
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read immutable Grite artifact %s: %w", filepath.Base(path), err)
	}
	return atomicWriteArtifact(path, data)
}

func atomicWriteArtifact(path string, data []byte) (returnErr error) {
	dir := filepath.Dir(path)
	if err := ensureArtifactDir(dir); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(dir, ".artifact-*")
	if err != nil {
		return fmt.Errorf("create temporary Grite artifact: %w", err)
	}
	temporaryName := temporary.Name()
	defer func() {
		if err := os.Remove(temporaryName); err != nil && !errors.Is(err, os.ErrNotExist) {
			returnErr = errors.Join(returnErr, err)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return fmt.Errorf("publish Grite artifact %s: %w", filepath.Base(path), err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return err
	}
	return syncDirectory(dir)
}

func writeRollbackMarkdown(path string, report *ImportReport) error {
	content := "# Grite migration rollback\n\n" +
		"Workspace: `" + report.WorkspaceID.String() + "`\n\n" +
		"Expected post-import Gavel checksum: `" + report.Validation.TargetChecksum + "`\n\n" +
		"Expected post-import Captain checksum: `" + report.Validation.CaptainChecksum + "`\n\n" +
		"Import fingerprint: `" + report.Validation.ImportFingerprint + "`\n\n" +
		"Stop all Gavel writers, verify both checksums still match, retain the previous Gavel binary, " +
		"and apply the guarded inverse operations described by `rollback.json` in reverse order. " +
		"Do not delete Captain plans or revisions unless their IDs are listed in that artifact.\n"
	return atomicWriteArtifact(path, []byte(content))
}

func writeArtifactChecksums(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "checksums.sha256" || strings.HasPrefix(entry.Name(), ".artifact-") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	var builder strings.Builder
	for _, name := range names {
		file, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, bufio.NewReader(file))
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		builder.WriteString(hex.EncodeToString(hash.Sum(nil)))
		builder.WriteString("  ")
		builder.WriteString(name)
		builder.WriteByte('\n')
	}
	return atomicWriteArtifact(filepath.Join(dir, "checksums.sha256"), []byte(builder.String()))
}

func artifactFiles(state string) map[string]bool {
	files := []string{
		"checksums.sha256", "manifest.json", "report-initial.json", "rollback.json", "rollback.md",
		"source-initial.json", "validation-initial.json",
	}
	if state == "final" {
		files = append(files, "report-final.json", "source-final-delta.json", "source-final-probe.json", "validation-final.json")
	} else if state != "initial" {
		return nil
	}
	out := make(map[string]bool, len(files))
	for _, file := range files {
		out[file] = true
	}
	return out
}

func syncDirectory(dir string) error {
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func cloneRaw(raw json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), raw...)
}

func sourceInfoFor(provenance []ArtifactProvenance) ArtifactProvenance {
	if len(provenance) == 0 {
		return ArtifactProvenance{}
	}
	return provenance[0]
}

func validateArtifactProvenance(info ArtifactProvenance, final bool) error {
	if info.SourceMode != "live" && info.SourceMode != "file" {
		return fmt.Errorf("Grite artifact source mode must be live or file, got %q", info.SourceMode)
	}
	if strings.TrimSpace(info.SourceRoot) == "" {
		return errors.New("Grite artifact source root is required")
	}
	if !final && info.CursorMS != 0 {
		return fmt.Errorf("initial Grite artifact cursor must be zero, got %d", info.CursorMS)
	}
	if final && !info.FrozenValidated {
		return errors.New("final Grite artifact requires a validated frozen probe")
	}
	return nil
}

func equalRawJSON(left, right json.RawMessage) (bool, error) {
	canonical := func(raw json.RawMessage) ([]byte, error) {
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		var value any
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		var trailing any
		if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
			if err == nil {
				return nil, errors.New("multiple JSON values")
			}
			return nil, err
		}
		return json.Marshal(value)
	}
	leftJSON, err := canonical(left)
	if err != nil {
		return false, err
	}
	rightJSON, err := canonical(right)
	if err != nil {
		return false, err
	}
	return bytes.Equal(leftJSON, rightJSON), nil
}

func ValidateFrozenWALHeads(delta, probe json.RawMessage) error {
	delta = bytes.TrimSpace(delta)
	probe = bytes.TrimSpace(probe)
	if len(delta) == 0 || len(probe) == 0 || bytes.Equal(delta, []byte("null")) || bytes.Equal(probe, []byte("null")) {
		return errors.New("live Grite delta or probe is missing its WAL head")
	}
	equal, err := equalRawJSON(delta, probe)
	if err != nil {
		return fmt.Errorf("compare live Grite WAL heads: %w", err)
	}
	if !equal {
		return errors.New("Grite source is not frozen: delta and probe WAL heads differ")
	}
	return nil
}

func safeArtifactComponent(value string) string {
	value = strings.TrimSpace(value)
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			builder.WriteRune(r)
		default:
			builder.WriteByte('_')
		}
	}
	return strings.Trim(builder.String(), "._")
}

func uniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	var out []uuid.UUID
	for _, value := range values {
		if value != uuid.Nil && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sortUUIDs(out)
	return out
}

func uniqueAliases(values []native.ImportAliasKey) []native.ImportAliasKey {
	seen := map[string]native.ImportAliasKey{}
	for _, value := range values {
		seen[value.WorkspaceID.String()+"\x00"+value.Alias] = value
	}
	var out []native.ImportAliasKey
	for _, value := range seen {
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Alias < out[j].Alias })
	return out
}

func uniqueEventKeys(values []native.ImportEventKey) []native.ImportEventKey {
	seen := map[string]native.ImportEventKey{}
	for _, value := range values {
		seen[value.Source+"\x00"+value.SourceID] = value
	}
	var out []native.ImportEventKey
	for _, value := range seen {
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Source+"\x00"+out[i].SourceID < out[j].Source+"\x00"+out[j].SourceID
	})
	return out
}

func uniqueRelationships(values []native.Relationship) []native.Relationship {
	seen := map[string]native.Relationship{}
	for _, value := range values {
		seen[nativeRelationshipKey(value)] = value
	}
	var out []native.Relationship
	for _, value := range seen {
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool {
		left := out[i].IssueID.String() + "\x00" + out[i].TargetIssueID.String() + "\x00" + string(out[i].Relation)
		right := out[j].IssueID.String() + "\x00" + out[j].TargetIssueID.String() + "\x00" + string(out[j].Relation)
		return left < right
	})
	return out
}

func nativeRelationshipKey(value native.Relationship) string {
	return value.WorkspaceID.String() + "\x00" + value.IssueID.String() + "\x00" + value.TargetIssueID.String() + "\x00" + string(value.Relation)
}

func filterRelationships(values []native.Relationship, remove map[string]bool) []native.Relationship {
	out := values[:0]
	for _, value := range values {
		if !remove[nativeRelationshipKey(value)] {
			out = append(out, value)
		}
	}
	return out
}

func uniquePromptLinks(values []native.PromptRunLink) []native.PromptRunLink {
	seen := map[uuid.UUID]native.PromptRunLink{}
	for _, value := range values {
		seen[value.PromptRunID] = value
	}
	var out []native.PromptRunLink
	for _, value := range seen {
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PromptRunID.String() < out[j].PromptRunID.String() })
	return out
}

func uniquePlanLinks(values []native.PlanLink) []native.PlanLink {
	seen := map[string]native.PlanLink{}
	for _, value := range values {
		seen[value.IssueID.String()+"\x00"+value.PlanID.String()] = value
	}
	var out []native.PlanLink
	for _, value := range seen {
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].IssueID.String()+"\x00"+out[i].PlanID.String() < out[j].IssueID.String()+"\x00"+out[j].PlanID.String()
	})
	return out
}
