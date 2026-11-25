package github

import (
	"regexp"
	"strings"

	"github.com/flanksource/gavel/pr/model"
)

type actionsLog struct {
	lines  []string
	groups []actionsLogGroup
	steps  []actionsStepLog
}

type actionsLogGroup struct {
	title             string
	start, end, depth int
}

type actionsStepLog struct {
	start, end            int
	firstGroup, lastGroup int
}

type actionsLogScope struct {
	group  int
	action string
	echo   bool
}

type actionsLogParser struct {
	log      actionsLog
	steps    []model.Step
	scopes   []actionsLogScope
	step     int
	nextStep int
}

func attachLogsToSteps(job *model.Job, rawLog string, tailLines int) {
	log := parseActionsLog(rawLog, job.Steps)
	job.Logs, job.LogHeadings = log.tail(0, len(log.lines), log.groups, tailLines, 0)
	for i := range job.Steps {
		step := &job.Steps[i]
		if !model.NeedsJobLogs(step.Conclusion) {
			continue
		}
		section := log.steps[i]
		step.Logs, step.LogHeadings = "", nil
		if section.lastGroup > section.firstGroup {
			step.Logs, step.LogHeadings = log.tail(section.start, section.end, log.groups[section.firstGroup+1:section.lastGroup], tailLines, 1)
		}
	}
}

func parseActionsLog(raw string, steps []model.Step) actionsLog {
	parser := actionsLogParser{log: actionsLog{steps: make([]actionsStepLog, len(steps))}, steps: steps, step: -1}
	if raw != "" {
		for line := range strings.SplitSeq(strings.TrimSuffix(raw, "\n"), "\n") {
			parser.addLine(strings.TrimSuffix(line, "\r"))
		}
	}
	for len(parser.scopes) > 0 {
		parser.closeScope()
	}
	parser.closeStep()
	return parser.log
}

func (p *actionsLogParser) addLine(line string) {
	display, control := prepareLogLine(line)
	switch {
	case strings.HasPrefix(control, "##[group]"):
		title := strings.TrimPrefix(control, "##[group]")
		p.openStep(title)
		p.openScope(title, "", strings.HasPrefix(title, "Run "))
	case control == "##[endgroup]":
		if len(p.scopes) > 0 && p.scopes[len(p.scopes)-1].action == "" {
			p.closeScope()
		}
	case strings.HasPrefix(control, "##[start-action "):
		p.openAction(control)
	case strings.HasPrefix(control, "##[end-action "):
		p.closeAction(control)
	default:
		if len(p.scopes) > 0 && p.scopes[len(p.scopes)-1].echo {
			return
		}
		if len(p.scopes) == 0 && control == "Post job cleanup." {
			p.closeStep()
		}
		p.log.lines = append(p.log.lines, actionsAnnotation.ReplaceAllString(display, ""))
		if len(p.scopes) == 0 && strings.HasPrefix(control, "##[error]Process completed") {
			p.closeStep()
		}
	}
}

func (p *actionsLogParser) openStep(title string) {
	if len(p.scopes) > 0 {
		return
	}
	for i := p.nextStep; i < len(p.steps); i++ {
		if p.steps[i].Name == title {
			p.closeStep()
			p.step, p.nextStep = i, i+1
			p.log.steps[i] = actionsStepLog{start: len(p.log.lines), firstGroup: len(p.log.groups)}
			return
		}
	}
}

func (p *actionsLogParser) closeStep() {
	if p.step < 0 {
		return
	}
	step := &p.log.steps[p.step]
	step.end, step.lastGroup = len(p.log.lines), len(p.log.groups)
	p.log.groups[step.firstGroup].end = step.end
	p.step = -1
}

func (p *actionsLogParser) openScope(title, action string, echo bool) {
	depth := len(p.scopes)
	// The runner closes a step's header group before printing its output.
	if p.step >= 0 {
		root := p.log.steps[p.step].firstGroup
		if len(p.log.groups) > root && (len(p.scopes) == 0 || p.scopes[0].group != root) {
			depth++
		}
	}
	echo = echo || (len(p.scopes) > 0 && p.scopes[len(p.scopes)-1].echo)
	p.scopes = append(p.scopes, actionsLogScope{group: len(p.log.groups), action: action, echo: echo})
	p.log.groups = append(p.log.groups, actionsLogGroup{title: title, start: len(p.log.lines), depth: depth})
}

func (p *actionsLogParser) closeScope() {
	last := len(p.scopes) - 1
	p.log.groups[p.scopes[last].group].end = len(p.log.lines)
	p.scopes = p.scopes[:last]
}

func (p *actionsLogParser) openAction(line string) {
	fields, ok := strings.CutPrefix(strings.TrimSuffix(line, "]"), "##[start-action display=")
	if !ok {
		return
	}
	if i := strings.LastIndex(fields, ";id="); i > 0 && i+len(";id=") < len(fields) {
		p.openScope(fields[:i], fields[i+len(";id="):], false)
	}
}

func (p *actionsLogParser) closeAction(line string) {
	fields, ok := strings.CutPrefix(strings.TrimSuffix(line, "]"), "##[end-action id=")
	if !ok {
		return
	}
	id, _, _ := strings.Cut(fields, ";")
	for i := len(p.scopes) - 1; i >= 0 && id != ""; i-- {
		if p.scopes[i].action == id {
			for len(p.scopes) > i {
				p.closeScope()
			}
			return
		}
	}
}

func (log actionsLog) tail(start, end int, groups []actionsLogGroup, limit, depth int) (string, []model.LogHeading) {
	if limit > 0 {
		start = max(start, end-limit)
	}
	var headings []model.LogHeading
	for _, group := range groups {
		if group.end < start || (group.end == start && group.start < start) {
			continue
		}
		headings = append(headings, model.LogHeading{
			Title: group.title, Line: max(group.start, start) - start,
			EndLine: group.end - start, Depth: group.depth - depth,
		})
	}
	return strings.Join(log.lines[start:end], "\n"), headings
}

var logTimestampPrefix = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T[\d:.]+Z\s?`)
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
var actionsAnnotation = regexp.MustCompile(`##\[(error|warning|debug|notice|command)\]`)

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

func prepareLogLine(line string) (display string, control string) {
	display = logTimestampPrefix.ReplaceAllString(line, "")
	display = ansiEscape.ReplaceAllStringFunc(display, func(sequence string) string {
		if strings.HasSuffix(sequence, "m") {
			return sequence
		}
		return ""
	})
	return display, stripANSI(display)
}
