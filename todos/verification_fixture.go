package todos

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// verificationFixtureSection is the "## Verification" section that holds the
// TODO's executable test fixtures (fixtures.FixtureNode trees), parsed by
// ParseTODO/ParseTODOContent into TODO.Verification. It is intentionally
// distinct from "## Verification Result" (see verification.go), which holds an
// AI-scored verdict rather than the fixture source itself.
const verificationFixtureSection = "Verification"

// verificationFixtureInsertBefore keeps a newly-added Verification fixture
// section above the operational history sections, mirroring criteriaInsertBefore.
var verificationFixtureInsertBefore = []string{
	"Custom Validations", "Latest Failure", "Verification Result",
	"Attempts", "Failure History",
}

type verificationSectionRange struct {
	start        int
	contentStart int
	end          int
}

// SplitVerificationFixture removes top-level H1 and H2 Verification sections
// from body and returns their markdown in document order.
func SplitVerificationFixture(body string) (cleanBody, verification string, found bool) {
	// A matching heading's text is compared against verificationFixtureSection,
	// so a body that never contains that word cannot hold one — whatever heading
	// syntax it uses. Checking first keeps the markdown parse off the common
	// path: every todo body is split when a backlog is listed, and on a
	// developer database only 42 of 326 bodies mention the word at all.
	if !strings.Contains(strings.ToLower(body), strings.ToLower(verificationFixtureSection)) {
		return body, "", false
	}

	source := []byte(body)
	document := goldmark.New().Parser().Parse(text.NewReader(source))
	headings := documentHeadings(document)
	ranges := make([]verificationSectionRange, 0)
	coveredUntil := -1
	for i, heading := range headings {
		if heading.Level > 2 || !strings.EqualFold(strings.TrimSpace(markdownNodeText(heading, source)), verificationFixtureSection) {
			continue
		}
		start := headingLineStart(source, heading)
		if start < coveredUntil {
			continue
		}
		contentStart := headingContentStart(source, heading)
		frontmatterEnd := fixtureFrontmatterEnd(source, contentStart)
		end := len(source)
		for _, next := range headings[i+1:] {
			nextStart := headingLineStart(source, next)
			if nextStart < frontmatterEnd {
				continue
			}
			if frontmatterEnd < 0 {
				if next.Level > heading.Level {
					continue
				}
			} else if !closesFixtureDocument(source, next) {
				continue
			}
			end = nextStart
			break
		}
		ranges = append(ranges, verificationSectionRange{
			start: start, contentStart: contentStart, end: end,
		})
		coveredUntil = end
	}
	if len(ranges) == 0 {
		return body, "", false
	}

	var bodyBuilder strings.Builder
	parts := make([]string, 0, len(ranges))
	cursor := 0
	for _, section := range ranges {
		bodyBuilder.WriteString(body[cursor:section.start])
		if part := strings.TrimSpace(body[section.contentStart:section.end]); part != "" {
			parts = append(parts, part)
		}
		cursor = section.end
	}
	bodyBuilder.WriteString(body[cursor:])
	return strings.TrimSpace(bodyBuilder.String()), strings.Join(parts, "\n\n"), true
}

// fixtureFrontmatterEnd reports the offset just past the closing "---" of the
// YAML frontmatter opening a Verification section, or -1 when the section does
// not start with frontmatter.
//
// A fixture's closing "---" underlines the YAML above it, so CommonMark reads it
// as a setext H2. Left alone that phantom heading ends the section at the first
// frontmatter key, storing "---" as the whole definition of done.
func fixtureFrontmatterEnd(source []byte, contentStart int) int {
	cursor := contentStart
	for cursor < len(source) {
		end := lineEnd(source, cursor)
		line := strings.TrimSpace(string(source[cursor:end]))
		if line == "" {
			cursor = end
			continue
		}
		if !isFrontmatterDelimiter(line) {
			return -1
		}
		for cursor = end; cursor < len(source); cursor = end {
			end = lineEnd(source, cursor)
			if isFrontmatterDelimiter(strings.TrimSpace(string(source[cursor:end]))) {
				return end
			}
		}
		return -1
	}
	return -1
}

func isFrontmatterDelimiter(line string) bool {
	return len(line) >= 3 && strings.Trim(line, "-") == ""
}

// closesFixtureDocument reports whether heading ends a Verification section that
// declared itself a fixture document with frontmatter. Such a section owns its
// H2 step subsections, so only a top-level heading, another Verification
// section, or one of the operational sections rendered after it can close it.
func closesFixtureDocument(source []byte, heading *ast.Heading) bool {
	if heading.Level <= 1 {
		return true
	}
	title := strings.TrimSpace(markdownNodeText(heading, source))
	if strings.EqualFold(title, verificationFixtureSection) {
		return true
	}
	for _, operational := range verificationFixtureInsertBefore {
		if strings.EqualFold(title, operational) {
			return true
		}
	}
	return false
}

func documentHeadings(document ast.Node) []*ast.Heading {
	headings := make([]*ast.Heading, 0)
	for node := document.FirstChild(); node != nil; node = node.NextSibling() {
		if heading, ok := node.(*ast.Heading); ok {
			headings = append(headings, heading)
		}
	}
	return headings
}

func markdownNodeText(node ast.Node, source []byte) string {
	var builder strings.Builder
	_ = ast.Walk(node, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if textNode, ok := child.(*ast.Text); ok {
				builder.Write(textNode.Segment.Value(source))
			}
		}
		return ast.WalkContinue, nil
	})
	return builder.String()
}

func headingLineStart(source []byte, heading *ast.Heading) int {
	if heading.Lines().Len() == 0 {
		return 0
	}
	start := heading.Lines().At(0).Start
	for start > 0 && source[start-1] != '\n' {
		start--
	}
	return start
}

func headingContentStart(source []byte, heading *ast.Heading) int {
	start := headingLineStart(source, heading)
	end := lineEnd(source, start)
	if strings.HasPrefix(strings.TrimLeft(string(source[start:end]), " \t"), "#") {
		return end
	}
	return lineEnd(source, end)
}

func lineEnd(source []byte, start int) int {
	for start < len(source) {
		start++
		if source[start-1] == '\n' {
			break
		}
	}
	return start
}

// CombineVerificationFixtures joins non-empty verification inputs in priority
// order, separated by one blank line.
func CombineVerificationFixtures(fixtures ...string) string {
	parts := make([]string, 0, len(fixtures))
	for _, fixture := range fixtures {
		if fixture = strings.TrimSpace(fixture); fixture != "" {
			parts = append(parts, fixture)
		}
	}
	return strings.Join(parts, "\n\n")
}

// ExtractVerificationFixture returns the raw markdown inside top-level H1 or
// H2 Verification sections, excluding the heading lines.
func ExtractVerificationFixture(body string) string {
	_, verification, _ := SplitVerificationFixture(body)
	return verification
}

// UpsertVerificationFixture replaces (or inserts) the "## Verification" section
// in body with fixture, the fixture markdown edited via the FixtureEditor.
func UpsertVerificationFixture(body, fixture string) string {
	body, _, _ = SplitVerificationFixture(body)
	var b strings.Builder
	b.WriteString("## " + verificationFixtureSection + "\n\n")
	b.WriteString(strings.TrimSpace(fixture))
	b.WriteString("\n")
	return ReplaceOrAppendSection(body, verificationFixtureSection, b.String(), verificationFixtureInsertBefore...)
}

// UpsertMarkedVerificationBlock inserts or replaces a generated block inside
// the "## Verification" fixture section without clobbering any user-authored
// fixture markdown in the same section.
func UpsertMarkedVerificationBlock(body, marker, block string) string {
	marker = strings.TrimSpace(marker)
	block = strings.TrimSpace(block)
	if marker == "" || block == "" {
		return body
	}

	startTag := "<!-- " + marker + " -->"
	endTag := "<!-- /" + marker + " -->"
	wrapped := startTag + "\n" + block + "\n" + endTag

	fixture := ExtractVerificationFixture(body)
	if fixture == "" {
		return UpsertVerificationFixture(body, wrapped)
	}

	if start := strings.Index(fixture, startTag); start >= 0 {
		searchFrom := start + len(startTag)
		if endRel := strings.Index(fixture[searchFrom:], endTag); endRel >= 0 {
			end := searchFrom + endRel + len(endTag)
			var parts []string
			if before := strings.TrimSpace(fixture[:start]); before != "" {
				parts = append(parts, before)
			}
			parts = append(parts, wrapped)
			if after := strings.TrimSpace(fixture[end:]); after != "" {
				parts = append(parts, after)
			}
			return UpsertVerificationFixture(body, strings.Join(parts, "\n\n"))
		}
	}

	return UpsertVerificationFixture(body, strings.TrimSpace(fixture)+"\n\n"+wrapped)
}
