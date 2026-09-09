package outline

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func nodeCollectionError(framework parsers.Framework, pkg string, err error) *Entry {
	rel := cleanRelPath(pkg)
	if rel == "" {
		rel = "."
	}
	return &Entry{
		Framework: framework,
		File:      path.Join(rel, "package.json"),
		Name:      "<" + framework.String() + " collection failed>",
		Error:     summarizeCollectionErr(err),
	}
}

func entryFromParsedTest(test parsers.Test, pkg, workDir string) *Entry {
	file := test.File
	if filepath.IsAbs(file) {
		file = relativeTo(file, workDir)
	} else {
		file = filepath.ToSlash(filepath.Clean(file))
		pkg = cleanRelPath(pkg)
		if pkg != "" && file != pkg && !strings.HasPrefix(file, pkg+"/") {
			file = path.Join(pkg, file)
		}
	}
	return &Entry{
		Framework: test.Framework,
		File:      file,
		Line:      test.Line,
		Name:      test.Name,
		Suite:     append([]string(nil), test.Suite...),
	}
}

func summarizeCollectionErr(err error) string {
	clean := ansiPattern.ReplaceAllString(err.Error(), "")
	for _, line := range strings.Split(clean, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Could not resolve") ||
			strings.Contains(line, "Cannot find") ||
			strings.Contains(line, "ERR_MODULE_NOT_FOUND") {
			return line
		}
	}
	first, _, _ := strings.Cut(clean, "\n")
	return strings.TrimSpace(first)
}
