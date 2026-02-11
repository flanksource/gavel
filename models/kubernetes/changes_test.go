package kubernetes

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKubernetesChanges(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubernetes Changes Suite")
}

var _ = Describe("VersionChange Analysis", func() {
	Describe("AnalyzeVersionChange", func() {
		It("should detect major version change", func() {
			vc := AnalyzeVersionChange("v1.2.3", "v2.0.0")
			Expect(vc.OldVersion).To(Equal("v1.2.3"))
			Expect(vc.NewVersion).To(Equal("v2.0.0"))
			Expect(vc.ChangeType).To(Equal(VersionChangeMajor))
		})

		It("should detect minor version change", func() {
			vc := AnalyzeVersionChange("v1.2.3", "v1.3.0")
			Expect(vc.OldVersion).To(Equal("v1.2.3"))
			Expect(vc.NewVersion).To(Equal("v1.3.0"))
			Expect(vc.ChangeType).To(Equal(VersionChangeMinor))
		})

		It("should detect patch version change", func() {
			vc := AnalyzeVersionChange("v1.2.3", "v1.2.4")
			Expect(vc.OldVersion).To(Equal("v1.2.3"))
			Expect(vc.NewVersion).To(Equal("v1.2.4"))
			Expect(vc.ChangeType).To(Equal(VersionChangePatch))
		})

		It("should handle versions without v prefix", func() {
			vc := AnalyzeVersionChange("1.2.3", "1.3.0")
			Expect(vc.ChangeType).To(Equal(VersionChangeMinor))
		})

		It("should handle mixed v prefix", func() {
			vc := AnalyzeVersionChange("v1.2.3", "2.0.0")
			Expect(vc.ChangeType).To(Equal(VersionChangeMajor))
		})

		It("should return unknown for non-semver versions", func() {
			vc := AnalyzeVersionChange("latest", "stable")
			Expect(vc.OldVersion).To(Equal("latest"))
			Expect(vc.NewVersion).To(Equal("stable"))
			Expect(vc.ChangeType).To(Equal(VersionChangeUnknown))
		})

		It("should return unknown for git SHA versions", func() {
			vc := AnalyzeVersionChange("abc123", "def456")
			Expect(vc.ChangeType).To(Equal(VersionChangeUnknown))
		})

		It("should return unknown for custom tags", func() {
			vc := AnalyzeVersionChange("2024-01-15", "2024-01-16")
			Expect(vc.ChangeType).To(Equal(VersionChangeUnknown))
		})

		It("should handle downgrade scenarios", func() {
			// Downgrade shouldn't set a change type
			vc := AnalyzeVersionChange("v2.0.0", "v1.9.0")
			Expect(vc.ChangeType).To(Equal(VersionChangeUnknown))
		})

		It("should handle same version", func() {
			vc := AnalyzeVersionChange("v1.2.3", "v1.2.3")
			Expect(vc.ChangeType).To(Equal(VersionChangeUnknown))
		})

		It("should handle pre-release versions", func() {
			vc := AnalyzeVersionChange("v1.2.3", "v1.3.0-alpha.1")
			// Should still detect minor bump even with pre-release
			Expect(vc.ChangeType).To(Equal(VersionChangeMinor))
		})
	})

	Describe("VersionChange Pretty", func() {
		It("should include change type in output for major version", func() {
			vc := AnalyzeVersionChange("v1.2.3", "v2.0.0")
			text := vc.Pretty().String()
			Expect(text).To(ContainSubstring("major"))
		})

		It("should include change type in output for minor version", func() {
			vc := AnalyzeVersionChange("v1.2.3", "v1.3.0")
			text := vc.Pretty().String()
			Expect(text).To(ContainSubstring("minor"))
		})

		It("should include change type in output for patch version", func() {
			vc := AnalyzeVersionChange("v1.2.3", "v1.2.4")
			text := vc.Pretty().String()
			Expect(text).To(ContainSubstring("patch"))
		})

		It("should not include change type for unknown versions", func() {
			vc := AnalyzeVersionChange("latest", "stable")
			text := vc.Pretty().String()
			Expect(text).NotTo(ContainSubstring("major"))
			Expect(text).NotTo(ContainSubstring("minor"))
			Expect(text).NotTo(ContainSubstring("patch"))
		})
	})
})
