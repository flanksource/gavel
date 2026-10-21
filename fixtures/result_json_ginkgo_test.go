package fixtures

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FixtureResult JSON", func() {
	It("rejects malformed stored attempts with decoding context", func() {
		var result FixtureResult
		err := json.Unmarshal([]byte(`{"metadata":{"attempts":[{"attempt":"first"}]}}`), &result)
		Expect(err).To(MatchError(ContainSubstring("decode stored attempts")))
	})

	It("preserves other metadata while restoring the attempt record", func() {
		var result FixtureResult
		Expect(json.Unmarshal([]byte(`{"name":"probe","metadata":{"custom":{"origin":"worker"},"attempts":[{"attempt":1,"status":"failed","result":{"name":"probe","stdout":"refused"}}]}}`), &result)).To(Succeed())
		Expect(result.Metadata["custom"]).To(Equal(map[string]any{"origin": "worker"}))
		Expect(result.Metadata["attempts"]).To(Equal([]AttemptReport{{Attempt: 1, Status: AttemptFailed, Result: &FixtureResult{Name: "probe", Stdout: "refused"}}}))
	})
})
