package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/flanksource/clicky/api"
)

type ViolationSeverity string

const (
	SeverityInfo    ViolationSeverity = "info"
	SeverityWarning ViolationSeverity = "warning"
	SeverityError   ViolationSeverity = "error"
)

type Violation struct {
	File             string            `json:"file,omitempty"`
	Line             int               `json:"line,omitempty"`
	Column           int               `json:"column,omitempty"`
	EndLine          int               `json:"end_line,omitempty"`
	EndColumn        int               `json:"end_column,omitempty"`
	Message          *string           `json:"message,omitempty"`
	Source           string            `json:"source,omitempty"`
	Rule             *Rule             `json:"rule,omitempty"`
	Severity         ViolationSeverity `json:"severity,omitempty" yaml:"severity,omitempty"`
	Fixable          bool              `json:"fixable,omitempty"`
	FixApplicability string            `json:"fix_applicability,omitempty"`
	Code             *string           `json:"code,omitempty"`
	Excerpt          *SourceExcerpt    `json:"excerpt,omitempty"`
	Anchor           *ViolationAnchor  `json:"anchor,omitempty"`
	// Locations lists every place one fix resolves, File and Line naming the
	// first; a violation found at one place leaves it empty. The fingerprint
	// ignores it, so the message must not name a single place.
	Locations []ViolationLocation `json:"locations,omitempty"`
	// Fingerprint identifies the violation across runs without its line; see
	// linters.AssignFingerprints.
	Fingerprint string `json:"fingerprint,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`
}

// ViolationLocation is one place a violation applies, Message saying what is
// there (e.g. the statement a plan finding came from). The place is one
// target: a file (optionally at a git Ref), an entity, a URL or an image
// region. In JSON it is either that object or a ParseLocation shortcut.
type ViolationLocation struct {
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	EndLine int    `json:"end_line,omitempty"`
	// Ref is the git ref (branch, tag or sha) File is read at.
	Ref     string        `json:"ref,omitempty"`
	Entity  *EntityTarget `json:"entity,omitempty"`
	URL     *URLTarget    `json:"url,omitempty"`
	Image   *ImageTarget  `json:"image,omitempty"`
	Message string        `json:"message,omitempty"`
}

// EntityTarget is a domain record a host can link, its Kind free-form
// (Policy, Activity, AsActivity).
type EntityTarget struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// URLTarget is a page, and optionally one element on it named by at most one
// selector strategy: CSS, XPath, TestID, Role (with an optional accessible
// Name) or Text.
type URLTarget struct {
	URL    string `json:"url"`
	CSS    string `json:"css,omitempty"`
	XPath  string `json:"xpath,omitempty"`
	TestID string `json:"testId,omitempty"`
	Role   string `json:"role,omitempty"`
	Name   string `json:"name,omitempty"`
	Text   string `json:"text,omitempty"`
}

// ImageTarget is a region of a screenshot or artefact; Src is a path or URL.
type ImageTarget struct {
	Src    string `json:"src"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

func (l ViolationLocation) String() string {
	switch {
	case l.Entity != nil:
		return "entity://" + l.Entity.Kind + "/" + l.Entity.ID
	case l.URL != nil:
		return l.URL.String()
	case l.Image != nil:
		return fmt.Sprintf("%s#%d,%d,%dx%d", l.Image.Src, l.Image.X, l.Image.Y, l.Image.Width, l.Image.Height)
	}
	file := l.File
	if l.Ref != "" {
		file = l.Ref + ":" + file
	}
	if l.Line > 0 {
		return fmt.Sprintf("%s:%d", file, l.Line)
	}
	return file
}

func (u URLTarget) String() string {
	for _, s := range u.strategies() {
		if s.value == "" {
			continue
		}
		out := u.URL + "#" + s.key + "=" + s.value
		if s.key == "role" && u.Name != "" {
			out += "&name=" + u.Name
		}
		return out
	}
	return u.URL
}

type urlStrategy struct{ key, value string }

func (u URLTarget) strategies() []urlStrategy {
	return []urlStrategy{{"css", u.CSS}, {"xpath", u.XPath}, {"testid", u.TestID}, {"role", u.Role}, {"text", u.Text}}
}

// Validate refuses a location that names no target or several, a URL with
// two selector strategies, and line/column fields without a file.
func (l ViolationLocation) Validate() error {
	var targets []string
	for i, set := range []bool{l.File != "", l.Entity != nil, l.URL != nil, l.Image != nil} {
		if set {
			targets = append(targets, []string{"file", "entity", "url", "image"}[i])
		}
	}
	switch len(targets) {
	case 0:
		return errors.New("location names no target: set file, entity, url or image")
	case 1:
	default:
		return fmt.Errorf("location names %s: a location is one target", strings.Join(targets, " and "))
	}
	if l.File == "" && (l.Ref != "" || l.Line != 0 || l.Column != 0 || l.EndLine != 0) {
		return errors.New("location ref, line, column and end_line need a file")
	}
	if l.Line < 0 || l.Column < 0 || l.EndLine < 0 || (l.Column > 0 && l.Line == 0) || (l.EndLine > 0 && l.EndLine < l.Line) {
		return fmt.Errorf("location %s: line %d, column %d, end_line %d are not a valid range", l.File, l.Line, l.Column, l.EndLine)
	}
	switch {
	case l.Entity != nil:
		return l.Entity.validate()
	case l.URL != nil:
		return l.URL.validate()
	case l.Image != nil:
		return l.Image.validate()
	}
	return nil
}

func (e EntityTarget) validate() error {
	if e.Kind == "" || e.ID == "" {
		return fmt.Errorf("entity target needs a kind and an id, got kind %q id %q", e.Kind, e.ID)
	}
	return nil
}

func (u URLTarget) validate() error {
	if parsed, err := url.Parse(u.URL); err != nil || !parsed.IsAbs() || parsed.Host == "" {
		return fmt.Errorf("url target %q is not an absolute URL", u.URL)
	}
	var set []string
	for _, s := range u.strategies() {
		if s.value != "" {
			set = append(set, s.key)
		}
	}
	if len(set) > 1 {
		return fmt.Errorf("url target %s sets %s: use at most one selector strategy", u.URL, strings.Join(set, " and "))
	}
	if u.Name != "" && u.Role == "" {
		return fmt.Errorf("url target %s sets name without role", u.URL)
	}
	return nil
}

func (i ImageTarget) validate() error {
	if i.Src == "" {
		return errors.New("image target needs a src")
	}
	if i.X < 0 || i.Y < 0 || i.Width <= 0 || i.Height <= 0 {
		return fmt.Errorf("image target %s needs a region with x, y >= 0 and width, height > 0, got %d,%d %dx%d", i.Src, i.X, i.Y, i.Width, i.Height)
	}
	return nil
}

// ViolationAnchor names what a violation is about independently of where it
// sits: a uir symbol id, the declaration witness of a syntax-only finding, or
// the heading path of a markdown section.
type ViolationAnchor struct {
	Symbol      string `json:"symbol,omitempty"`
	Declaration string `json:"declaration,omitempty"`
	Section     string `json:"section,omitempty"`
}

func (v Violation) Pretty() api.Text {
	t := api.Text{}.Append(v.File).Append(":").Append(v.Line)
	if len(v.Locations) > 1 {
		t = t.Append(fmt.Sprintf(" (%d locations)", len(v.Locations)), "text-muted")
	}
	return t
}

type AnalysisResult struct {
	Violations []Violation `json:"violations,omitempty"`
	FileCount  int         `json:"file_count,omitempty"`
	RuleCount  int         `json:"rule_count,omitempty"`
}

func StringPtr(s string) *string {
	return &s
}
