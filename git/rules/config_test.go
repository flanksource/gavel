package rules

import (
	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Severity Config", func() {
	Describe("DefaultSeverityConfig", func() {
		It("should return non-nil config", func() {
			config := DefaultSeverityConfig()

			Expect(config).ToNot(BeNil())
			Expect(config.Default).To(Equal(models.Medium))
			Expect(config.Rules).ToNot(BeNil())
			Expect(len(config.Rules)).To(BeNumerically(">", 0))
		})

		It("should include critical rules for deletions", func() {
			config := DefaultSeverityConfig()

			severity, exists := config.Rules[`change.type == "deleted"`]
			Expect(exists).To(BeTrue())
			Expect(severity).To(Equal(models.Critical))
		})

		It("should include critical rules for Secrets", func() {
			config := DefaultSeverityConfig()

			severity, exists := config.Rules[`kubernetes.kind == "Secret"`]
			Expect(exists).To(BeTrue())
			Expect(severity).To(Equal(models.Critical))
		})

		It("should include high severity for RBAC resources", func() {
			config := DefaultSeverityConfig()

			severity, exists := config.Rules[`kubernetes.kind == "ServiceAccount"`]
			Expect(exists).To(BeTrue())
			Expect(severity).To(Equal(models.High))

			severity, exists = config.Rules[`kubernetes.kind == "Role"`]
			Expect(exists).To(BeTrue())
			Expect(severity).To(Equal(models.High))

			severity, exists = config.Rules[`kubernetes.kind == "ClusterRole"`]
			Expect(exists).To(BeTrue())
			Expect(severity).To(Equal(models.High))
		})

		It("should include rules for version changes", func() {
			config := DefaultSeverityConfig()

			severity, exists := config.Rules[`kubernetes.version_downgrade != ""`]
			Expect(exists).To(BeTrue())
			Expect(severity).To(Equal(models.High))

			severity, exists = config.Rules[`kubernetes.version_upgrade == "major"`]
			Expect(exists).To(BeTrue())
			Expect(severity).To(Equal(models.High))
		})

		It("should include rules for impact metrics", func() {
			config := DefaultSeverityConfig()

			severity, exists := config.Rules[`commit.line_changes > 500`]
			Expect(exists).To(BeTrue())
			Expect(severity).To(Equal(models.Critical))

			severity, exists = config.Rules[`commit.file_count > 10`]
			Expect(exists).To(BeTrue())
			Expect(severity).To(Equal(models.High))
		})
	})

	Describe("LoadSeverityConfig", func() {
		It("should load valid YAML config", func() {
			yamlData := []byte(`
default: high
rules:
  'change.type == "deleted"': critical
  'kubernetes.kind == "Pod"': medium
  'file.is_test': low
`)

			config, err := LoadSeverityConfig(yamlData)

			Expect(err).To(BeNil())
			Expect(config).ToNot(BeNil())
			Expect(config.Default).To(Equal(models.High))
			Expect(len(config.Rules)).To(Equal(3))
			Expect(config.Rules[`change.type == "deleted"`]).To(Equal(models.Critical))
			Expect(config.Rules[`kubernetes.kind == "Pod"`]).To(Equal(models.Medium))
			Expect(config.Rules[`file.is_test`]).To(Equal(models.Low))
		})

		It("should handle empty rules map", func() {
			yamlData := []byte(`
default: medium
rules: {}
`)

			config, err := LoadSeverityConfig(yamlData)

			Expect(err).To(BeNil())
			Expect(config.Default).To(Equal(models.Medium))
			Expect(config.Rules).ToNot(BeNil())
			Expect(len(config.Rules)).To(Equal(0))
		})

		It("should use medium default when not specified", func() {
			yamlData := []byte(`
rules:
  'change.type == "added"': low
`)

			config, err := LoadSeverityConfig(yamlData)

			Expect(err).To(BeNil())
			Expect(config.Default).To(Equal(models.Medium))
		})

		It("should fail with invalid YAML", func() {
			yamlData := []byte(`
this is not: valid: yaml: syntax
  - broken
`)

			config, err := LoadSeverityConfig(yamlData)

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to parse"))
			Expect(config).To(BeNil())
		})

		It("should fail with invalid severity value", func() {
			yamlData := []byte(`
default: medium
rules:
  'change.type == "deleted"': invalid_severity
`)

			config, err := LoadSeverityConfig(yamlData)

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid severity"))
			Expect(config).To(BeNil())
		})

		It("should accept all valid severity levels", func() {
			yamlData := []byte(`
default: info
rules:
  'rule1': critical
  'rule2': high
  'rule3': medium
  'rule4': low
  'rule5': info
`)

			config, err := LoadSeverityConfig(yamlData)

			Expect(err).To(BeNil())
			Expect(config.Default).To(Equal(models.Info))
			Expect(config.Rules["rule1"]).To(Equal(models.Critical))
			Expect(config.Rules["rule2"]).To(Equal(models.High))
			Expect(config.Rules["rule3"]).To(Equal(models.Medium))
			Expect(config.Rules["rule4"]).To(Equal(models.Low))
			Expect(config.Rules["rule5"]).To(Equal(models.Info))
		})

		It("should handle complex CEL expressions", func() {
			yamlData := []byte(`
default: medium
rules:
  'kubernetes.kind == "Deployment" && kubernetes.namespace == "production"': critical
  'commit.line_changes > 100 || commit.file_count > 10': high
  'file.is_config && change.type == "modified"': medium
`)

			config, err := LoadSeverityConfig(yamlData)

			Expect(err).To(BeNil())
			Expect(len(config.Rules)).To(Equal(3))
		})
	})

	Describe("Merge", func() {
		var defaultConfig *SeverityConfig

		BeforeEach(func() {
			defaultConfig = &SeverityConfig{
				Default: models.Medium,
				Rules: map[string]models.Severity{
					`default_rule1`: models.High,
					`default_rule2`: models.Medium,
					`default_rule3`: models.Low,
				},
			}
		})

		It("should return original config when override is nil", func() {
			merged := defaultConfig.Merge(nil)

			Expect(merged).ToNot(BeNil())
			Expect(merged.Default).To(Equal(models.Medium))
			Expect(len(merged.Rules)).To(Equal(3))
			Expect(merged.Rules["default_rule1"]).To(Equal(models.High))
		})

		It("should override default severity", func() {
			override := &SeverityConfig{
				Default: models.Critical,
				Rules:   map[string]models.Severity{},
			}

			merged := defaultConfig.Merge(override)

			Expect(merged.Default).To(Equal(models.Critical))
		})

		It("should combine rules from both configs", func() {
			override := &SeverityConfig{
				Default: models.Medium,
				Rules: map[string]models.Severity{
					`override_rule1`: models.Critical,
					`override_rule2`: models.High,
				},
			}

			merged := defaultConfig.Merge(override)

			Expect(len(merged.Rules)).To(Equal(5))
			Expect(merged.Rules["default_rule1"]).To(Equal(models.High))
			Expect(merged.Rules["default_rule2"]).To(Equal(models.Medium))
			Expect(merged.Rules["default_rule3"]).To(Equal(models.Low))
			Expect(merged.Rules["override_rule1"]).To(Equal(models.Critical))
			Expect(merged.Rules["override_rule2"]).To(Equal(models.High))
		})

		It("should override existing rules", func() {
			override := &SeverityConfig{
				Default: models.Medium,
				Rules: map[string]models.Severity{
					`default_rule1`: models.Critical, // Override from High to Critical
					`new_rule`:      models.Low,
				},
			}

			merged := defaultConfig.Merge(override)

			Expect(merged.Rules["default_rule1"]).To(Equal(models.Critical)) // Changed
			Expect(merged.Rules["default_rule2"]).To(Equal(models.Medium))   // Unchanged
			Expect(merged.Rules["new_rule"]).To(Equal(models.Low))           // Added
		})

		It("should not modify original configs", func() {
			override := &SeverityConfig{
				Default: models.Critical,
				Rules: map[string]models.Severity{
					`override_rule`: models.High,
				},
			}

			// Save original values
			originalDefaultSeverity := defaultConfig.Default
			originalRuleCount := len(defaultConfig.Rules)

			merged := defaultConfig.Merge(override)

			// Original config should be unchanged
			Expect(defaultConfig.Default).To(Equal(originalDefaultSeverity))
			Expect(len(defaultConfig.Rules)).To(Equal(originalRuleCount))

			// Merged should have new values
			Expect(merged.Default).To(Equal(models.Critical))
			Expect(len(merged.Rules)).To(Equal(4))
		})

		It("should preserve default severity when override has zero value", func() {
			override := &SeverityConfig{
				Default: "", // Zero value
				Rules: map[string]models.Severity{
					`override_rule`: models.High,
				},
			}

			merged := defaultConfig.Merge(override)

			Expect(merged.Default).To(Equal(models.Medium)) // Should keep default
		})

		It("should handle empty override rules", func() {
			override := &SeverityConfig{
				Default: models.High,
				Rules:   map[string]models.Severity{},
			}

			merged := defaultConfig.Merge(override)

			Expect(len(merged.Rules)).To(Equal(3)) // Should keep all default rules
			Expect(merged.Default).To(Equal(models.High))
		})

		It("should handle real-world merge scenario", func() {
			// Embedded defaults
			defaults := &SeverityConfig{
				Default: models.Medium,
				Rules: map[string]models.Severity{
					`change.type == "deleted"`:     models.Critical,
					`kubernetes.kind == "Secret"`:  models.Critical,
					`kubernetes.kind == "Service"`: models.High,
					`commit.file_count > 10`:       models.High,
				},
			}

			// User's arch.yaml custom rules
			user := &SeverityConfig{
				Default: models.Low, // More permissive default
				Rules: map[string]models.Severity{
					// Override Secret to be High instead of Critical
					`kubernetes.kind == "Secret"`: models.High,
					// Add custom rules
					`kubernetes.namespace == "prod"`:  models.Critical,
					`file.directory.contains("hack")`: models.Low,
				},
			}

			merged := defaults.Merge(user)

			// User's default wins
			Expect(merged.Default).To(Equal(models.Low))

			// User's override wins
			Expect(merged.Rules[`kubernetes.kind == "Secret"`]).To(Equal(models.High))

			// Default rules are preserved
			Expect(merged.Rules[`change.type == "deleted"`]).To(Equal(models.Critical))
			Expect(merged.Rules[`kubernetes.kind == "Service"`]).To(Equal(models.High))

			// User's new rules are added
			Expect(merged.Rules[`kubernetes.namespace == "prod"`]).To(Equal(models.Critical))
			Expect(merged.Rules[`file.directory.contains("hack")`]).To(Equal(models.Low))

			// Total rule count
			Expect(len(merged.Rules)).To(Equal(6))
		})
	})
})
