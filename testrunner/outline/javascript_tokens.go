package outline

import (
	"fmt"
	"strings"
)

type jsTokenKind uint8

const (
	jsIdentifier jsTokenKind = iota
	jsString
	jsTemplate
	jsPunctuation
	jsOther
)

type jsToken struct {
	kind    jsTokenKind
	text    string
	line    int
	endLine int
	dynamic bool
}

func tokenizeJSTestSource(file string, source []byte) ([]jsToken, error) {
	line := 1
	var tokens []jsToken
	for i := 0; i < len(source); {
		switch {
		case isJSSpace(source[i]):
			if source[i] == '\n' {
				line++
			}
			i++
		case hasJSPrefix(source, i, "//"):
			i = skipJSLineComment(source, i+2)
		case hasJSPrefix(source, i, "/*"):
			startLine := line
			var closed bool
			i, line, closed = skipJSBlockComment(source, i+2, line)
			if !closed {
				return nil, fmt.Errorf("%s:%d: unterminated block comment", file, startLine)
			}
		case isJSIdentifierStart(source[i]):
			start := i
			for i++; i < len(source) && isJSIdentifierPart(source[i]); i++ {
			}
			tokens = append(tokens, jsToken{kind: jsIdentifier, text: string(source[start:i]), line: line, endLine: line})
		case source[i] == '\'' || source[i] == '"':
			token, next, nextLine, err := scanJSString(file, source, i, line)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			i, line = next, nextLine
		case source[i] == '`':
			token, next, nextLine, err := scanJSTemplate(file, source, i, line)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			i, line = next, nextLine
		case source[i] == '/' && startsJSRegex(tokens):
			next, nextLine, err := skipJSRegex(file, source, i, line)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, jsToken{kind: jsOther, text: "<regex>", line: line, endLine: nextLine})
			i, line = next, nextLine
		default:
			startLine := line
			text := string(source[i])
			if i+1 < len(source) && (string(source[i:i+2]) == "=>" || string(source[i:i+2]) == "?.") {
				text = string(source[i : i+2])
				i += 2
			} else {
				i++
			}
			tokens = append(tokens, jsToken{kind: jsPunctuation, text: text, line: startLine, endLine: line})
		}
	}
	return tokens, nil
}

func scanJSString(file string, source []byte, start, line int) (jsToken, int, int, error) {
	quote := source[start]
	startLine := line
	var value strings.Builder
	for i := start + 1; i < len(source); i++ {
		if source[i] == '\n' {
			return jsToken{}, 0, line, fmt.Errorf("%s:%d: unterminated string", file, startLine)
		}
		if source[i] == '\\' {
			if i+1 >= len(source) {
				break
			}
			i++
			value.WriteByte(decodeJSEscape(source[i]))
			continue
		}
		if source[i] == quote {
			return jsToken{kind: jsString, text: value.String(), line: startLine, endLine: line}, i + 1, line, nil
		}
		value.WriteByte(source[i])
	}
	return jsToken{}, 0, line, fmt.Errorf("%s:%d: unterminated string", file, startLine)
}

func scanJSTemplate(file string, source []byte, start, line int) (jsToken, int, int, error) {
	startLine := line
	dynamic := false
	var value strings.Builder
	for i := start + 1; i < len(source); i++ {
		if source[i] == '\n' {
			line++
			value.WriteByte(source[i])
			continue
		}
		if source[i] == '\\' {
			if i+1 >= len(source) {
				break
			}
			i++
			value.WriteByte(decodeJSEscape(source[i]))
			continue
		}
		if source[i] == '$' && i+1 < len(source) && source[i+1] == '{' {
			dynamic = true
		}
		if source[i] == '`' {
			return jsToken{kind: jsTemplate, text: value.String(), line: startLine, endLine: line, dynamic: dynamic}, i + 1, line, nil
		}
		value.WriteByte(source[i])
	}
	return jsToken{}, 0, line, fmt.Errorf("%s:%d: unterminated template literal", file, startLine)
}

func skipJSRegex(file string, source []byte, start, line int) (int, int, error) {
	startLine := line
	inClass := false
	for i := start + 1; i < len(source); i++ {
		switch source[i] {
		case '\n':
			return 0, line, fmt.Errorf("%s:%d: unterminated regular expression", file, startLine)
		case '\\':
			i++
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '/':
			if !inClass {
				for i++; i < len(source) && isJSIdentifierPart(source[i]); i++ {
				}
				return i, line, nil
			}
		}
	}
	return 0, line, fmt.Errorf("%s:%d: unterminated regular expression", file, startLine)
}

func startsJSRegex(tokens []jsToken) bool {
	if len(tokens) == 0 {
		return true
	}
	last := tokens[len(tokens)-1]
	if last.kind == jsIdentifier {
		switch last.text {
		case "await", "case", "delete", "in", "instanceof", "of", "return", "throw", "typeof", "void", "yield":
			return true
		default:
			return false
		}
	}
	if last.kind == jsString || last.kind == jsTemplate || last.kind == jsOther {
		return false
	}
	switch last.text {
	case ")", "]", "}":
		return false
	default:
		return true
	}
}

func skipJSLineComment(source []byte, i int) int {
	for i < len(source) && source[i] != '\n' {
		i++
	}
	return i
}

func skipJSBlockComment(source []byte, i, line int) (int, int, bool) {
	for i < len(source) {
		if source[i] == '\n' {
			line++
		}
		if hasJSPrefix(source, i, "*/") {
			return i + 2, line, true
		}
		i++
	}
	return i, line, false
}

func decodeJSEscape(value byte) byte {
	switch value {
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	default:
		return value
	}
}

func hasJSPrefix(source []byte, offset int, prefix string) bool {
	return offset+len(prefix) <= len(source) && string(source[offset:offset+len(prefix)]) == prefix
}

func isJSSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\r' || value == '\n'
}

func isJSIdentifierStart(value byte) bool {
	return value == '_' || value == '$' || value >= 0x80 || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}

func isJSIdentifierPart(value byte) bool {
	return isJSIdentifierStart(value) || value >= '0' && value <= '9'
}
