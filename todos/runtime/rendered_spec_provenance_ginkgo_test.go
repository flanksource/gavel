package runtime

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/commons-db/shell"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func renderedProfileFixture() (*runtimeprofiles.Resolution, []api.SpecLayer) {
	profile := &runtimeprofiles.Resolution{Profile: runtimeprofiles.Profile{
		ID: "file:project:reviewer", Name: "reviewer", Source: runtimeprofiles.SourceInfo{ID: "project", Label: "project"},
	}}
	trace := []api.SpecLayer{{
		Name: "request", Scope: api.SpecLayerUser, Source: api.SpecLayerSourceRequest,
		Spec: api.Spec{Setup: &shell.Setup{Cwd: "/work/request"}},
	}}
	return profile, trace
}

var _ = Describe("rendered spec", func() {
	It("stores the resolved spec without additional verification or metadata", func() {
		spec := api.Spec{Setup: &shell.Setup{Cwd: "/work/prepared"}}
		rendered, err := renderedSpec(spec)
		Expect(err).NotTo(HaveOccurred())
		encoded, err := json.Marshal(spec)
		Expect(err).NotTo(HaveOccurred())
		var expected map[string]any
		Expect(json.Unmarshal(encoded, &expected)).To(Succeed())
		Expect(rendered).To(Equal(expected))
	})

	It("replaces the admission spec with the prepared execution spec", func() {
		original, err := renderedSpec(api.Spec{Setup: &shell.Setup{Cwd: "/work/request"}})
		Expect(err).NotTo(HaveOccurred())
		refreshed, err := renderedSpec(api.Spec{Setup: &shell.Setup{Cwd: "/work/prepared"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(refreshed).To(HaveLen(1))
		Expect(refreshed["setup"]).To(HaveKeyWithValue("cwd", "/work/prepared"))
		Expect(original["setup"]).To(HaveKeyWithValue("cwd", "/work/request"))
	})
})
