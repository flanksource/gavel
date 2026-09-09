package lifecycle

import (
	"github.com/flanksource/captain/pkg/api"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("lifecycle session provenance", func() {
	g.It("updates both exported and prepared attribution when the host pins a session", func() {
		original := api.FieldProvenance{Source: api.FieldSource{Kind: api.FieldSourceLayer, Name: "request", Key: "/sessionId"}}
		resolution := &Resolution{
			Spec: api.Spec{SessionID: "old-session"}, Provenance: map[string]api.FieldProvenance{"/sessionId": original},
			prepared: &preparedStep{request: api.Spec{SessionID: "old-session"}, provenance: map[string]api.FieldProvenance{"/sessionId": original}},
		}
		resolution.UseSession("continued-session")
		expected := api.FieldProvenance{Source: api.FieldSource{Kind: api.FieldSourceContext, Name: "lifecycle session", Key: "/sessionId"}}
		o.Expect(resolution.Spec.SessionID).To(o.Equal("continued-session"))
		o.Expect(resolution.prepared.request.SessionID).To(o.Equal("continued-session"))
		o.Expect(resolution.Provenance["/sessionId"]).To(o.Equal(expected))
		o.Expect(resolution.prepared.provenance["/sessionId"]).To(o.Equal(expected))
	})
})
