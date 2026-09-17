package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
)

func runtimePresetTestCatalog() (*runtimeprofiles.Catalog, runtimeprofiles.SourceInfo) {
	root := GinkgoT().TempDir()
	presets, err := runtimeprofiles.NewFileSource(runtimeprofiles.FileSourceOptions{
		Kind: runtimeprofiles.KindPreset, Dir: filepath.Join(root, "presets"), Label: "project presets", Implicit: true,
	})
	Expect(err).NotTo(HaveOccurred())
	profiles, err := runtimeprofiles.NewFileSource(runtimeprofiles.FileSourceOptions{
		Kind: runtimeprofiles.KindProfile, Dir: filepath.Join(root, "profiles"), Label: "project profiles", Implicit: true,
	})
	Expect(err).NotTo(HaveOccurred())
	catalog, err := runtimeprofiles.NewCatalog(presets, profiles)
	Expect(err).NotTo(HaveOccurred())
	return catalog, presets.Info()
}

func runtimePresetRequest(method, id string, body any) *http.Request {
	var encoded []byte
	if body != nil {
		var err error
		encoded, err = json.Marshal(body)
		Expect(err).NotTo(HaveOccurred())
	}
	req := httptest.NewRequest(method, "/api/settings/runtime-presets", bytes.NewReader(encoded))
	req.SetPathValue("id", id)
	return req
}

var _ = Describe("runtime preset settings API", func() {
	It("lists complete presets, referencing profiles, and storage sources", func() {
		catalog, source := runtimePresetTestCatalog()
		preset, err := catalog.CreatePreset(context.Background(), source.ID, runtimeprofiles.PresetInput{
			Name: "Review", Description: "Review changes", Scope: api.SpecLayerContext,
			Spec: api.RuntimePresetSpec(api.Spec{Model: api.Model{Name: "example-model", Mode: api.ModeAgent}}),
		})
		Expect(err).NotTo(HaveOccurred())
		_, err = catalog.CreateProfile(context.Background(), catalog.Sources()[1].ID, runtimeprofiles.ProfileInput{
			Name: "Plan and review", Presets: []string{preset.ID},
		})
		Expect(err).NotTo(HaveOccurred())

		rec := httptest.NewRecorder()
		serveRuntimePresetLibrary(rec, runtimePresetRequest(http.MethodGet, "", nil), catalog)

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		var response runtimePresetLibraryResponse
		Expect(json.Unmarshal(rec.Body.Bytes(), &response)).To(Succeed())
		Expect(response.Presets).To(HaveExactElements(preset))
		Expect(response.Profiles).To(HaveLen(1))
		Expect(response.Profiles[0].Presets).To(HaveExactElements(preset.ID))
		Expect(response.Sources).To(HaveLen(2))
		Expect(response.Sources[0]).To(Equal(source))
	})

	It("updates the selected preset without changing its source identity", func() {
		catalog, source := runtimePresetTestCatalog()
		preset, err := catalog.CreatePreset(context.Background(), source.ID, runtimeprofiles.PresetInput{
			Name: "Review", Scope: api.SpecLayerContext,
		})
		Expect(err).NotTo(HaveOccurred())
		input := runtimeprofiles.PresetInput{
			Name: "Review carefully", Description: "Inspect every changed path", Scope: api.SpecLayerSurface,
			Spec: api.RuntimePresetSpec(api.Spec{Budget: api.Budget{MaxTurns: 8}}),
		}

		rec := httptest.NewRecorder()
		serveRuntimePresetLibrary(rec, runtimePresetRequest(http.MethodPut, preset.ID, input), catalog)

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		var updated runtimeprofiles.Preset
		Expect(json.Unmarshal(rec.Body.Bytes(), &updated)).To(Succeed())
		Expect(updated.ID).To(Equal(preset.ID))
		Expect(updated.Source).To(Equal(source))
		Expect(updated.Name).To(Equal(input.Name))
		Expect(updated.Description).To(Equal(input.Description))
		Expect(updated.Spec.ToSpec().Budget.MaxTurns).To(Equal(8))
	})

	It("refuses referenced deletion, then deletes an unreferenced preset", func() {
		catalog, source := runtimePresetTestCatalog()
		preset, err := catalog.CreatePreset(context.Background(), source.ID, runtimeprofiles.PresetInput{
			Name: "Review", Scope: api.SpecLayerContext,
		})
		Expect(err).NotTo(HaveOccurred())
		profile, err := catalog.CreateProfile(context.Background(), catalog.Sources()[1].ID, runtimeprofiles.ProfileInput{
			Name: "Plan and review", Presets: []string{preset.ID},
		})
		Expect(err).NotTo(HaveOccurred())

		rec := httptest.NewRecorder()
		serveRuntimePresetLibrary(rec, runtimePresetRequest(http.MethodDelete, preset.ID, nil), catalog)
		Expect(rec.Code).To(Equal(http.StatusConflict), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring("Plan and review"))

		Expect(catalog.DeleteProfile(context.Background(), profile.ID)).To(Succeed())
		rec = httptest.NewRecorder()
		serveRuntimePresetLibrary(rec, runtimePresetRequest(http.MethodDelete, preset.ID, nil), catalog)
		Expect(rec.Code).To(Equal(http.StatusNoContent), rec.Body.String())
		_, err = catalog.GetPreset(context.Background(), preset.ID)
		Expect(errors.Is(err, runtimeprofiles.ErrNotFound)).To(BeTrue())
	})

	It("resolves the selected presets for the editor preview", func() {
		request := api.RuntimePresetResolveRequest{
			Selected: []string{"review"},
			Presets: []api.RuntimePreset{{
				ID: "review", Name: "Review", Scope: api.SpecLayerContext,
				Spec: api.RuntimePresetSpec(api.Spec{Model: api.Model{Name: "claude-example"}}),
			}},
		}

		rec := httptest.NewRecorder()
		(&Server{}).handleSettingsRuntimePresetResolve(rec, runtimePresetRequest(http.MethodPost, "", request))

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		var response runtimePresetResolveResponse
		Expect(json.Unmarshal(rec.Body.Bytes(), &response)).To(Succeed())
		Expect(response.Resolved.Spec.Model.Name).To(Equal("claude-example"))
		Expect(response.Tools).NotTo(BeEmpty())
	})
})
