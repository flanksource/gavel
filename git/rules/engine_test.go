package rules

import (
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/models/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Severity Engine", func() {
	Describe("NewEngine", func() {
		It("should create engine with default config", func() {
			config := DefaultSeverityConfig()
			engine, err := NewEngine(config)

			Expect(err).To(BeNil())
			Expect(engine).ToNot(BeNil())
			Expect(engine.RuleCount()).To(BeNumerically(">", 0))
			Expect(engine.GetConfig()).To(Equal(config))
		})

		It("should create engine with nil config using defaults", func() {
			engine, err := NewEngine(nil)

			Expect(err).To(BeNil())
			Expect(engine).ToNot(BeNil())
			Expect(engine.RuleCount()).To(BeNumerically(">", 0))
		})

		It("should create engine with custom config", func() {
			config := &SeverityConfig{
				Default: models.Low,
				Rules: map[string]models.Severity{
					`change.type == "deleted"`: models.Critical,
					`kubernetes.kind == "Pod"`: models.High,
				},
			}
			engine, err := NewEngine(config)

			Expect(err).To(BeNil())
			Expect(engine.RuleCount()).To(Equal(2))
		})

		It("should fail with invalid CEL expression", func() {
			config := &SeverityConfig{
				Default: models.Medium,
				Rules: map[string]models.Severity{
					`this is not valid CEL syntax!!`: models.Critical,
				},
			}
			engine, err := NewEngine(config)

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to compile rule"))
			Expect(engine).To(BeNil())
		})

		It("should fail with undefined variable in CEL", func() {
			config := &SeverityConfig{
				Default: models.Medium,
				Rules: map[string]models.Severity{
					`undefined_variable == "test"`: models.Critical,
				},
			}
			engine, err := NewEngine(config)

			Expect(err).ToNot(BeNil())
			Expect(engine).To(BeNil())
		})
	})

	Describe("Evaluate", func() {
		var engine *Engine

		BeforeEach(func() {
			config := &SeverityConfig{
				Default: models.Medium,
				Rules: map[string]models.Severity{
					// Test rules in priority order
					`change.type == "deleted"`:                                    models.Critical,
					`kubernetes.kind == "Secret"`:                                 models.Critical,
					`kubernetes.kind == "Deployment"`:                             models.High,
					`commit.line_changes > 100`:                                   models.High,
					`kubernetes.has_env_change`:                                   models.Medium,
					`file.is_test`:                                                models.Low,
					`kubernetes.namespace == "dev"`:                               models.Low,
					`change.type == "added" && file.is_config`:                    models.Medium,
					`kubernetes.replica_delta > 5`:                                models.High,
					`kubernetes.version_upgrade == "major"`:                       models.High,
					`kubernetes.version_downgrade != ""`:                          models.Critical,
					`commit.file_count > 10`:                                      models.High,
					`file.extension == ".env"`:                                    models.High,
					`kubernetes.kind == "ConfigMap" && change.type == "modified"`: models.Low,
				},
			}
			var err error
			engine, err = NewEngine(config)
			Expect(err).To(BeNil())
		})

		It("should return critical for deleted resources", func() {
			ctx := map[string]any{
				"change": map[string]any{
					"type":           "deleted",
					"file":           "deployment.yaml",
					"adds":           0,
					"dels":           50,
					"fields_changed": []string{},
					"field_count":    0,
				},
				"commit":     buildEmptyCommitContext(),
				"kubernetes": buildEmptyKubernetesContext(),
				"file":       buildEmptyFileContext(),
			}

			severity := engine.Evaluate(ctx)
			Expect(severity).To(Equal(models.Critical))
		})

		It("should return critical for Secret resources", func() {
			ctx := map[string]any{
				"kubernetes": map[string]any{
					"is_kubernetes":       true,
					"kind":                "Secret",
					"api_version":         "v1",
					"namespace":           "default",
					"name":                "my-secret",
					"version_upgrade":     "",
					"version_downgrade":   "",
					"has_sha_change":      false,
					"replica_delta":       0,
					"has_env_change":      false,
					"has_resource_change": false,
				},
				"change": buildEmptyChangeContext(),
				"commit": buildEmptyCommitContext(),
				"file":   buildEmptyFileContext(),
			}

			severity := engine.Evaluate(ctx)
			Expect(severity).To(Equal(models.Critical))
		})

		It("should return high for Deployment resources", func() {
			ctx := map[string]any{
				"kubernetes": map[string]any{
					"is_kubernetes":       true,
					"kind":                "Deployment",
					"api_version":         "apps/v1",
					"namespace":           "production",
					"name":                "web-app",
					"version_upgrade":     "",
					"version_downgrade":   "",
					"has_sha_change":      false,
					"replica_delta":       0,
					"has_env_change":      false,
					"has_resource_change": false,
				},
				"change": buildEmptyChangeContext(),
				"commit": buildEmptyCommitContext(),
				"file":   buildEmptyFileContext(),
			}

			severity := engine.Evaluate(ctx)
			Expect(severity).To(Equal(models.High))
		})

		It("should return high for large commits", func() {
			ctx := map[string]any{
				"commit": map[string]any{
					"hash":           "abc123",
					"author":         "test-user",
					"author_email":   "test@example.com",
					"subject":        "Large refactor",
					"body":           "",
					"type":           "feat",
					"scope":          "",
					"file_count":     5,
					"line_changes":   250,
					"resource_count": 0,
				},
				"change":     buildEmptyChangeContext(),
				"kubernetes": buildEmptyKubernetesContext(),
				"file":       buildEmptyFileContext(),
			}

			severity := engine.Evaluate(ctx)
			Expect(severity).To(Equal(models.High))
		})

		It("should return low for test files", func() {
			ctx := map[string]any{
				"file": map[string]any{
					"extension": ".go",
					"directory": "pkg/handlers",
					"is_test":   true,
					"is_config": false,
					"tech":      "go",
				},
				"change":     buildEmptyChangeContext(),
				"commit":     buildEmptyCommitContext(),
				"kubernetes": buildEmptyKubernetesContext(),
			}

			severity := engine.Evaluate(ctx)
			Expect(severity).To(Equal(models.Low))
		})

		It("should return low for dev namespace", func() {
			ctx := map[string]any{
				"kubernetes": map[string]any{
					"is_kubernetes":       true,
					"kind":                "Pod",
					"api_version":         "v1",
					"namespace":           "dev",
					"name":                "test-pod",
					"version_upgrade":     "",
					"version_downgrade":   "",
					"has_sha_change":      false,
					"replica_delta":       0,
					"has_env_change":      false,
					"has_resource_change": false,
				},
				"change": buildEmptyChangeContext(),
				"commit": buildEmptyCommitContext(),
				"file":   buildEmptyFileContext(),
			}

			severity := engine.Evaluate(ctx)
			Expect(severity).To(Equal(models.Low))
		})

		It("should return default severity when no rules match", func() {
			ctx := map[string]any{
				"change": map[string]any{
					"type":           "modified",
					"file":           "README.md",
					"adds":           1,
					"dels":           1,
					"fields_changed": []string{},
					"field_count":    0,
				},
				"commit":     buildEmptyCommitContext(),
				"kubernetes": buildEmptyKubernetesContext(),
				"file": map[string]any{
					"extension": ".md",
					"directory": ".",
					"is_test":   false,
					"is_config": false,
					"tech":      "markdown",
				},
			}

			severity := engine.Evaluate(ctx)
			Expect(severity).To(Equal(models.Medium)) // Default is medium
		})

		It("should return high for significant scaling up", func() {
			ctx := map[string]any{
				"kubernetes": map[string]any{
					"is_kubernetes":       true,
					"kind":                "Deployment",
					"api_version":         "apps/v1",
					"namespace":           "production",
					"name":                "api-server",
					"version_upgrade":     "",
					"version_downgrade":   "",
					"has_sha_change":      false,
					"replica_delta":       10,
					"has_env_change":      false,
					"has_resource_change": false,
				},
				"change": buildEmptyChangeContext(),
				"commit": buildEmptyCommitContext(),
				"file":   buildEmptyFileContext(),
			}

			severity := engine.Evaluate(ctx)
			Expect(severity).To(Equal(models.High))
		})

		It("should return high for major version upgrade", func() {
			ctx := map[string]any{
				"kubernetes": map[string]any{
					"is_kubernetes":       true,
					"kind":                "Deployment",
					"api_version":         "apps/v1",
					"namespace":           "production",
					"name":                "api-server",
					"version_upgrade":     "major",
					"version_downgrade":   "",
					"has_sha_change":      false,
					"replica_delta":       0,
					"has_env_change":      false,
					"has_resource_change": false,
				},
				"change": buildEmptyChangeContext(),
				"commit": buildEmptyCommitContext(),
				"file":   buildEmptyFileContext(),
			}

			severity := engine.Evaluate(ctx)
			Expect(severity).To(Equal(models.High))
		})

		It("should return critical for version downgrade", func() {
			ctx := map[string]any{
				"kubernetes": map[string]any{
					"is_kubernetes":       true,
					"kind":                "Pod", // Use Pod which has no specific rule
					"api_version":         "v1",
					"namespace":           "production",
					"name":                "api-server",
					"version_upgrade":     "",
					"version_downgrade":   "minor",
					"has_sha_change":      false,
					"replica_delta":       0,
					"has_env_change":      false,
					"has_resource_change": false,
				},
				"change": buildEmptyChangeContext(),
				"commit": buildEmptyCommitContext(),
				"file":   buildEmptyFileContext(),
			}

			severity := engine.Evaluate(ctx)
			Expect(severity).To(Equal(models.Critical))
		})

		It("should use first matching rule (not highest severity)", func() {
			// ConfigMap + modified should match the specific low rule,
			// not the general Deployment high rule (even though kind doesn't match)
			ctx := map[string]any{
				"kubernetes": map[string]any{
					"is_kubernetes":       true,
					"kind":                "ConfigMap",
					"api_version":         "v1",
					"namespace":           "default",
					"name":                "app-config",
					"version_upgrade":     "",
					"version_downgrade":   "",
					"has_sha_change":      false,
					"replica_delta":       0,
					"has_env_change":      false,
					"has_resource_change": false,
				},
				"change": map[string]any{
					"type":           "modified",
					"file":           "config.yaml",
					"adds":           5,
					"dels":           3,
					"fields_changed": []string{"data.config"},
					"field_count":    1,
				},
				"commit": buildEmptyCommitContext(),
				"file":   buildEmptyFileContext(),
			}

			severity := engine.Evaluate(ctx)
			Expect(severity).To(Equal(models.Low))
		})
	})

	Describe("EvaluateWithDetails", func() {
		var engine *Engine

		BeforeEach(func() {
			config := &SeverityConfig{
				Default: models.Medium,
				Rules: map[string]models.Severity{
					`change.type == "deleted"`:    models.Critical,
					`kubernetes.kind == "Secret"`: models.Critical,
					`commit.line_changes > 100`:   models.High,
					`kubernetes.has_env_change`:   models.Medium,
					`file.is_test`:                models.Low,
				},
			}
			var err error
			engine, err = NewEngine(config)
			Expect(err).To(BeNil())
		})

		It("should return severity and matched rule expression", func() {
			ctx := map[string]any{
				"change": map[string]any{
					"type":           "deleted",
					"file":           "deployment.yaml",
					"adds":           0,
					"dels":           50,
					"fields_changed": []string{},
					"field_count":    0,
				},
				"commit":     buildEmptyCommitContext(),
				"kubernetes": buildEmptyKubernetesContext(),
				"file":       buildEmptyFileContext(),
			}

			severity, rule, err := engine.EvaluateWithDetails(ctx)

			Expect(err).To(BeNil())
			Expect(severity).To(Equal(models.Critical))
			Expect(rule).To(Equal(`change.type == "deleted"`))
		})

		It("should return empty rule when using default severity", func() {
			ctx := map[string]any{
				"change": map[string]any{
					"type":           "modified",
					"file":           "something.txt",
					"adds":           1,
					"dels":           1,
					"fields_changed": []string{},
					"field_count":    0,
				},
				"commit":     buildEmptyCommitContext(),
				"kubernetes": buildEmptyKubernetesContext(),
				"file":       buildEmptyFileContext(),
			}

			severity, rule, err := engine.EvaluateWithDetails(ctx)

			Expect(err).To(BeNil())
			Expect(severity).To(Equal(models.Medium)) // Default
			Expect(rule).To(BeEmpty())
		})
	})

	Describe("TestExpression", func() {
		var engine *Engine

		BeforeEach(func() {
			config := DefaultSeverityConfig()
			var err error
			engine, err = NewEngine(config)
			Expect(err).To(BeNil())
		})

		It("should evaluate simple true expression", func() {
			ctx := map[string]any{
				"change":     map[string]any{"type": "deleted"},
				"commit":     buildEmptyCommitContext(),
				"kubernetes": buildEmptyKubernetesContext(),
				"file":       buildEmptyFileContext(),
			}

			result, err := engine.TestExpression(`change.type == "deleted"`, ctx)

			Expect(err).To(BeNil())
			Expect(result).To(BeTrue())
		})

		It("should evaluate simple false expression", func() {
			ctx := map[string]any{
				"change":     map[string]any{"type": "added"},
				"commit":     buildEmptyCommitContext(),
				"kubernetes": buildEmptyKubernetesContext(),
				"file":       buildEmptyFileContext(),
			}

			result, err := engine.TestExpression(`change.type == "deleted"`, ctx)

			Expect(err).To(BeNil())
			Expect(result).To(BeFalse())
		})

		It("should handle complex expressions with multiple conditions", func() {
			ctx := map[string]any{
				"kubernetes": map[string]any{
					"kind":      "Deployment",
					"namespace": "production",
				},
				"change": map[string]any{"type": "modified"},
				"commit": buildEmptyCommitContext(),
				"file":   buildEmptyFileContext(),
			}

			result, err := engine.TestExpression(
				`kubernetes.kind == "Deployment" && kubernetes.namespace == "production"`,
				ctx,
			)

			Expect(err).To(BeNil())
			Expect(result).To(BeTrue())
		})

		It("should fail with invalid CEL syntax", func() {
			ctx := buildFullContext(nil, nil, nil)

			result, err := engine.TestExpression(`this is not valid!!!`, ctx)

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("CEL"))
			Expect(result).To(BeFalse())
		})

		It("should fail with non-boolean expression", func() {
			ctx := map[string]any{
				"commit":     map[string]any{"line_changes": 100},
				"change":     buildEmptyChangeContext(),
				"kubernetes": buildEmptyKubernetesContext(),
				"file":       buildEmptyFileContext(),
			}

			result, err := engine.TestExpression(`commit.line_changes`, ctx)

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("bool"))
			Expect(result).To(BeFalse())
		})
	})
})

// Helper functions to build empty context maps

func buildEmptyCommitContext() map[string]any {
	return map[string]any{
		"hash":           "",
		"author":         "",
		"author_email":   "",
		"subject":        "",
		"body":           "",
		"type":           "",
		"scope":          "",
		"file_count":     0,
		"line_changes":   0,
		"resource_count": 0,
	}
}

func buildEmptyChangeContext() map[string]any {
	return map[string]any{
		"type":           "",
		"file":           "",
		"adds":           0,
		"dels":           0,
		"fields_changed": []string{},
		"field_count":    0,
	}
}

func buildEmptyKubernetesContext() map[string]any {
	return map[string]any{
		"is_kubernetes":       false,
		"kind":                "",
		"api_version":         "",
		"namespace":           "",
		"name":                "",
		"version_upgrade":     "",
		"version_downgrade":   "",
		"has_sha_change":      false,
		"replica_delta":       0,
		"has_env_change":      false,
		"has_resource_change": false,
	}
}

func buildEmptyFileContext() map[string]any {
	return map[string]any{
		"extension": "",
		"directory": "",
		"is_test":   false,
		"is_config": false,
		"tech":      "",
	}
}

func buildFullContext(commit *models.CommitAnalysis, change *models.CommitChange, k8sChange *kubernetes.KubernetesChange) map[string]any {
	return BuildContext(commit, change, k8sChange)
}
