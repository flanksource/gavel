package outline

import (
	"fmt"

	"github.com/flanksource/gavel/testrunner/parsers"
)

type jestSymbol struct {
	container bool
	pending   bool
	focused   bool
}

type jestSourceParser struct {
	file    string
	tokens  []jsToken
	pairs   map[int]int
	symbols map[string]jestSymbol
}

func parseJestSource(file string, source []byte) ([]*Entry, error) {
	tokens, err := tokenizeJSTestSource(file, source)
	if err != nil {
		return nil, err
	}
	pairs, err := pairJSTokens(file, tokens)
	if err != nil {
		return nil, err
	}
	parser := jestSourceParser{file: file, tokens: tokens, pairs: pairs, symbols: defaultJestSymbols()}
	parser.registerImportAliases()
	return parser.parseScope(0, len(tokens), nil), nil
}

func defaultJestSymbols() map[string]jestSymbol {
	return map[string]jestSymbol{
		"describe":  {container: true},
		"fdescribe": {container: true, focused: true},
		"xdescribe": {container: true, pending: true},
		"test":      {},
		"it":        {},
		"fit":       {focused: true},
		"xit":       {pending: true},
		"xtest":     {pending: true},
	}
}

func (p *jestSourceParser) parseScope(start, end int, suite []string) []*Entry {
	var entries []*Entry
	for i := start; i < end; {
		entry, next, bodyStart, bodyEnd, ok := p.parseCall(i, suite)
		if ok {
			if entry.Container && bodyStart >= 0 {
				childSuite := append(append([]string(nil), suite...), entry.Name)
				entry.Children = p.parseScope(bodyStart, bodyEnd, childSuite)
			}
			entries = append(entries, entry)
			i = next
			continue
		}
		if p.tokens[i].text == "{" {
			if close, found := p.pairs[i]; found {
				i = close + 1
				continue
			}
		}
		i++
	}
	return entries
}

func (p *jestSourceParser) parseCall(index int, suite []string) (*Entry, int, int, int, bool) {
	if index >= len(p.tokens) || p.tokens[index].kind != jsIdentifier {
		return nil, index + 1, -1, -1, false
	}
	symbol, found := p.symbols[p.tokens[index].text]
	if !found {
		return nil, index + 1, -1, -1, false
	}

	dynamicEach := false
	open := index + 1
	for open+1 < len(p.tokens) && (p.tokens[open].text == "." || p.tokens[open].text == "?.") {
		modifier := p.tokens[open+1].text
		open += 2
		switch modifier {
		case "only":
			symbol.focused = true
		case "skip", "todo":
			symbol.pending = true
		case "each":
			dynamicEach = true
			if open < len(p.tokens) && p.tokens[open].text == "(" {
				close, ok := p.pairs[open]
				if !ok {
					return nil, index + 1, -1, -1, false
				}
				open = close + 1
			} else if open < len(p.tokens) && p.tokens[open].kind == jsTemplate {
				open++
			}
		case "concurrent", "failing":
		default:
			return nil, index + 1, -1, -1, false
		}
	}
	if open >= len(p.tokens) || p.tokens[open].text != "(" {
		return nil, index + 1, -1, -1, false
	}
	close, ok := p.pairs[open]
	if !ok {
		return nil, index + 1, -1, -1, false
	}

	name, dynamic := p.callName(open, close)
	if dynamic {
		name = "<dynamic>"
	}
	bodyStart, bodyEnd := p.callbackBody(open, close)
	endLine := p.tokens[close].endLine
	if bodyEnd >= 0 {
		endLine = p.tokens[bodyEnd].endLine
	}
	entry := &Entry{
		Framework: parsers.Jest,
		File:      p.file,
		Line:      p.tokens[index].line,
		EndLine:   endLine,
		Name:      name,
		Suite:     append([]string(nil), suite...),
		Container: symbol.container,
		Dynamic:   dynamic || dynamicEach,
		Pending:   symbol.pending,
		Focused:   symbol.focused,
	}
	if !entry.Container && entry.EndLine >= entry.Line {
		entry.SizeLines = entry.EndLine - entry.Line + 1
	}
	return entry, close + 1, bodyStart, bodyEnd, true
}

func (p *jestSourceParser) callName(open, close int) (string, bool) {
	start := open + 1
	end := p.firstArgumentEnd(start, close)
	if end == start+1 && (p.tokens[start].kind == jsString || p.tokens[start].kind == jsTemplate) {
		return p.tokens[start].text, p.tokens[start].dynamic
	}
	return "", true
}

func (p *jestSourceParser) firstArgumentEnd(start, close int) int {
	for i := start; i < close; i++ {
		if paired, ok := p.pairs[i]; ok && paired > i {
			i = paired
			continue
		}
		if p.tokens[i].text == "," {
			return i
		}
	}
	return close
}

func (p *jestSourceParser) callbackBody(open, close int) (int, int) {
	for i := open + 1; i < close; i++ {
		if p.tokens[i].text == "=>" && i+1 < close && p.tokens[i+1].text == "{" {
			return i + 2, p.pairs[i+1]
		}
		if p.tokens[i].text == "function" {
			for j := i + 1; j < close; j++ {
				if p.tokens[j].text == "{" {
					return j + 1, p.pairs[j]
				}
				if p.tokens[j].text == "," {
					break
				}
			}
		}
	}
	return -1, -1
}

func (p *jestSourceParser) registerImportAliases() {
	for i := 0; i+3 < len(p.tokens); i++ {
		if p.tokens[i].text != "import" || p.tokens[i+1].text != "{" {
			continue
		}
		close, ok := p.pairs[i+1]
		if !ok || close+2 >= len(p.tokens) || p.tokens[close+1].text != "from" || p.tokens[close+2].text != "@jest/globals" {
			continue
		}
		p.registerNamedAliases(i+2, close, "as")
	}
}

func (p *jestSourceParser) registerNamedAliases(start, end int, separator string) {
	for i := start; i < end; {
		original := p.tokens[i].text
		alias := original
		if i+2 < end && p.tokens[i+1].text == separator {
			alias = p.tokens[i+2].text
			i += 3
		} else {
			i++
		}
		if symbol, ok := p.symbols[original]; ok {
			p.symbols[alias] = symbol
		}
		for i < end && p.tokens[i].text != "," {
			i++
		}
		if i < end {
			i++
		}
	}
}

func pairJSTokens(file string, tokens []jsToken) (map[int]int, error) {
	matching := map[string]string{")": "(", "}": "{", "]": "["}
	openers := map[string]bool{"(": true, "{": true, "[": true}
	var stack []int
	pairs := map[int]int{}
	for i, token := range tokens {
		if openers[token.text] {
			stack = append(stack, i)
			continue
		}
		expected, closing := matching[token.text]
		if !closing {
			continue
		}
		if len(stack) == 0 || tokens[stack[len(stack)-1]].text != expected {
			return nil, fmt.Errorf("%s:%d: unmatched %s", file, token.line, token.text)
		}
		open := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		pairs[open] = i
	}
	if len(stack) > 0 {
		open := stack[len(stack)-1]
		return nil, fmt.Errorf("%s:%d: unmatched %s", file, tokens[open].line, tokens[open].text)
	}
	return pairs, nil
}
