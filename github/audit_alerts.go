package github

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/flanksource/gavel/models"
	"net/url"
	"strings"
)

type auditAlert struct {
	Number   int
	State    string
	HTMLURL  string `json:"html_url"`
	Tool     struct{ Name string }
	Rule     struct{ ID, Description, Severity string }
	Instance struct {
		Message  struct{ Text string }
		Ref      string
		Location struct {
			Path        string
			StartLine   int `json:"start_line"`
			EndLine     int `json:"end_line"`
			StartColumn int `json:"start_column"`
			EndColumn   int `json:"end_column"`
		}
	} `json:"most_recent_instance"`
	Dependency struct {
		Manifest string `json:"manifest_path"`
		Package  struct{ Name, Ecosystem string }
	}
	Advisory struct {
		ID      string `json:"ghsa_id"`
		Summary string
	} `json:"security_advisory"`
	Vulnerability struct {
		Range   string                       `json:"vulnerable_version_range"`
		Patched *struct{ Identifier string } `json:"first_patched_version"`
	} `json:"security_vulnerability"`
}

func fetchAuditAlerts(opts Options, kind string) ([]models.Violation, error) {
	path := fmt.Sprintf("/repos/%s/%s/alerts?state=open&per_page=100", opts.Repo, kind)
	if kind == "code-scanning" {
		path += "&tool_name=CodeQL"
	}
	token, err := opts.token()
	if err != nil {
		return nil, err
	}
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	var found []models.Violation
	seen := map[string]bool{}
	for path != "" {
		if seen[path] {
			return found, fmt.Errorf("%s pagination did not advance", kind)
		}
		seen[path] = true
		response, err := newClient(token).R(ctx).Get(githubAPIBase() + path)
		if err != nil {
			return found, err
		}
		body, err := response.AsString()
		if err != nil {
			return found, err
		}
		if !response.IsOK() {
			return found, fmt.Errorf("%s alerts HTTP %d (check security feature and read permissions): %s", kind, response.StatusCode, body)
		}
		var alerts []auditAlert
		if err := json.Unmarshal([]byte(body), &alerts); err != nil {
			return found, err
		}
		if alerts == nil {
			return found, fmt.Errorf("%s alerts returned null", kind)
		}
		for _, alert := range alerts {
			if alert.State != "open" {
				continue
			}
			if kind == "code-scanning" && alert.Tool.Name != "CodeQL" {
				continue
			}
			v, err := alert.violation(kind)
			if err != nil {
				return found, err
			}
			found = append(found, v)
		}
		path, err = auditNextPage(response.Header.Get("Link"))
		if err != nil {
			return found, err
		}
	}
	return found, nil
}

func auditNextPage(link string) (string, error) {
	for _, part := range strings.Split(link, ",") {
		if !strings.Contains(part, `rel="next"`) {
			continue
		}
		start, end := strings.Index(part, "<"), strings.Index(part, ">")
		if start < 0 || end <= start {
			return "", fmt.Errorf("invalid GitHub pagination link")
		}
		next, err := url.Parse(part[start+1 : end])
		if err != nil {
			return "", err
		}
		base, err := url.Parse(githubAPIBase())
		if err != nil {
			return "", err
		}
		if next.Host != base.Host || next.Scheme != base.Scheme {
			return "", fmt.Errorf("GitHub pagination changed origin")
		}
		return next.RequestURI(), nil
	}
	return "", nil
}

func (a auditAlert) violation(kind string) (models.Violation, error) {
	if kind == "dependabot" {
		patched := "no patched version available"
		if a.Vulnerability.Patched != nil {
			patched = "patched in " + a.Vulnerability.Patched.Identifier
		}
		v := auditFinding("dependabot-alert", models.SeverityError, fmt.Sprintf("%s %s: %s; vulnerable %s; %s; %s: %s", a.Dependency.Package.Ecosystem, a.Dependency.Package.Name, a.Advisory.ID, a.Vulnerability.Range, patched, a.Advisory.Summary, a.HTMLURL))
		v.File = a.Dependency.Manifest
		return v, nil
	}
	levels := map[string]models.ViolationSeverity{"error": models.SeverityError, "warning": models.SeverityWarning, "note": models.SeverityInfo}
	severity, ok := levels[a.Rule.Severity]
	if !ok {
		return models.Violation{}, fmt.Errorf("unknown CodeQL severity %q", a.Rule.Severity)
	}
	v := auditFinding("codeql-alert", severity, fmt.Sprintf("%s [%s]: %s (%s)", a.Rule.Description, a.Rule.ID, a.Instance.Message.Text, a.HTMLURL))
	v.File, v.Line, v.EndLine = a.Instance.Location.Path, a.Instance.Location.StartLine, a.Instance.Location.EndLine
	v.Column, v.EndColumn = a.Instance.Location.StartColumn, a.Instance.Location.EndColumn
	v.Locations = []models.ViolationLocation{{File: v.File, Line: v.Line, Column: v.Column, EndLine: v.EndLine, Ref: a.Instance.Ref}}
	return v, nil
}
