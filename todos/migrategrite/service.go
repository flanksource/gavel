package migrategrite

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/griteexport"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrWarnings = errors.New("Grite import produced warnings")

type Service struct {
	db       *gorm.DB
	readFile func(string) ([]byte, error)
}

func NewService(db *gorm.DB) (*Service, error) {
	if db == nil {
		return nil, errors.New("Grite import database is nil")
	}
	return &Service{db: db, readFile: os.ReadFile}, nil
}

type ImportOptions struct {
	Workspace native.ImportWorkspace
	PlanRoot  string
	Strict    bool
	// DeferActivePromptRuns imports every historical Captain link but leaves
	// active_prompt_run_id unset. The CLI uses this for its initial comparison
	// pass so Captain projection cannot drift the guarded target before the
	// frozen final delta activates the authoritative current run.
	DeferActivePromptRuns  bool
	RequireNoPriorImport   bool
	ExpectedTargetChecksum string
	BeforeCommit           func(*ImportReport) error
}

type ImportValidation struct {
	SourceHash         string `json:"sourceHash"`
	ImportFingerprint  string `json:"importFingerprint"`
	TargetChecksum     string `json:"targetChecksum"`
	CaptainChecksum    string `json:"captainChecksum"`
	SourceIssueCount   int    `json:"sourceIssueCount"`
	SourceEventCount   int    `json:"sourceEventCount"`
	RelationshipCount  int    `json:"relationshipCount"`
	PromptRunLinkCount int    `json:"promptRunLinkCount"`
	PlanLinkCount      int    `json:"planLinkCount"`
	WarningCount       int    `json:"warningCount"`
}

type Rollback struct {
	Native                     native.ImportRollback `json:"native"`
	CreatedCaptainPlanIDs      []uuid.UUID           `json:"createdCaptainPlanIds,omitempty"`
	AppendedCaptainRevisionIDs []uuid.UUID           `json:"appendedCaptainRevisionIds,omitempty"`
}

type ImportReport struct {
	WorkspaceID uuid.UUID             `json:"workspaceId"`
	Watermark   griteexport.Watermark `json:"watermark"`
	Counts      native.ImportCounts   `json:"counts"`
	Validation  ImportValidation      `json:"validation"`
	Warnings    []Warning             `json:"warnings,omitempty"`
	Rollback    Rollback              `json:"rollback"`
}

type resolvedCaptain struct {
	promptLinks       []native.ImportPromptRunLink
	planLinks         []native.ImportPlanLink
	activeRuns        map[string]uuid.UUID
	selectedPlans     map[string]uuid.UUID
	warnings          []Warning
	createdPlans      []uuid.UUID
	appendedRevisions []uuid.UUID
	planRevisions     []pendingPlanRevision
}

type pendingPlanRevision struct {
	IssueSourceID string
	PlanID        uuid.UUID
	Content       string
	ContentHash   string
}

// Import normalizes a snapshot, resolves authoritative Captain rows, and
// applies the entire migration through one shared database transaction.
func (service *Service) Import(ctx context.Context, snapshot griteexport.Snapshot, options ImportOptions) (*ImportReport, error) {
	if service == nil || service.db == nil {
		return nil, errors.New("Grite import service is nil")
	}
	if err := ValidateFullSnapshotHistory(snapshot); err != nil {
		return nil, err
	}
	document, err := Normalize(snapshot)
	if err != nil {
		return nil, err
	}
	workspace := importWorkspace(document, options.Workspace)
	planRoot := strings.TrimSpace(options.PlanRoot)
	if planRoot == "" {
		planRoot = workspace.RootPath
	}

	var report *ImportReport
	err = service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		captain, err := captaindb.Use(tx)
		if err != nil {
			return fmt.Errorf("reuse Gavel transaction for Captain import: %w", err)
		}
		resolved, err := service.resolveCaptain(ctx, tx, captain, document, planRoot)
		if err != nil {
			return err
		}
		allWarnings := dedupeWarnings(append(append([]Warning(nil), document.Warnings...), resolved.warnings...))
		if options.Strict && len(allWarnings) > 0 {
			return fmt.Errorf("%w: %d warning(s), first: %s", ErrWarnings, len(allWarnings), allWarnings[0].Message)
		}

		resolvedForNative := resolved
		if options.DeferActivePromptRuns {
			resolvedForNative.activeRuns = map[string]uuid.UUID{}
		}
		fingerprint, err := importFingerprint(document.SourceHash, resolvedForNative, allWarnings)
		if err != nil {
			return err
		}
		batch := buildNativeBatch(document, workspace, resolvedForNative, allWarnings)
		batch.Fingerprint = fingerprint
		batch.RequireNoPriorImport = options.RequireNoPriorImport
		batch.ExpectedChecksum = options.ExpectedTargetChecksum
		repository, err := native.NewRepository(tx)
		if err != nil {
			return err
		}
		nativeResult, err := repository.ApplyImport(ctx, batch)
		if err != nil {
			return err
		}
		if err := applyPendingPlanRevisions(ctx, captain, &resolved); err != nil {
			return err
		}
		captainChecksum, err := checksumResolvedCaptain(ctx, captain, resolved)
		if err != nil {
			return err
		}
		sortUUIDs(resolved.createdPlans)
		sortUUIDs(resolved.appendedRevisions)
		report = buildImportReport(document, resolved, allWarnings, nativeResult, fingerprint, captainChecksum)
		if report.Counts.IssuesSeen != report.Validation.SourceIssueCount {
			return fmt.Errorf("import validation mismatch: saw %d issues, source has %d", report.Counts.IssuesSeen, report.Validation.SourceIssueCount)
		}
		if strings.TrimSpace(report.Validation.TargetChecksum) == "" || strings.TrimSpace(report.Validation.CaptainChecksum) == "" {
			return errors.New("import validation produced an empty target checksum")
		}
		if options.BeforeCommit != nil {
			if err := options.BeforeCommit(report); err != nil {
				return fmt.Errorf("write pre-commit Grite migration artifact: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return report, nil
}

func buildImportReport(
	document Document,
	resolved resolvedCaptain,
	warnings []Warning,
	nativeResult *native.ImportResult,
	fingerprint, captainChecksum string,
) *ImportReport {
	return &ImportReport{
		WorkspaceID: nativeResult.WorkspaceID,
		Watermark:   document.Watermark,
		Counts:      nativeResult.Counts,
		Validation: ImportValidation{
			SourceHash:         document.SourceHash,
			ImportFingerprint:  fingerprint,
			TargetChecksum:     nativeResult.Checksum,
			CaptainChecksum:    captainChecksum,
			SourceIssueCount:   len(document.Issues),
			SourceEventCount:   len(document.Events),
			RelationshipCount:  len(document.Relationships),
			PromptRunLinkCount: len(resolved.promptLinks),
			PlanLinkCount:      len(resolved.planLinks),
			WarningCount:       len(warnings),
		},
		Warnings: warnings,
		Rollback: Rollback{
			Native:                     nativeResult.Rollback,
			CreatedCaptainPlanIDs:      resolved.createdPlans,
			AppendedCaptainRevisionIDs: resolved.appendedRevisions,
		},
	}
}

func importWorkspace(document Document, workspace native.ImportWorkspace) native.ImportWorkspace {
	if len(document.Issues) == 0 {
		return workspace
	}
	if workspace.CreatedAt.IsZero() {
		workspace.CreatedAt = document.Issues[0].CreatedAt
	}
	if workspace.UpdatedAt.IsZero() {
		workspace.UpdatedAt = document.Issues[0].UpdatedAt
	}
	for _, issue := range document.Issues[1:] {
		if issue.CreatedAt.Before(workspace.CreatedAt) {
			workspace.CreatedAt = issue.CreatedAt
		}
		if issue.UpdatedAt.After(workspace.UpdatedAt) {
			workspace.UpdatedAt = issue.UpdatedAt
		}
	}
	return workspace
}

type promptCandidate struct {
	issueID string
	step    native.StepKind
	run     captaindb.PromptRun
}

func (service *Service) resolveCaptain(
	ctx context.Context,
	tx *gorm.DB,
	captain *captaindb.DB,
	document Document,
	planRoot string,
) (resolvedCaptain, error) {
	resolved := resolvedCaptain{activeRuns: map[string]uuid.UUID{}, selectedPlans: map[string]uuid.UUID{}}

	sessionCache := map[string]*captaindb.Session{}
	resolveSession := func(issueID, identity string) (*captaindb.Session, bool, error) {
		identity = strings.TrimSpace(identity)
		if session, ok := sessionCache[identity]; ok {
			return session, true, nil
		}
		session, err := captain.GetSessionByIdentity(ctx, identity, "", "", "")
		if err != nil {
			if errors.Is(err, captaindb.ErrSessionNotFound) || errors.Is(err, captaindb.ErrSessionConflict) {
				resolved.warnings = append(resolved.warnings, Warning{
					Code: "captain_session_unresolved", IssueID: issueID,
					Message: fmt.Sprintf("Captain session %q could not be resolved: %v", identity, err),
				})
				return nil, false, nil
			}
			return nil, false, err
		}
		sessionCache[identity] = session
		return session, true, nil
	}

	// A session can appear more than once in label history. The latest explicit
	// mode wins; conflicting history is surfaced rather than attaching one run
	// as two different steps.
	type sessionKey struct{ issue, identity string }
	latestHints := map[sessionKey]SessionHint{}
	hintModes := map[sessionKey]map[string]bool{}
	for _, hint := range document.Sessions {
		key := sessionKey{issue: hint.IssueSourceID, identity: hint.Identity}
		if hintModes[key] == nil {
			hintModes[key] = map[string]bool{}
		}
		hintModes[key][strings.TrimSpace(hint.Mode)] = true
		if current, ok := latestHints[key]; !ok || hint.ObservedAt.After(current.ObservedAt) ||
			(hint.ObservedAt.Equal(current.ObservedAt) && hint.Mode > current.Mode) {
			latestHints[key] = hint
		}
	}
	var hintKeys []sessionKey
	for key := range latestHints {
		hintKeys = append(hintKeys, key)
	}
	sort.Slice(hintKeys, func(i, j int) bool {
		if hintKeys[i].issue != hintKeys[j].issue {
			return hintKeys[i].issue < hintKeys[j].issue
		}
		return hintKeys[i].identity < hintKeys[j].identity
	})
	candidatesByRun := map[uuid.UUID][]promptCandidate{}
	for _, key := range hintKeys {
		hint := latestHints[key]
		if len(hintModes[key]) > 1 {
			var modes []string
			for mode := range hintModes[key] {
				if mode == "" {
					mode = "<missing>"
				}
				modes = append(modes, mode)
			}
			sort.Strings(modes)
			resolved.warnings = append(resolved.warnings, Warning{
				Code: "captain_session_mode_conflict", IssueID: hint.IssueSourceID,
				Message: fmt.Sprintf("Captain session %q has conflicting historical modes: %s; prompt runs were not linked", hint.Identity, strings.Join(modes, ", ")),
			})
			continue
		}
		step, ok := stepForMode(hint.Mode)
		if !ok {
			resolved.warnings = append(resolved.warnings, Warning{
				Code: "captain_step_unknown", IssueID: hint.IssueSourceID,
				Message: fmt.Sprintf("Captain session %q has no explicit plan/run/verify mode", hint.Identity),
			})
			continue
		}
		session, ok, err := resolveSession(hint.IssueSourceID, hint.Identity)
		if err != nil {
			return resolved, err
		}
		if !ok {
			continue
		}
		runs, err := captain.ListPromptRuns(ctx, captaindb.PromptRunFilter{SessionID: &session.ID})
		if err != nil {
			return resolved, err
		}
		if len(runs) == 0 {
			resolved.warnings = append(resolved.warnings, Warning{
				Code: "captain_prompt_runs_missing", IssueID: hint.IssueSourceID,
				Message: fmt.Sprintf("Captain session %q has no prompt runs", hint.Identity),
			})
			continue
		}
		for _, run := range runs {
			candidatesByRun[run.ID] = append(candidatesByRun[run.ID], promptCandidate{issueID: hint.IssueSourceID, step: step, run: run})
		}
	}

	byIssueStep := map[string][]promptCandidate{}
	for runID, candidates := range candidatesByRun {
		ownerSet := map[string]bool{}
		stepSet := map[native.StepKind]bool{}
		for _, candidate := range candidates {
			ownerSet[candidate.issueID] = true
			stepSet[candidate.step] = true
		}
		if len(ownerSet) != 1 || len(stepSet) != 1 {
			for issueID := range ownerSet {
				resolved.warnings = append(resolved.warnings, Warning{
					Code: "captain_prompt_run_conflict", IssueID: issueID,
					Message: fmt.Sprintf("Captain prompt run %s is referenced by multiple issues or step kinds", runID),
				})
			}
			continue
		}
		candidate := candidates[0]
		owned, err := promptRunOwnedByOtherIssue(tx, runID, candidate.issueID)
		if err != nil {
			return resolved, err
		}
		if owned {
			resolved.warnings = append(resolved.warnings, Warning{
				Code: "captain_prompt_run_owned", IssueID: candidate.issueID,
				Message: fmt.Sprintf("Captain prompt run %s is already linked to another native issue", runID),
			})
			continue
		}
		key := candidate.issueID + "\x00" + string(candidate.step)
		byIssueStep[key] = append(byIssueStep[key], candidate)
	}
	for _, candidates := range byIssueStep {
		sort.Slice(candidates, func(i, j int) bool {
			if !candidates[i].run.CreatedAt.Equal(candidates[j].run.CreatedAt) {
				return candidates[i].run.CreatedAt.Before(candidates[j].run.CreatedAt)
			}
			return candidates[i].run.ID.String() < candidates[j].run.ID.String()
		})
		for ordinal, candidate := range candidates {
			resolved.promptLinks = append(resolved.promptLinks, native.ImportPromptRunLink{
				IssueSourceID: candidate.issueID,
				PromptRunID:   candidate.run.ID,
				StepKind:      candidate.step,
				Ordinal:       ordinal,
				CreatedAt:     candidate.run.CreatedAt,
			})
		}
	}
	// Keep every candidate until legacy status has selected the newest qualifying
	// run. A newer terminal run must not hide an older live/waiting run that still
	// explains an in_progress or ask snapshot.
	candidateRuns := map[string][]promptCandidate{}
	seenCandidateRuns := map[string]map[uuid.UUID]bool{}
	for _, candidates := range byIssueStep {
		for _, candidate := range candidates {
			if seenCandidateRuns[candidate.issueID] == nil {
				seenCandidateRuns[candidate.issueID] = map[uuid.UUID]bool{}
			}
			if !seenCandidateRuns[candidate.issueID][candidate.run.ID] {
				seenCandidateRuns[candidate.issueID][candidate.run.ID] = true
				candidateRuns[candidate.issueID] = append(candidateRuns[candidate.issueID], candidate)
			}
		}
	}
	activeRuns, executionWarnings, err := selectLegacyActiveRuns(tx, document.Issues, candidateRuns)
	if err != nil {
		return resolved, err
	}
	resolved.activeRuns = activeRuns
	resolved.warnings = append(resolved.warnings, executionWarnings...)
	sort.Slice(resolved.promptLinks, func(i, j int) bool {
		left, right := resolved.promptLinks[i], resolved.promptLinks[j]
		if left.IssueSourceID != right.IssueSourceID {
			return left.IssueSourceID < right.IssueSourceID
		}
		if left.StepKind != right.StepKind {
			return left.StepKind < right.StepKind
		}
		if left.Ordinal != right.Ordinal {
			return left.Ordinal < right.Ordinal
		}
		return left.PromptRunID.String() < right.PromptRunID.String()
	})

	planOrdinals := map[string]int{}
	for _, hint := range document.Plans {
		if hint.Clear {
			if hint.Selected {
				delete(resolved.selectedPlans, hint.IssueSourceID)
			}
			continue
		}
		// Ordinals describe normalized historical observation order, not the
		// subset that happened to resolve during this pass. Keeping gaps makes a
		// later repair of an older unresolved plan replay without renumbering the
		// links that were already durable.
		ordinal := planOrdinals[hint.IssueSourceID]
		planOrdinals[hint.IssueSourceID] = ordinal + 1
		var content []byte
		var readErr error
		if !hint.Pathless {
			path, pathErr := planFilePath(planRoot, hint.Path)
			readErr = pathErr
			if readErr == nil {
				content, readErr = service.readFile(path)
			}
			if readErr != nil {
				resolved.warnings = append(resolved.warnings, Warning{
					Code: "plan_file_missing", IssueID: hint.IssueSourceID,
					Message: fmt.Sprintf("legacy plan file %q could not be read: %v", hint.Path, readErr),
				})
			}
		}
		if strings.TrimSpace(hint.SessionIdentity) == "" {
			resolved.warnings = append(resolved.warnings, Warning{
				Code: "plan_session_missing", IssueID: hint.IssueSourceID,
				Message: fmt.Sprintf("legacy plan %q has no Captain session reference", hint.Path),
			})
			continue
		}
		session, ok, err := resolveSession(hint.IssueSourceID, hint.SessionIdentity)
		if err != nil {
			return resolved, err
		}
		if !ok {
			continue
		}
		plans, err := captain.ListPlans(ctx, captaindb.PlanFilter{SourceSessionID: &session.ID})
		if err != nil {
			return resolved, err
		}
		matching := plans
		if !hint.Pathless {
			matching = matchingPlans(plans, hint.Path)
		}
		if len(matching) == 0 {
			description := fmt.Sprintf("legacy plan %q", hint.Path)
			if hint.Pathless {
				description = "pathless legacy plan"
			}
			resolved.warnings = append(resolved.warnings, Warning{
				Code: "captain_plan_unresolved", IssueID: hint.IssueSourceID,
				Message: description + " has no authoritative Captain plan row",
			})
			continue
		}
		if len(matching) > 1 {
			description := fmt.Sprintf("legacy plan %q", hint.Path)
			if hint.Pathless {
				description = "pathless legacy plan"
			}
			resolved.warnings = append(resolved.warnings, Warning{
				Code: "captain_plan_ambiguous", IssueID: hint.IssueSourceID,
				Message: fmt.Sprintf("%s matches %d Captain plan rows; none were linked", description, len(matching)),
			})
			continue
		}
		plan := matching[0]
		if hint.Pathless {
			if plan.LatestRevision == nil {
				resolved.warnings = append(resolved.warnings, Warning{
					Code: "captain_plan_revision_missing", IssueID: hint.IssueSourceID,
					Message: "pathless legacy plan has no durable Captain revision; it was not linked",
				})
				continue
			}
		} else if readErr == nil {
			normalizedContent := strings.TrimSpace(strings.ReplaceAll(string(content), "\r\n", "\n"))
			if normalizedContent == "" {
				resolved.warnings = append(resolved.warnings, Warning{
					Code: "plan_content_empty", IssueID: hint.IssueSourceID,
					Message: fmt.Sprintf("legacy plan file %q is empty", hint.Path),
				})
				if plan.LatestRevision == nil {
					// A readable empty legacy file cannot make a revisionless
					// Captain plan safe to attach or select.
					continue
				}
			} else {
				sum := sha256.Sum256([]byte(normalizedContent))
				resolved.planRevisions = append(resolved.planRevisions, pendingPlanRevision{
					IssueSourceID: hint.IssueSourceID, PlanID: plan.ID,
					Content: normalizedContent, ContentHash: hex.EncodeToString(sum[:]),
				})
			}
		} else if plan.LatestRevision == nil {
			// Do not create a durable link to an empty/dead plan row.
			continue
		}
		resolved.planLinks = append(resolved.planLinks, native.ImportPlanLink{
			IssueSourceID: hint.IssueSourceID, PlanID: plan.ID, Ordinal: ordinal, CreatedAt: hint.ObservedAt,
		})
		if hint.Selected {
			resolved.selectedPlans[hint.IssueSourceID] = plan.ID
		}
	}
	return resolved, nil
}

func buildNativeBatch(document Document, workspace native.ImportWorkspace, resolved resolvedCaptain, warnings []Warning) native.ImportBatch {
	batch := native.ImportBatch{Source: native.DefaultImportSource, Workspace: workspace}
	for _, issue := range document.Issues {
		input := native.ImportIssue{
			SourceID: issue.SourceID, Title: issue.Title, Body: issue.Body, Verification: issue.Verification,
			Labels: issue.Labels, Priority: issue.Priority, Status: issue.Status, ExecutionState: issue.ExecutionState,
			CreatedAt: issue.CreatedAt, UpdatedAt: issue.UpdatedAt,
		}
		if runID, ok := resolved.activeRuns[issue.SourceID]; ok {
			input.ActivePromptRunID = &runID
		}
		if planID, ok := resolved.selectedPlans[issue.SourceID]; ok {
			input.SelectedPlanID = &planID
		}
		batch.Issues = append(batch.Issues, input)
	}
	for _, event := range document.Events {
		batch.Events = append(batch.Events, native.ImportEvent{
			IssueSourceID: event.IssueSourceID, SourceID: event.SourceID, Order: event.Order, Kind: event.Kind,
			Actor: event.Actor, Body: event.Body, Payload: event.Payload, CreatedAt: event.CreatedAt,
		})
	}
	for _, relationship := range document.Relationships {
		batch.Relationships = append(batch.Relationships, native.ImportRelationship{
			IssueSourceID: relationship.IssueSourceID, TargetIssueSourceID: relationship.TargetSourceID,
			Relation: relationship.Relation, CreatedAt: relationship.CreatedAt,
		})
	}
	for _, relationship := range document.RemovedRelationships {
		batch.RelationshipDeletes = append(batch.RelationshipDeletes, native.ImportRelationship{
			IssueSourceID: relationship.IssueSourceID, TargetIssueSourceID: relationship.TargetSourceID,
			Relation: relationship.Relation, CreatedAt: relationship.CreatedAt,
		})
	}
	batch.PromptRunLinks = resolved.promptLinks
	batch.PlanLinks = resolved.planLinks
	issueTimes := make(map[string]time.Time, len(document.Issues))
	for _, issue := range document.Issues {
		// Warning source IDs are content-addressed across initial/final imports;
		// use the immutable issue creation time so a changed snapshot timestamp
		// cannot turn an exact warning replay into an event conflict.
		issueTimes[issue.SourceID] = issue.CreatedAt
	}
	for _, warning := range warnings {
		payload, _ := json.Marshal(map[string]string{"code": warning.Code})
		batch.Warnings = append(batch.Warnings, native.ImportWarning{
			IssueSourceID: warning.IssueID, SourceID: warning.SourceID(), Code: warning.Code,
			Message: warning.Message, Payload: payload, CreatedAt: issueTimes[warning.IssueID], Order: math.MaxInt,
		})
	}
	return batch
}

func importFingerprint(sourceHash string, resolved resolvedCaptain, warnings []Warning) (string, error) {
	type pointer struct {
		IssueID string    `json:"issueId"`
		ID      uuid.UUID `json:"id"`
	}
	var active, selected []pointer
	for issueID, id := range resolved.activeRuns {
		active = append(active, pointer{IssueID: issueID, ID: id})
	}
	for issueID, id := range resolved.selectedPlans {
		selected = append(selected, pointer{IssueID: issueID, ID: id})
	}
	sort.Slice(active, func(i, j int) bool { return active[i].IssueID < active[j].IssueID })
	sort.Slice(selected, func(i, j int) bool { return selected[i].IssueID < selected[j].IssueID })
	revisions := append([]pendingPlanRevision(nil), resolved.planRevisions...)
	sort.Slice(revisions, func(i, j int) bool {
		if revisions[i].PlanID != revisions[j].PlanID {
			return revisions[i].PlanID.String() < revisions[j].PlanID.String()
		}
		return revisions[i].ContentHash < revisions[j].ContentHash
	})
	warningIDs := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		warningIDs = append(warningIDs, warning.SourceID())
	}
	sort.Strings(warningIDs)
	payload := struct {
		SourceHash  string                       `json:"sourceHash"`
		PromptLinks []native.ImportPromptRunLink `json:"promptLinks"`
		PlanLinks   []native.ImportPlanLink      `json:"planLinks"`
		Active      []pointer                    `json:"active"`
		Selected    []pointer                    `json:"selected"`
		Revisions   []pendingPlanRevision        `json:"revisions"`
		Warnings    []string                     `json:"warnings"`
	}{
		SourceHash: sourceHash, PromptLinks: resolved.promptLinks, PlanLinks: resolved.planLinks,
		Active: active, Selected: selected, Revisions: revisions, Warnings: warningIDs,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func applyPendingPlanRevisions(ctx context.Context, captain *captaindb.DB, resolved *resolvedCaptain) error {
	revisions := append([]pendingPlanRevision(nil), resolved.planRevisions...)
	sort.Slice(revisions, func(i, j int) bool {
		if revisions[i].PlanID != revisions[j].PlanID {
			return revisions[i].PlanID.String() < revisions[j].PlanID.String()
		}
		if revisions[i].ContentHash != revisions[j].ContentHash {
			return revisions[i].ContentHash < revisions[j].ContentHash
		}
		return revisions[i].IssueSourceID < revisions[j].IssueSourceID
	})
	for _, pending := range revisions {
		revision, created, err := captain.AppendPlanRevisionWithResult(ctx, captaindb.AppendPlanRevisionInput{
			PlanID: pending.PlanID, PlanMarkdown: pending.Content, CreatedBy: native.DefaultImportSource,
		})
		if err != nil {
			return err
		}
		if created {
			resolved.appendedRevisions = append(resolved.appendedRevisions, revision.ID)
		}
	}
	return nil
}

func checksumResolvedCaptain(ctx context.Context, captain *captaindb.DB, resolved resolvedCaptain) (string, error) {
	type prompt struct {
		ID        uuid.UUID                `json:"id"`
		SessionID uuid.UUID                `json:"sessionId"`
		Phase     captaindb.PromptRunPhase `json:"phase"`
		State     captaindb.PromptRunState `json:"state"`
		Version   int64                    `json:"version"`
	}
	type plan struct {
		ID        uuid.UUID                `json:"id"`
		SessionID uuid.UUID                `json:"sessionId"`
		Revisions []captaindb.PlanRevision `json:"revisions"`
	}
	var prompts []prompt
	seenRuns := map[uuid.UUID]bool{}
	for _, link := range resolved.promptLinks {
		if seenRuns[link.PromptRunID] {
			continue
		}
		seenRuns[link.PromptRunID] = true
		run, err := captain.GetPromptRun(ctx, link.PromptRunID)
		if err != nil {
			return "", err
		}
		prompts = append(prompts, prompt{ID: run.ID, SessionID: run.SessionID, Phase: run.Phase, State: run.State, Version: run.Version})
	}
	sort.Slice(prompts, func(i, j int) bool { return prompts[i].ID.String() < prompts[j].ID.String() })
	var plans []plan
	seenPlans := map[uuid.UUID]bool{}
	for _, link := range resolved.planLinks {
		if seenPlans[link.PlanID] {
			continue
		}
		seenPlans[link.PlanID] = true
		current, err := captain.GetPlan(ctx, link.PlanID)
		if err != nil {
			return "", err
		}
		revisions, err := captain.ListPlanRevisions(ctx, link.PlanID)
		if err != nil {
			return "", err
		}
		plans = append(plans, plan{ID: current.ID, SessionID: current.SourceSessionID, Revisions: revisions})
	}
	sort.Slice(plans, func(i, j int) bool { return plans[i].ID.String() < plans[j].ID.String() })
	encoded, err := json.Marshal(struct {
		Prompts []prompt `json:"prompts"`
		Plans   []plan   `json:"plans"`
	}{Prompts: prompts, Plans: plans})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func stepForMode(mode string) (native.StepKind, bool) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "plan":
		return native.StepPlan, true
	case "run":
		return native.StepRun, true
	case "verify":
		return native.StepVerify, true
	default:
		return "", false
	}
}

func promptRunLater(left, right captaindb.PromptRun) bool {
	if !left.CreatedAt.Equal(right.CreatedAt) {
		return left.CreatedAt.After(right.CreatedAt)
	}
	return left.ID.String() > right.ID.String()
}

func selectLegacyActiveRuns(
	tx *gorm.DB,
	issues []Issue,
	candidateRuns map[string][]promptCandidate,
) (map[string]uuid.UUID, []Warning, error) {
	active := map[string]uuid.UUID{}
	var warnings []Warning
	for _, issue := range issues {
		candidates := append([]promptCandidate(nil), candidateRuns[issue.SourceID]...)
		sort.Slice(candidates, func(i, j int) bool { return promptRunLater(candidates[i].run, candidates[j].run) })
		switch issue.LegacyStatus {
		case "in_progress":
			var selected *captaindb.PromptRun
			for i := range candidates {
				run := &candidates[i].run
				if run.State == captaindb.PromptRunStatePending || run.State == captaindb.PromptRunStateRunning {
					selected = run
					break
				}
			}
			if selected == nil {
				warnings = append(warnings, Warning{
					Code: "captain_live_run_missing", IssueID: issue.SourceID,
					Message: "Grite in_progress status has no live Captain prompt run; execution remains idle or terminal",
				})
			} else {
				active[issue.SourceID] = selected.ID
			}
		case "ask":
			var selected *captaindb.PromptRun
			for i := range candidates {
				run := &candidates[i].run
				if run.State == captaindb.PromptRunStateWaiting {
					selected = run
					break
				}
				var pending bool
				if err := tx.Raw(`
					SELECT EXISTS(
						SELECT 1 FROM captain_turn_requests
						WHERE state::text = 'pending'
						  AND (
							prompt_run_id = ?
							OR (prompt_run_id IS NULL AND session_id = ?)
						  )
					)`, run.ID, run.SessionID).Scan(&pending).Error; err != nil {
					return nil, nil, err
				}
				if pending {
					selected = run
					break
				}
			}
			if selected == nil {
				warnings = append(warnings, Warning{
					Code: "captain_live_request_missing", IssueID: issue.SourceID,
					Message: "Grite ask status has no pending Captain request; execution remains idle or terminal",
				})
			} else {
				active[issue.SourceID] = selected.ID
			}
		}
	}
	return active, warnings, nil
}

func promptRunOwnedByOtherIssue(tx *gorm.DB, runID uuid.UUID, sourceIssueID string) (bool, error) {
	var rows []struct{ Alias string }
	result := tx.Raw(`
		SELECT COALESCE(alias.alias, '') AS alias
		FROM todo_issue_prompt_runs AS link
		LEFT JOIN todo_issue_aliases AS alias
		  ON alias.issue_id = link.issue_id AND alias.kind = 'grite'
		WHERE link.prompt_run_id = ?`, runID).Scan(&rows)
	if result.Error != nil {
		return false, result.Error
	}
	if len(rows) == 0 {
		return false, nil
	}
	for _, row := range rows {
		if row.Alias == sourceIssueID {
			return false, nil
		}
	}
	return true, nil
}

func matchingPlans(plans []captaindb.Plan, path string) []captaindb.Plan {
	want := filepath.Clean(strings.TrimSpace(path))
	var matching []captaindb.Plan
	for i := range plans {
		if filepath.Clean(strings.TrimSpace(plans[i].Path)) == want {
			matching = append(matching, plans[i])
		}
	}
	sort.Slice(matching, func(i, j int) bool { return matching[i].ID.String() < matching[j].ID.String() })
	return matching
}

func planFilePath(root, path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("legacy plan path is empty")
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return "", errors.New("workspace has no root path for resolving a legacy plan")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	target := path
	if !filepath.IsAbs(target) {
		target = filepath.Join(absoluteRoot, target)
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(absoluteRoot, target)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("legacy plan path %q escapes workspace root %q", path, absoluteRoot)
	}
	return filepath.Clean(target), nil
}

func dedupeWarnings(warnings []Warning) []Warning {
	byID := make(map[string]Warning, len(warnings))
	for _, warning := range warnings {
		if strings.TrimSpace(warning.IssueID) == "" || strings.TrimSpace(warning.Code) == "" || strings.TrimSpace(warning.Message) == "" {
			continue
		}
		byID[warning.SourceID()] = warning
	}
	out := make([]Warning, 0, len(byID))
	for _, warning := range byID {
		out = append(out, warning)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].IssueID != out[j].IssueID {
			return out[i].IssueID < out[j].IssueID
		}
		if out[i].Code != out[j].Code {
			return out[i].Code < out[j].Code
		}
		return out[i].Message < out[j].Message
	})
	return out
}

func sortUUIDs(values []uuid.UUID) {
	sort.Slice(values, func(i, j int) bool { return values[i].String() < values[j].String() })
}
