package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	locationSHA     = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)
	locationRefName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)
	locationDigits  = regexp.MustCompile(`^[0-9]+$`)
)

// ParseLocation reads a location shortcut:
//
//	path/to/file.sql:42            {file, line}
//	file.sql:42:7                  {file, line, column}
//	<ref>:path/file.xml:42[:7]     {ref, file, line, column}
//	entity://Policy/<id>           {entity: {kind, id}}
//	https://…#css=… | #xpath=… | #testid=… | #role=…[&name=…] | #text=…
//
// No git access is needed to tell a ref from a path. After the trailing line
// (and column) numbers are removed, the text before the first colon is a ref
// when it is a 7–40 digit hex sha, or when it is a git ref name (letters,
// digits, `._/-`, no `..`) and the text after it looks like a path (contains
// `/`, `\` or `.`). A one-letter segment followed by `/` or `\` is a Windows
// drive, never a ref. Anything else is a plain path, so `a:b:3` is the file
// `a:b` at line 3; write the location as an object to say otherwise. A ref
// needs a line: `main:a.xml` is the file `main:a.xml`.
func ParseLocation(shortcut string) (ViolationLocation, error) {
	shortcut = strings.TrimSpace(shortcut)
	var location ViolationLocation
	var err error
	switch {
	case shortcut == "":
		return ViolationLocation{}, fmt.Errorf("location shortcut is empty")
	case strings.HasPrefix(shortcut, "entity://"):
		location, err = parseEntityLocation(shortcut)
	case strings.HasPrefix(shortcut, "http://"), strings.HasPrefix(shortcut, "https://"):
		location, err = parseURLLocation(shortcut)
	default:
		location, err = parseFileLocation(shortcut)
	}
	if err != nil {
		return ViolationLocation{}, err
	}
	if err := location.Validate(); err != nil {
		return ViolationLocation{}, fmt.Errorf("location %q: %w", shortcut, err)
	}
	return location, nil
}

// UnmarshalJSON accepts a ParseLocation shortcut string or the object. The
// object is decoded as written; callers that accept authored locations call
// Validate, so a stored lint result is read back unchanged.
func (l *ViolationLocation) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '"' {
		var shortcut string
		if err := json.Unmarshal(trimmed, &shortcut); err != nil {
			return err
		}
		parsed, err := ParseLocation(shortcut)
		if err != nil {
			return err
		}
		*l = parsed
		return nil
	}
	type plain ViolationLocation
	var decoded plain
	if err := json.Unmarshal(trimmed, &decoded); err != nil {
		return err
	}
	*l = ViolationLocation(decoded)
	return nil
}

func parseEntityLocation(shortcut string) (ViolationLocation, error) {
	kind, id, ok := strings.Cut(strings.TrimPrefix(shortcut, "entity://"), "/")
	if !ok || kind == "" || id == "" {
		return ViolationLocation{}, fmt.Errorf("entity shortcut %q is not entity://<kind>/<id>", shortcut)
	}
	return ViolationLocation{Entity: &EntityTarget{Kind: kind, ID: id}}, nil
}

func parseURLLocation(shortcut string) (ViolationLocation, error) {
	base, fragment, hasFragment := strings.Cut(shortcut, "#")
	parsed, err := url.Parse(base)
	if err != nil {
		return ViolationLocation{}, fmt.Errorf("url shortcut %q: %w", shortcut, err)
	}
	if parsed.Host == "" {
		return ViolationLocation{}, fmt.Errorf("url shortcut %q has no host", shortcut)
	}
	key, value, isSelector := strings.Cut(fragment, "=")
	if !hasFragment || !isSelector {
		return ViolationLocation{URL: &URLTarget{URL: shortcut}}, nil
	}
	target := &URLTarget{URL: base}
	switch key {
	case "css":
		target.CSS = value
	case "xpath":
		target.XPath = value
	case "testid":
		target.TestID = value
	case "text":
		target.Text = value
	case "role":
		target.Role, target.Name, _ = strings.Cut(value, "&name=")
	default:
		return ViolationLocation{}, fmt.Errorf("url shortcut %q: unknown selector %q (want css, xpath, testid, role or text)", shortcut, key)
	}
	return ViolationLocation{URL: target}, nil
}

func parseFileLocation(shortcut string) (ViolationLocation, error) {
	segments := strings.Split(shortcut, ":")
	numbers := 0
	for numbers < 2 && numbers < len(segments)-1 && locationDigits.MatchString(segments[len(segments)-1-numbers]) {
		numbers++
	}
	if numbers == 0 {
		return ViolationLocation{File: shortcut}, nil
	}
	head := strings.Join(segments[:len(segments)-numbers], ":")
	positions := make([]int, numbers)
	for i := range positions {
		n, err := strconv.Atoi(segments[len(segments)-numbers+i])
		if err != nil || n <= 0 {
			return ViolationLocation{}, fmt.Errorf("location %q: line must be positive", shortcut)
		}
		positions[i] = n
	}
	location := ViolationLocation{Line: positions[0]}
	if numbers == 2 {
		location.Column = positions[1]
	}
	location.Ref, location.File = splitLocationRef(head)
	return location, nil
}

func splitLocationRef(head string) (ref, file string) {
	first, rest, ok := strings.Cut(head, ":")
	if !ok || rest == "" {
		return "", head
	}
	if len(first) == 1 && strings.ContainsAny(rest[:1], `/\`) {
		return "", head
	}
	if locationSHA.MatchString(first) {
		return first, rest
	}
	if locationRefName.MatchString(first) && !strings.Contains(first, "..") && strings.ContainsAny(rest, `/\.`) {
		return first, rest
	}
	return "", head
}
