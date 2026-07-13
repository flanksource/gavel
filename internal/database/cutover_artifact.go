package database

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const CutoverArtifactSchemaVersion = 1

const (
	cutoverManifestFile  = "manifest.json"
	cutoverReportFile    = "report.json"
	cutoverRollbackFile  = "rollback.json"
	cutoverRollbackGuide = "rollback.md"
	cutoverChecksumsFile = "checksums.sha256"
)

var (
	ErrCutoverArtifactCommitted = errors.New("database cutover artifact generation is already committed")
	ErrCutoverArtifactPending   = errors.New("database cutover artifact generation has an incomplete pending sibling")
	ErrCutoverArtifactInvalid   = errors.New("database cutover artifact is invalid")
)

// CutoverArtifactSnapshot is one exact pre-mutation source capture. Data may be
// textual schema/JSON or an opaque binary pg_dump archive; the writer never
// interprets or normalizes it.
type CutoverArtifactSnapshot struct {
	Name        string
	ContentType string
	Data        []byte
}

// CutoverArtifactRequest is the complete rollback evidence that must be
// durable before a database cutover starts. Report and Rollback are encoded as
// JSON. Snapshot names are flat filenames so an artifact cannot escape its
// private generation directory.
type CutoverArtifactRequest struct {
	Cutover              string
	Generation           string
	CreatedAt            time.Time
	Metadata             map[string]any
	Report               any
	Rollback             any
	RollbackInstructions string
	Snapshots            []CutoverArtifactSnapshot
}

// CutoverArtifactFile records immutable payload metadata. manifest.json and
// checksums.sha256 are excluded to avoid a circular manifest hash; the checksum
// file covers the manifest and every payload.
type CutoverArtifactFile struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
	SHA256      string `json:"sha256"`
}

type CutoverArtifactManifest struct {
	SchemaVersion int                   `json:"schemaVersion"`
	Cutover       string                `json:"cutover"`
	Generation    string                `json:"generation"`
	CreatedAt     time.Time             `json:"createdAt"`
	Metadata      map[string]any        `json:"metadata,omitempty"`
	Files         []CutoverArtifactFile `json:"files"`
}

// CutoverArtifactResult is safe to persist in an orchestrator checkpoint. The
// checksums map contains every artifact file except checksums.sha256 itself.
type CutoverArtifactResult struct {
	Directory string
	Manifest  CutoverArtifactManifest
	Checksums map[string]string
}

type cutoverArtifactPayload struct {
	metadata CutoverArtifactFile
	data     []byte
}

// WriteCutoverArtifact atomically commits a private rollback generation. It
// returns only after every file and directory entry has been fsynced and the
// finished generation has passed VerifyCutoverArtifact. Callers should invoke
// this before their first database mutation.
func WriteCutoverArtifact(target string, request CutoverArtifactRequest) (*CutoverArtifactResult, error) {
	target, err := normalizeCutoverArtifactTarget(target)
	if err != nil {
		return nil, err
	}
	request.Cutover = strings.TrimSpace(request.Cutover)
	if request.Cutover == "" {
		return nil, fmt.Errorf("%w: cutover name is required", ErrCutoverArtifactInvalid)
	}
	if request.Report == nil {
		return nil, fmt.Errorf("%w: report is required", ErrCutoverArtifactInvalid)
	}
	if request.Rollback == nil {
		return nil, fmt.Errorf("%w: rollback description is required", ErrCutoverArtifactInvalid)
	}
	if len(request.Snapshots) == 0 {
		return nil, fmt.Errorf("%w: at least one legacy schema/data snapshot is required", ErrCutoverArtifactInvalid)
	}

	report, err := marshalCutoverArtifactJSON(cutoverReportFile, request.Report)
	if err != nil {
		return nil, err
	}
	rollback, err := marshalCutoverArtifactJSON(cutoverRollbackFile, request.Rollback)
	if err != nil {
		return nil, err
	}
	generation := strings.TrimSpace(request.Generation)
	if generation == "" {
		generation = filepath.Base(target)
	}
	createdAt := request.CreatedAt.UTC()
	if request.CreatedAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	payloads := []cutoverArtifactPayload{
		newCutoverPayload(cutoverReportFile, "report", "application/json", report),
		newCutoverPayload(cutoverRollbackFile, "rollback", "application/json", rollback),
	}
	guide := strings.TrimSpace(request.RollbackInstructions)
	if guide == "" {
		guide = defaultCutoverRollbackGuide(request.Cutover, generation)
	}
	payloads = append(payloads, newCutoverPayload(cutoverRollbackGuide, "rollback-guide", "text/markdown", []byte(guide+"\n")))

	seen := map[string]bool{
		cutoverManifestFile:  true,
		cutoverReportFile:    true,
		cutoverRollbackFile:  true,
		cutoverRollbackGuide: true,
		cutoverChecksumsFile: true,
	}
	for _, snapshot := range request.Snapshots {
		if err := validateCutoverArtifactFilename(snapshot.Name); err != nil {
			return nil, err
		}
		if seen[snapshot.Name] {
			return nil, fmt.Errorf("%w: duplicate or reserved artifact filename %q", ErrCutoverArtifactInvalid, snapshot.Name)
		}
		seen[snapshot.Name] = true
		contentType := strings.TrimSpace(snapshot.ContentType)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		data := append([]byte(nil), snapshot.Data...)
		payloads = append(payloads, newCutoverPayload(snapshot.Name, "snapshot", contentType, data))
	}
	sort.Slice(payloads, func(i, j int) bool { return payloads[i].metadata.Name < payloads[j].metadata.Name })

	manifest := CutoverArtifactManifest{
		SchemaVersion: CutoverArtifactSchemaVersion,
		Cutover:       request.Cutover,
		Generation:    generation,
		CreatedAt:     createdAt,
		Metadata:      request.Metadata,
		Files:         make([]CutoverArtifactFile, 0, len(payloads)),
	}
	for _, payload := range payloads {
		manifest.Files = append(manifest.Files, payload.metadata)
	}
	manifestData, err := marshalCutoverArtifactJSON(cutoverManifestFile, manifest)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{cutoverManifestFile: manifestData}
	for _, payload := range payloads {
		files[payload.metadata.Name] = payload.data
	}
	fileNames := make([]string, 0, len(files))
	for name := range files {
		fileNames = append(fileNames, name)
	}
	sort.Strings(fileNames)
	checksums := make(map[string]string, len(fileNames))
	for _, name := range fileNames {
		checksums[name] = cutoverSHA256(files[name])
	}

	parent := filepath.Dir(target)
	if err := ensureCutoverArtifactParent(parent); err != nil {
		return nil, err
	}
	if err := rejectExistingCutoverTarget(target); err != nil {
		return nil, err
	}
	pending := filepath.Join(parent, "."+filepath.Base(target)+".pending-generation")
	if _, err := os.Lstat(pending); err == nil {
		existing, verifyErr := VerifyCutoverArtifact(pending)
		if verifyErr == nil && existing.Cutover == manifest.Cutover && existing.Generation == manifest.Generation {
			expected := files
			if request.CreatedAt.IsZero() {
				expected = cloneCutoverArtifactFiles(files)
				resumedManifest := manifest
				resumedManifest.CreatedAt = existing.CreatedAt
				expected[cutoverManifestFile], err = marshalCutoverArtifactJSON(cutoverManifestFile, resumedManifest)
				if err != nil {
					return nil, err
				}
			}
			matches, matchErr := cutoverArtifactMatchesFiles(pending, existing, expected)
			if matchErr != nil {
				return nil, matchErr
			}
			if matches {
				if err := syncCutoverDirectory(pending); err != nil {
					return nil, err
				}
				if err := rejectExistingCutoverTarget(target); err != nil {
					return nil, err
				}
				if err := os.Rename(pending, target); err != nil {
					return nil, fmt.Errorf("publish recovered database cutover artifact %s: %w", target, err)
				}
				if err := syncCutoverDirectory(parent); err != nil {
					return nil, fmt.Errorf("persist recovered database cutover artifact publication: %w", err)
				}
				return verifiedCutoverArtifactResult(target)
			}
		}
		if _, err := quarantinePendingCutoverArtifact(pending); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if err := os.Mkdir(pending, 0o700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("%w: %s", ErrCutoverArtifactPending, pending)
		}
		return nil, fmt.Errorf("reserve database cutover artifact %s: %w", pending, err)
	}
	cleanupPending := true
	defer func() {
		if cleanupPending {
			_ = os.RemoveAll(pending)
			_ = syncCutoverDirectory(parent)
		}
	}()
	if err := os.Chmod(pending, 0o700); err != nil {
		return nil, err
	}
	if err := syncCutoverDirectory(parent); err != nil {
		return nil, fmt.Errorf("persist database cutover artifact reservation: %w", err)
	}

	for _, name := range fileNames {
		if err := atomicWriteCutoverFile(filepath.Join(pending, name), files[name]); err != nil {
			return nil, err
		}
	}
	checksumData := encodeCutoverChecksums(checksums)
	if err := atomicWriteCutoverFile(filepath.Join(pending, cutoverChecksumsFile), checksumData); err != nil {
		return nil, err
	}
	if _, err := VerifyCutoverArtifact(pending); err != nil {
		return nil, fmt.Errorf("verify pending database cutover artifact: %w", err)
	}
	if err := syncCutoverDirectory(pending); err != nil {
		return nil, err
	}
	if err := rejectExistingCutoverTarget(target); err != nil {
		return nil, err
	}
	if err := os.Rename(pending, target); err != nil {
		return nil, fmt.Errorf("publish database cutover artifact %s: %w", target, err)
	}
	cleanupPending = false
	if err := syncCutoverDirectory(parent); err != nil {
		return nil, fmt.Errorf("persist database cutover artifact publication: %w", err)
	}
	verified, err := VerifyCutoverArtifact(target)
	if err != nil {
		return nil, err
	}
	return &CutoverArtifactResult{Directory: target, Manifest: *verified, Checksums: checksums}, nil
}

func cloneCutoverArtifactFiles(files map[string][]byte) map[string][]byte {
	cloned := make(map[string][]byte, len(files))
	for name, data := range files {
		cloned[name] = data
	}
	return cloned
}

func cutoverArtifactMatchesFiles(dir string, manifest *CutoverArtifactManifest, expected map[string][]byte) (bool, error) {
	if len(expected) != len(manifest.Files)+1 {
		return false, nil
	}
	for name, expectedData := range expected {
		actual, err := readPrivateCutoverFile(filepath.Join(dir, name))
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if !bytes.Equal(actual, expectedData) {
			return false, nil
		}
	}
	return true, nil
}

func quarantinePendingCutoverArtifact(pending string) (string, error) {
	parent := filepath.Dir(pending)
	prefix := strings.TrimSuffix(filepath.Base(pending), ".pending-generation") + ".incomplete-evidence-" +
		time.Now().UTC().Format("20060102T150405.000000000Z")
	for attempt := 0; attempt < 1000; attempt++ {
		candidate := filepath.Join(parent, prefix)
		if attempt > 0 {
			candidate = fmt.Sprintf("%s-%d", candidate, attempt)
		}
		if _, err := os.Lstat(candidate); err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		if err := os.Rename(pending, candidate); err != nil {
			return "", fmt.Errorf("preserve incomplete database cutover artifact %s: %w", pending, err)
		}
		if err := syncCutoverDirectory(parent); err != nil {
			return "", fmt.Errorf("persist incomplete database cutover artifact evidence: %w", err)
		}
		return candidate, nil
	}
	return "", fmt.Errorf("preserve incomplete database cutover artifact %s: no unique evidence name available", pending)
}

func verifiedCutoverArtifactResult(dir string) (*CutoverArtifactResult, error) {
	manifest, err := VerifyCutoverArtifact(dir)
	if err != nil {
		return nil, err
	}
	checksumData, err := readPrivateCutoverFile(filepath.Join(dir, cutoverChecksumsFile))
	if err != nil {
		return nil, err
	}
	checksums, err := decodeCutoverChecksums(checksumData)
	if err != nil {
		return nil, err
	}
	return &CutoverArtifactResult{Directory: dir, Manifest: *manifest, Checksums: checksums}, nil
}

// VerifyCutoverArtifact validates privacy, file membership, manifest metadata,
// sizes, and every SHA-256 entry. It never changes the artifact.
func VerifyCutoverArtifact(dir string) (*CutoverArtifactManifest, error) {
	dir, err := filepath.Abs(strings.TrimSpace(dir))
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o700 {
		return nil, fmt.Errorf("%w: artifact directory must be a real 0700 directory: %s", ErrCutoverArtifactInvalid, dir)
	}
	manifestData, err := readPrivateCutoverFile(filepath.Join(dir, cutoverManifestFile))
	if err != nil {
		return nil, err
	}
	var manifest CutoverArtifactManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("%w: decode manifest: %v", ErrCutoverArtifactInvalid, err)
	}
	if manifest.SchemaVersion != CutoverArtifactSchemaVersion || strings.TrimSpace(manifest.Cutover) == "" ||
		strings.TrimSpace(manifest.Generation) == "" || manifest.CreatedAt.IsZero() {
		return nil, fmt.Errorf("%w: unsupported or incomplete manifest", ErrCutoverArtifactInvalid)
	}

	expected := map[string]bool{
		cutoverManifestFile:  true,
		cutoverChecksumsFile: true,
	}
	manifestFiles := make(map[string]CutoverArtifactFile, len(manifest.Files))
	snapshotCount := 0
	for _, file := range manifest.Files {
		if err := validateCutoverArtifactFilename(file.Name); err != nil {
			return nil, err
		}
		if expected[file.Name] || file.Size < 0 || len(file.SHA256) != sha256.Size*2 {
			return nil, fmt.Errorf("%w: invalid or duplicate manifest file %q", ErrCutoverArtifactInvalid, file.Name)
		}
		if _, err := hex.DecodeString(file.SHA256); err != nil {
			return nil, fmt.Errorf("%w: invalid SHA-256 for %q", ErrCutoverArtifactInvalid, file.Name)
		}
		if err := validateCutoverArtifactFileMetadata(file); err != nil {
			return nil, err
		}
		if file.Role == "snapshot" {
			snapshotCount++
		}
		expected[file.Name] = true
		manifestFiles[file.Name] = file
	}
	for _, required := range []string{cutoverReportFile, cutoverRollbackFile, cutoverRollbackGuide} {
		if !expected[required] {
			return nil, fmt.Errorf("%w: manifest is missing %s", ErrCutoverArtifactInvalid, required)
		}
	}
	if snapshotCount == 0 {
		return nil, fmt.Errorf("%w: manifest has no legacy snapshot", ErrCutoverArtifactInvalid)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	actual := make(map[string]bool, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !expected[name] {
			return nil, fmt.Errorf("%w: unexpected artifact entry %q", ErrCutoverArtifactInvalid, name)
		}
		entryInfo, err := os.Lstat(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		if !entryInfo.Mode().IsRegular() || entryInfo.Mode()&os.ModeSymlink != 0 || entryInfo.Mode().Perm() != 0o600 {
			return nil, fmt.Errorf("%w: artifact file must be a regular 0600 file: %s", ErrCutoverArtifactInvalid, name)
		}
		actual[name] = true
	}
	for name := range expected {
		if !actual[name] {
			return nil, fmt.Errorf("%w: artifact is missing %s", ErrCutoverArtifactInvalid, name)
		}
	}

	checksumData, err := readPrivateCutoverFile(filepath.Join(dir, cutoverChecksumsFile))
	if err != nil {
		return nil, err
	}
	checksums, err := decodeCutoverChecksums(checksumData)
	if err != nil {
		return nil, err
	}
	if len(checksums) != len(expected)-1 {
		return nil, fmt.Errorf("%w: checksum file covers %d files, want %d", ErrCutoverArtifactInvalid, len(checksums), len(expected)-1)
	}
	for name := range expected {
		if name == cutoverChecksumsFile {
			continue
		}
		data, err := readPrivateCutoverFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		digest := cutoverSHA256(data)
		if checksums[name] != digest {
			return nil, fmt.Errorf("%w: checksum mismatch for %s", ErrCutoverArtifactInvalid, name)
		}
		if (name == cutoverReportFile || name == cutoverRollbackFile) && !json.Valid(data) {
			return nil, fmt.Errorf("%w: %s is not valid JSON", ErrCutoverArtifactInvalid, name)
		}
		if file, ok := manifestFiles[name]; ok {
			if file.SHA256 != digest || file.Size != int64(len(data)) {
				return nil, fmt.Errorf("%w: manifest metadata mismatch for %s", ErrCutoverArtifactInvalid, name)
			}
		}
	}
	return &manifest, nil
}

// ReadCutoverArtifactReport returns the exact report JSON after validating the
// complete generation. A caller cannot accidentally consume a report whose
// snapshot, rollback data, permissions, or checksums no longer match.
func ReadCutoverArtifactReport(dir string) (json.RawMessage, error) {
	if _, err := VerifyCutoverArtifact(dir); err != nil {
		return nil, err
	}
	report, err := readPrivateCutoverFile(filepath.Join(dir, cutoverReportFile))
	if err != nil {
		return nil, err
	}
	return json.RawMessage(report), nil
}

func validateCutoverArtifactFileMetadata(file CutoverArtifactFile) error {
	expectedRole := "snapshot"
	expectedContentType := ""
	switch file.Name {
	case cutoverReportFile:
		expectedRole = "report"
		expectedContentType = "application/json"
	case cutoverRollbackFile:
		expectedRole = "rollback"
		expectedContentType = "application/json"
	case cutoverRollbackGuide:
		expectedRole = "rollback-guide"
		expectedContentType = "text/markdown"
	}
	if file.Role != expectedRole {
		return fmt.Errorf("%w: invalid role %q for %s", ErrCutoverArtifactInvalid, file.Role, file.Name)
	}
	if strings.TrimSpace(file.ContentType) == "" || (expectedContentType != "" && file.ContentType != expectedContentType) {
		return fmt.Errorf("%w: invalid content type %q for %s", ErrCutoverArtifactInvalid, file.ContentType, file.Name)
	}
	return nil
}

func newCutoverPayload(name, role, contentType string, data []byte) cutoverArtifactPayload {
	return cutoverArtifactPayload{
		metadata: CutoverArtifactFile{
			Name: name, Role: role, ContentType: contentType,
			Size: int64(len(data)), SHA256: cutoverSHA256(data),
		},
		data: data,
	}
}

func normalizeCutoverArtifactTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("%w: target directory is required", ErrCutoverArtifactInvalid)
	}
	target, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	base := filepath.Base(target)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "", fmt.Errorf("%w: target must name a generation directory", ErrCutoverArtifactInvalid)
	}
	return target, nil
}

func rejectExistingCutoverTarget(target string) error {
	if _, err := os.Lstat(target); err == nil {
		return fmt.Errorf("%w: %s", ErrCutoverArtifactCommitted, target)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func validateCutoverArtifactFilename(name string) error {
	if strings.TrimSpace(name) != name || name == "" || name == "." || name == ".." ||
		filepath.Base(name) != name || strings.ContainsAny(name, "/\\\r\n\x00") || strings.HasPrefix(name, ".") {
		return fmt.Errorf("%w: unsafe artifact filename %q", ErrCutoverArtifactInvalid, name)
	}
	return nil
}

func marshalCutoverArtifactJSON(name string, value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode database cutover artifact %s: %w", name, err)
	}
	return append(data, '\n'), nil
}

func defaultCutoverRollbackGuide(cutover, generation string) string {
	return "# Database cutover rollback\n\n" +
		"Cutover: `" + cutover + "`\n\n" +
		"Generation: `" + generation + "`\n\n" +
		"Stop all writers and verify the post-cutover state recorded in `report.json` before applying the guarded inverse operations in `rollback.json`. Retain this generation and the previous binary until validation and rollback windows have closed."
}

func cutoverSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func encodeCutoverChecksums(checksums map[string]string) []byte {
	names := make([]string, 0, len(checksums))
	for name := range checksums {
		names = append(names, name)
	}
	sort.Strings(names)
	var out strings.Builder
	for _, name := range names {
		out.WriteString(checksums[name])
		out.WriteString("  ")
		out.WriteString(name)
		out.WriteByte('\n')
	}
	return []byte(out.String())
}

func decodeCutoverChecksums(data []byte) (map[string]string, error) {
	checksums := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 || len(parts[0]) != sha256.Size*2 {
			return nil, fmt.Errorf("%w: invalid checksum line %q", ErrCutoverArtifactInvalid, line)
		}
		if _, err := hex.DecodeString(parts[0]); err != nil {
			return nil, fmt.Errorf("%w: invalid checksum digest for %q", ErrCutoverArtifactInvalid, parts[1])
		}
		if err := validateCutoverArtifactFilename(parts[1]); err != nil {
			return nil, err
		}
		if parts[1] == cutoverChecksumsFile || checksums[parts[1]] != "" {
			return nil, fmt.Errorf("%w: duplicate or self-referential checksum for %q", ErrCutoverArtifactInvalid, parts[1])
		}
		checksums[parts[1]] = parts[0]
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return checksums, nil
}

func ensureCutoverArtifactParent(parent string) error {
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
				return fmt.Errorf("%w: artifact parent is not a real directory: %s", ErrCutoverArtifactInvalid, current)
			}
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		missing = append(missing, current)
		next := filepath.Dir(current)
		if next == current {
			return fmt.Errorf("find existing parent for database cutover artifact %s", parent)
		}
		current = next
	}
	for i := len(missing) - 1; i >= 0; i-- {
		dir := missing[i]
		if err := os.Mkdir(dir, 0o700); err != nil {
			return fmt.Errorf("create database cutover artifact parent %s: %w", dir, err)
		}
		if err := os.Chmod(dir, 0o700); err != nil {
			return err
		}
		if err := syncCutoverDirectory(dir); err != nil {
			return err
		}
		if err := syncCutoverDirectory(filepath.Dir(dir)); err != nil {
			return err
		}
	}
	return nil
}

func atomicWriteCutoverFile(path string, data []byte) (returnErr error) {
	dir := filepath.Dir(path)
	temporary, err := os.CreateTemp(dir, ".artifact-*")
	if err != nil {
		return err
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
		return err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return err
	}
	return syncCutoverDirectory(dir)
}

func readPrivateCutoverFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o600 {
		return nil, fmt.Errorf("%w: artifact file must be a regular 0600 file: %s", ErrCutoverArtifactInvalid, filepath.Base(path))
	}
	return os.ReadFile(path)
}

func syncCutoverDirectory(dir string) error {
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}
