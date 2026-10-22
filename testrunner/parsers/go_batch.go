package parsers

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const ginkgoMetadataPrefix = "GAVEL_GINKGO_V1:"

type goPackageStream struct {
	goOutput   bytes.Buffer
	chunks     map[int]string
	chunkCount int
	cached     bool
	finished   bool
	failed     bool
}

// ParseGoTestBatch separates interleaved go test -json events by package and
// combines native Go results with Ginkgo reports carried in cached stdout.
func ParseGoTestBatch(output io.Reader, workDir string, ginkgoPackages map[string]bool) (map[string]TestSuiteResults, error) {
	streams := map[string]*goPackageStream{}
	reader := bufio.NewReader(output)
	for {
		line, err := reader.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			var event goTestEvent
			if decodeErr := json.Unmarshal(line, &event); decodeErr != nil {
				return nil, fmt.Errorf("decode go test event: %w", decodeErr)
			}
			pkg := event.getPackage()
			if pkg == "" {
				return nil, fmt.Errorf("go test event without package: %s", strings.TrimSpace(string(line)))
			}
			stream := streams[pkg]
			if stream == nil {
				stream = &goPackageStream{chunks: map[int]string{}}
				streams[pkg] = stream
			}
			if event.Action == "output" && strings.HasPrefix(strings.TrimSpace(event.Output), ginkgoMetadataPrefix) {
				if parseErr := stream.addChunk(strings.TrimSpace(event.Output)); parseErr != nil {
					return nil, fmt.Errorf("package %s: %w", pkg, parseErr)
				}
			} else {
				stream.goOutput.Write(line)
			}
			if event.Action == "output" && event.Test == "" && strings.Contains(event.Output, "(cached)") {
				stream.cached = true
			}
			if event.Test == "" && (event.Action == "pass" || event.Action == "fail" || event.Action == "build-fail" || event.Action == "skip") {
				stream.finished = true
				stream.failed = event.Action == "fail" || event.Action == "build-fail"
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	results := make(map[string]TestSuiteResults, len(streams))
	goParser := NewGoTestJSON(workDir)
	ginkgoParser := NewGinkgoJSON(workDir)
	for pkg, stream := range streams {
		goTests, err := goParser.Parse(bytes.NewReader(stream.goOutput.Bytes()))
		if err != nil {
			return nil, fmt.Errorf("parse Go results for %s: %w", pkg, err)
		}
		for i := range goTests {
			goTests[i].Cached = stream.cached
		}
		var suites TestSuiteResults
		if len(goTests) > 0 {
			suites = append(suites, TestResults{Framework: GoTest, Tests: goTests})
		}
		if stream.chunkCount > 0 {
			report, err := stream.report()
			if err != nil {
				return nil, fmt.Errorf("package %s: %w", pkg, err)
			}
			ginkgoTests, err := ginkgoParser.Parse(bytes.NewReader(report))
			if err != nil {
				return nil, fmt.Errorf("parse Ginkgo metadata for %s: %w", pkg, err)
			}
			for i := range ginkgoTests {
				ginkgoTests[i].Cached = stream.cached
			}
			suites = append(suites, TestResults{Framework: Ginkgo, Tests: ginkgoTests})
		} else if ginkgoPackages[pkg] && stream.finished && !stream.failed {
			return nil, fmt.Errorf("package %s passed without Ginkgo metadata", pkg)
		}
		failedTests := 0
		for _, suite := range suites {
			failedTests += suite.Sum().Failed
		}
		if stream.failed && failedTests == 0 {
			suites = append(suites, TestResults{Framework: GoTest, Tests: Tests{{Name: "Package failed", Package: pkg, Failed: true, Message: "go test reported package failure"}}})
		}
		if !stream.finished {
			suites = append(suites, TestResults{Framework: GoTest, Tests: Tests{{Name: "Package incomplete", Package: pkg, Failed: true, Message: "go test did not emit a terminal package event"}}})
		}
		for i := range suites {
			suites[i].Stdout = stream.goOutput.String()
			if stream.failed {
				suites[i].ExitCode = 1
			}
		}
		results[pkg] = suites
	}
	return results, nil
}

func (s *goPackageStream) addChunk(output string) error {
	frame := strings.TrimPrefix(output, ginkgoMetadataPrefix)
	parts := strings.SplitN(frame, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid Ginkgo metadata frame")
	}
	position := strings.SplitN(parts[0], "/", 2)
	if len(position) != 2 {
		return fmt.Errorf("invalid Ginkgo metadata position")
	}
	index, indexErr := strconv.Atoi(position[0])
	count, countErr := strconv.Atoi(position[1])
	if indexErr != nil || countErr != nil || count < 1 || index < 0 || index >= count {
		return fmt.Errorf("invalid Ginkgo metadata chunk %q", parts[0])
	}
	if s.chunkCount != 0 && s.chunkCount != count {
		return fmt.Errorf("inconsistent Ginkgo metadata chunk count")
	}
	if _, exists := s.chunks[index]; exists {
		return fmt.Errorf("duplicate Ginkgo metadata chunk %d", index)
	}
	s.chunkCount = count
	s.chunks[index] = parts[1]
	return nil
}

func (s *goPackageStream) report() ([]byte, error) {
	if len(s.chunks) != s.chunkCount {
		return nil, fmt.Errorf("incomplete Ginkgo metadata: received %d of %d chunks", len(s.chunks), s.chunkCount)
	}
	var encoded strings.Builder
	for i := 0; i < s.chunkCount; i++ {
		encoded.WriteString(s.chunks[i])
	}
	data, err := base64.StdEncoding.DecodeString(encoded.String())
	if err != nil {
		return nil, fmt.Errorf("decode Ginkgo metadata: %w", err)
	}
	return data, nil
}
