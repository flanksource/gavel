package fixtures

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture worker errors", func() {
	It("keeps ErrNoFixtures identifiable after crossing the worker protocol", func() {
		dir := GinkgoT().TempDir()
		template := filepath.Join(dir, "template.md")
		Expect(os.WriteFile(template, []byte("# Shared template\n\nNo fixtures here.\n"), 0o600)).To(Succeed())
		input, err := json.Marshal(workerRequest{Options: RunnerOptions{Paths: []string{template}, WorkDir: dir}})
		Expect(err).NotTo(HaveOccurred())

		var output bytes.Buffer
		Expect(RunWorker(bytes.NewReader(input), &output)).To(Succeed())

		var result *workerEvent
		decoder := json.NewDecoder(&output)
		for decoder.More() {
			var event workerEvent
			Expect(decoder.Decode(&event)).To(Succeed())
			if event.Kind == "result" {
				result = &event
			}
		}
		Expect(result).NotTo(BeNil())
		Expect(result.err()).To(MatchError(ErrNoFixtures))
	})
})
