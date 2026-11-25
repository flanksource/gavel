package github

import (
	"context"
	"errors"
	"fmt"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/pr/model"
	"strings"
	"time"
)

type AuditOptions struct {
	Client     Options
	PR         *model.PRInfo
	StaleAfter time.Duration
	Now        time.Time
}

func Audit(ctx context.Context, opts AuditOptions) ([]models.Violation, error) {
	opts.Client.Context = ctx
	if opts.PR != nil {
		return auditPRCodeQL(opts.Client, opts.PR)
	}
	repo, err := opts.Client.resolveRepo()
	if err != nil {
		return nil, err
	}
	opts.Client.Repo = repo
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	if opts.StaleAfter == 0 {
		opts.StaleAfter = 30 * 24 * time.Hour
	}
	if opts.StaleAfter < 0 {
		return nil, fmt.Errorf("stale-after must be positive")
	}
	prs, err := fetchAuditPRs(opts.Client)
	var failures []error
	if err != nil {
		failures = append(failures, err)
	}
	var found []models.Violation
	for _, pr := range prs {
		if opts.Now.Sub(pr.UpdatedAt) >= opts.StaleAfter {
			found = append(found, auditFinding("pr-stale", models.SeverityWarning, fmt.Sprintf("PR #%d %s: no activity since %s (%s)", pr.Number, pr.Title, pr.UpdatedAt.UTC().Format(time.RFC3339), pr.URL)))
		}
		if pr.Mergeable == "CONFLICTING" {
			found = append(found, auditFinding("pr-conflicting", models.SeverityError, fmt.Sprintf("PR #%d %s conflicts with its base: %s", pr.Number, pr.Title, pr.URL)))
		}
		checks, err := fetchAuditChecks(opts.Client, pr)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		info := &model.PRInfo{NodeID: pr.ID, Number: pr.Number, URL: pr.URL, StatusCheckRollup: checks}
		for _, check := range checks {
			if model.IsFailureConclusion(check.Conclusion) {
				severity, err := auditCheckSeverity(check)
				if err != nil {
					failures = append(failures, err)
					continue
				}
				found = append(found, auditFinding("pr-check-failed", severity, fmt.Sprintf("PR #%d: %s failed: %s", pr.Number, check.Name, check.DetailsURL)))
			}
		}
		findings, err := auditPRCodeQL(opts.Client, info)
		found = append(found, findings...)
		if err != nil {
			failures = append(failures, err)
		}
	}
	for _, kind := range []string{"dependabot", "code-scanning"} {
		findings, err := fetchAuditAlerts(opts.Client, kind)
		found = append(found, findings...)
		if err != nil {
			failures = append(failures, err)
		}
	}
	return found, errors.Join(failures...)
}

func auditFinding(rule string, severity models.ViolationSeverity, message string) models.Violation {
	return models.Violation{Source: "github", Rule: &models.Rule{Method: rule}, Severity: severity, Message: models.StringPtr(message)}
}

func auditCheckSeverity(check model.StatusCheck) (models.ViolationSeverity, error) {
	if check.Required == nil {
		return "", fmt.Errorf("required status unavailable for check %s", check.Name)
	}
	if *check.Required && model.IsFailureConclusion(check.Conclusion) {
		return models.SeverityError, nil
	}
	return models.SeverityWarning, nil
}

func auditPRCodeQL(opts Options, pr *model.PRInfo) ([]models.Violation, error) {
	var found []models.Violation
	var failures []error
	for _, check := range pr.StatusCheckRollup {
		if check.AppSlug != "github-advanced-security" || !strings.EqualFold(check.Name, "CodeQL") {
			continue
		}
		if check.Required == nil {
			var response struct {
				Data struct {
					Node *struct {
						Required *bool `json:"isRequired"`
					}
				}
			}
			err := opts.auditQuery(`query($id:ID!,$pr:ID!){node(id:$id){... on CheckRun{isRequired(pullRequestId:$pr)}}}`, map[string]any{"id": check.NodeID, "pr": pr.NodeID}, &response)
			if err != nil {
				failures = append(failures, err)
				continue
			}
			if response.Data.Node != nil {
				check.Required = response.Data.Node.Required
			}
		}
		severity, err := auditCheckSeverity(check)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		findings, err := FetchCodeQLFindings(opts, check.NodeID)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		for _, v := range findings {
			v.Locations = []models.ViolationLocation{{File: v.File, Line: v.Line, Column: v.Column, EndLine: v.EndLine, Ref: fmt.Sprintf("refs/pull/%d/merge", pr.Number)}}
			v.Severity = severity
			v.Rule = &models.Rule{Method: "codeql-pr"}
			v.Message = models.StringPtr(fmt.Sprintf("PR #%d: %s (%s)", pr.Number, *v.Message, pr.URL))
			found = append(found, v)
		}
	}
	return found, errors.Join(failures...)
}
