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

var _ = Describe("rendered spec provenance", func() {
	It("records profile identity and ordered inputs separately from execution state", func() {
		profile, trace := renderedProfileFixture()
		rendered, err := renderedSpec(renderedSpecOptions{
			Spec: api.Spec{Setup: &shell.Setup{Cwd: "/work/prepared"}}, RuntimeProfile: profile, SpecTrace: trace,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(rendered).To(HaveKeyWithValue("runtimeProfile", HaveKeyWithValue("profile", HaveKeyWithValue("id", profile.Profile.ID))))
		Expect(rendered["runtimeProfile"]).NotTo(HaveKey("resolved"))
		Expect(rendered).To(HaveKeyWithValue("specTrace", HaveExactElements(HaveKeyWithValue("source", "request"))))
		encoded, err := json.Marshal(rendered)
		Expect(err).NotTo(HaveOccurred())
		var execution api.Spec
		Expect(json.Unmarshal(encoded, &execution)).To(Succeed())
		Expect(execution.Setup.Cwd).To(Equal("/work/prepared"))
		Expect(trace[0].Spec.Setup.Cwd).To(Equal("/work/request"))
	})

	It("preserves admission provenance when setup refreshes the executed spec", func() {
		profile, trace := renderedProfileFixture()
		original, err := renderedSpec(renderedSpecOptions{
			Spec: api.Spec{Setup: &shell.Setup{Cwd: "/work/request"}}, RuntimeProfile: profile, SpecTrace: trace,
		})
		Expect(err).NotTo(HaveOccurred())
		refreshed, err := renderedSpec(renderedSpecOptions{
			Spec: api.Spec{Setup: &shell.Setup{Cwd: "/work/prepared"}}, Previous: original,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(refreshed).To(HaveKeyWithValue("runtimeProfile", original["runtimeProfile"]))
		Expect(refreshed).To(HaveKeyWithValue("specTrace", original["specTrace"]))
		Expect(refreshed["setup"]).To(HaveKeyWithValue("cwd", "/work/prepared"))
		Expect(original["setup"]).To(HaveKeyWithValue("cwd", "/work/request"))
	})
})
