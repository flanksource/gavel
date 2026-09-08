package github

import (
	"archive/zip"
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("artifact extraction", func() {
	It("reconstructs a legacy crash envelope from its uploaded log", func() {
		logTail := "pre-build failed; to update it:\n\tgo mod tidy\n\x1b[31mfailed\x1b[0m\n"
		malformed := "{\"error\":\"gavel exited 1 before writing results\",\"exit_code\":1,\"log_tail\":\"" + logTail + "\"}\n"

		result, err := extractJSONFromZip(ginkgoArtifactZip(map[string]string{
			"gavel-results.json": malformed,
			"gavel.log":          logTail,
		}))
		Expect(err).NotTo(HaveOccurred())

		var envelope struct {
			Error    string `json:"error"`
			ExitCode int    `json:"exit_code"`
			LogTail  string `json:"log_tail"`
		}
		Expect(json.Unmarshal(result, &envelope)).To(Succeed())
		Expect(envelope).To(Equal(struct {
			Error    string `json:"error"`
			ExitCode int    `json:"exit_code"`
			LogTail  string `json:"log_tail"`
		}{
			Error:    "gavel exited 1 before writing results",
			ExitCode: 1,
			LogTail:  logTail,
		}))
	})

	It("leaves unrelated malformed result JSON invalid", func() {
		malformed := []byte("{\"tests\":[{\"name\":\"failed \x1b[31mtest\"}]}")
		result, err := extractJSONFromZip(ginkgoArtifactZip(map[string]string{
			"gavel-results.json": string(malformed),
			"gavel.log":          "unrelated log\n",
		}))

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(malformed))
		Expect(json.Valid(result)).To(BeFalse())
	})
})

func ginkgoArtifactZip(files map[string]string) []byte {
	GinkgoHelper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		entry, err := w.Create(name)
		Expect(err).NotTo(HaveOccurred())
		_, err = entry.Write([]byte(content))
		Expect(err).NotTo(HaveOccurred())
	}
	Expect(w.Close()).To(Succeed())
	return buf.Bytes()
}
