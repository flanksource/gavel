package types

import (
	"fmt"
	"strconv"
	"strings"
)

type PathRef struct {
	File    string
	Line    int // 0 = whole file
	EndLine int // 0 = single line (when Line > 0)
}

func ParsePathRef(s string) PathRef {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 1 {
		return PathRef{File: s}
	}
	ref := PathRef{File: parts[0]}
	lineSpec := parts[1]
	if startStr, endStr, ok := strings.Cut(lineSpec, "-"); ok {
		if start, err := strconv.Atoi(startStr); err == nil {
			ref.Line = start
		}
		if end, err := strconv.Atoi(endStr); err == nil {
			ref.EndLine = end
		}
	} else {
		if line, err := strconv.Atoi(lineSpec); err == nil {
			ref.Line = line
		}
	}
	return ref
}

func (p PathRef) String() string {
	if p.Line == 0 {
		return p.File
	}
	if p.EndLine == 0 {
		return fmt.Sprintf("%s:%d", p.File, p.Line)
	}
	return fmt.Sprintf("%s:%d-%d", p.File, p.Line, p.EndLine)
}

func (p PathRef) IsWholeFile() bool {
	return p.Line == 0
}

func (f TODOFrontmatter) PathRefs() []PathRef {
	refs := make([]PathRef, len(f.Path))
	for i, p := range f.Path {
		refs[i] = ParsePathRef(p)
	}
	return refs
}
