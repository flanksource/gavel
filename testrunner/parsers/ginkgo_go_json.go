package parsers

import (
	"fmt"
	"io"

	"github.com/flanksource/clicky/task"
)

type GinkgoGoJSON struct {
	workDir string
}

func NewGinkgoGoJSON(workDir string) *GinkgoGoJSON {
	return &GinkgoGoJSON{workDir: workDir}
}

func (p *GinkgoGoJSON) Name() string { return "ginkgo go test json" }

func (p *GinkgoGoJSON) Parse(output io.Reader) ([]Test, error) {
	results, err := ParseGoTestBatch(output, p.workDir, nil)
	if err != nil {
		return nil, err
	}
	var tests []Test
	for _, suites := range results {
		for _, suite := range suites {
			if suite.Framework == Ginkgo {
				tests = append(tests, suite.Tests...)
			}
		}
	}
	return tests, nil
}

func (p *GinkgoGoJSON) ParseStream(output io.Reader, stdout io.Writer, t *task.Task) (int, int, error) {
	return 0, 0, fmt.Errorf("ginkgo go test JSON requires a complete batch report")
}
