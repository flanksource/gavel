package tsc

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTSRS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "tsrs")
}

var _ = Describe("tsrs build diagnostics", func() {
	It("reads type errors with source positions and chained messages from the compiler API", func() {
		var input, output bytes.Buffer
		for _, response := range []string{
			`{"jsonrpc":"2.0","id":1,"result":{}}`,
			`{"jsonrpc":"2.0","id":2,"result":{"buildOrchestratorID":13}}`,
			`{"jsonrpc":"2.0","id":3,"result":{"status":2,"diagnostics":[` +
				`{"fileName":"src/example.ts","startPosition":{"line":2,"character":6},` +
				`"code":2322,"category":1,"text":"Incompatible types.","messageChain":[{"text":"The property is missing."}]},` +
				`{"code":5083,"category":1,"text":"Cannot read config."}]}}`,
		} {
			fmt.Fprintf(&output, "Content-Length: %d\r\n\r\n%s", len(response), response)
		}
		client := tsrsClient{input: &input, output: bufio.NewReader(&output)}
		diagnostics, err := client.build([]string{"tsconfig.custom.json"})
		Expect(err).NotTo(HaveOccurred())
		Expect(input.String()).To(And(
			ContainSubstring(`"rootNames":["tsconfig.custom.json","--force","--noEmit"]`),
			ContainSubstring(`"params":{"buildOrchestratorID":13}`),
		))
		Expect(diagnostics).To(HaveLen(2))
		Expect([]TSCDiagnostic{diagnostics[0].toDiagnostic(), diagnostics[1].toDiagnostic()}).To(Equal([]TSCDiagnostic{
			{File: "src/example.ts", Line: 3, Column: 7, Code: 2322, Category: "Error", Message: "Incompatible types.\n  The property is missing."},
			{Line: 1, Column: 1, Code: 5083, Category: "Error", Message: "Cannot read config."},
		}))
	})
})
