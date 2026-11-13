package editor

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Editor active run payload", func() {
	It("publishes an empty group array before a run has created task groups", func() {
		server := &Server{}
		server.setActive(ActiveRun{PID: 123, Command: "gavel lint"})
		data, err := json.Marshal(server.run.Active[0])
		Expect(err).NotTo(HaveOccurred())
		var wire map[string]json.RawMessage
		Expect(json.Unmarshal(data, &wire)).To(Succeed())
		Expect(string(wire["groups"])).To(Equal("[]"))
	})
})
