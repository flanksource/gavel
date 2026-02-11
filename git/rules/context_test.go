package rules

import (
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/models/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Context Builder", func() {
	Describe("BuildContext", func() {
		It("should build context with all nil inputs", func() {
			ctx := BuildContext(nil, nil, nil)

			Expect(ctx).ToNot(BeNil())
			Expect(ctx).To(HaveKey("commit"))
			Expect(ctx).To(HaveKey("change"))
			Expect(ctx).To(HaveKey("kubernetes"))
			Expect(ctx).To(HaveKey("file"))

			// Verify no nil values
			commitCtx := ctx["commit"].(map[string]any)
			Expect(commitCtx["hash"]).To(Equal(""))
			Expect(commitCtx["author"]).To(Equal(""))
			Expect(commitCtx["file_count"]).To(Equal(0))
		})

		It("should build full context with all data populated", func() {
			commit := &models.CommitAnalysis{
				Tech: []models.ScopeTechnology{models.Kubernetes},
			}
			commit.Hash = "abc123"
			commit.Subject = "feat: add feature"
			commit.Author.Name = "Test User"
			commit.Author.Email = "test@example.com"
			commit.CommitType = "feat"
			commit.Scope = models.ScopeTypeApp

			change := &models.CommitChange{
				File: "deployment.yaml",
				Type: models.SourceChangeTypeModified,
				Adds: 10,
				Dels: 5,
			}

			k8sChange := &kubernetes.KubernetesChange{
				KubernetesRef: kubernetes.KubernetesRef{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
					Namespace:  "production",
					Name:       "web-app",
				},
			}

			ctx := BuildContext(commit, change, k8sChange)

			// Verify commit context
			commitCtx := ctx["commit"].(map[string]any)
			Expect(commitCtx["hash"]).To(Equal("abc123"))
			Expect(commitCtx["author"]).To(Equal("Test User"))
			Expect(commitCtx["author_email"]).To(Equal("test@example.com"))
			Expect(commitCtx["subject"]).To(Equal("feat: add feature"))

			// Verify change context
			changeCtx := ctx["change"].(map[string]any)
			Expect(changeCtx["file"]).To(Equal("deployment.yaml"))
			Expect(changeCtx["type"]).To(Equal(string(models.SourceChangeTypeModified)))
			Expect(changeCtx["adds"]).To(Equal(10))
			Expect(changeCtx["dels"]).To(Equal(5))

			// Verify kubernetes context
			k8sCtx := ctx["kubernetes"].(map[string]any)
			Expect(k8sCtx["is_kubernetes"]).To(BeTrue())
			Expect(k8sCtx["kind"]).To(Equal("Deployment"))
			Expect(k8sCtx["api_version"]).To(Equal("apps/v1"))
			Expect(k8sCtx["namespace"]).To(Equal("production"))
			Expect(k8sCtx["name"]).To(Equal("web-app"))
		})
	})

	Describe("buildCommitContext", func() {
		It("should return empty context for nil commit", func() {
			ctx := buildCommitContext(nil)

			Expect(ctx).ToNot(BeNil())
			Expect(ctx["hash"]).To(Equal(""))
			Expect(ctx["author"]).To(Equal(""))
			Expect(ctx["author_email"]).To(Equal(""))
			Expect(ctx["file_count"]).To(Equal(0))
			Expect(ctx["line_changes"]).To(Equal(0))
			Expect(ctx["resource_count"]).To(Equal(0))
		})

		It("should calculate line changes correctly", func() {
			commit := &models.CommitAnalysis{}
			commit.Changes = []models.CommitChange{
				{Adds: 100, Dels: 50},
				{Adds: 25, Dels: 10},
			}
			commit.TotalLineChanges = 185 // Pre-computed

			ctx := buildCommitContext(commit)

			Expect(ctx["line_changes"]).To(Equal(185)) // 100+50+25+10
		})

		It("should count kubernetes resources correctly", func() {
			commit := &models.CommitAnalysis{}
			commit.Changes = []models.CommitChange{
				{
					KubernetesChanges: []kubernetes.KubernetesChange{
						{},
						{},
					},
				},
				{
					KubernetesChanges: []kubernetes.KubernetesChange{
						{},
					},
				},
			}
			commit.TotalResourceCount = 3 // Pre-computed

			ctx := buildCommitContext(commit)

			Expect(ctx["resource_count"]).To(Equal(3))
		})

		It("should extract commit metadata", func() {
			commit := &models.CommitAnalysis{}
			commit.Hash = "abc123"
			commit.Subject = "fix(api): resolve bug"
			commit.Body = "Detailed description"
			commit.Author.Name = "Jane Smith"
			commit.Author.Email = "jane@example.com"
			commit.CommitType = "fix"
			commit.Scope = models.ScopeTypeApp

			ctx := buildCommitContext(commit)

			Expect(ctx["hash"]).To(Equal("abc123"))
			Expect(ctx["subject"]).To(Equal("fix(api): resolve bug"))
			Expect(ctx["body"]).To(Equal("Detailed description"))
			Expect(ctx["author"]).To(Equal("Jane Smith"))
			Expect(ctx["author_email"]).To(Equal("jane@example.com"))
			Expect(ctx["type"]).To(Equal("fix"))
			Expect(ctx["scope"]).To(Equal(string(models.ScopeTypeApp)))
		})
	})

	Describe("buildChangeContext", func() {
		It("should return empty context for nil change", func() {
			ctx := buildChangeContext(nil)

			Expect(ctx).ToNot(BeNil())
			Expect(ctx["type"]).To(Equal(""))
			Expect(ctx["file"]).To(Equal(""))
			Expect(ctx["adds"]).To(Equal(0))
			Expect(ctx["dels"]).To(Equal(0))
			Expect(ctx["fields_changed"]).To(Equal([]string{}))
			Expect(ctx["field_count"]).To(Equal(0))
		})

		It("should extract basic change info", func() {
			change := &models.CommitChange{
				File: "api/handler.go",
				Type: models.SourceChangeTypeModified,
				Adds: 42,
				Dels: 18,
			}

			ctx := buildChangeContext(change)

			Expect(ctx["type"]).To(Equal(string(models.SourceChangeTypeModified)))
			Expect(ctx["file"]).To(Equal("api/handler.go"))
			Expect(ctx["adds"]).To(Equal(42))
			Expect(ctx["dels"]).To(Equal(18))
		})

		It("should aggregate kubernetes fields changed", func() {
			change := &models.CommitChange{
				File: "deployment.yaml",
				Type: models.SourceChangeTypeModified,
				KubernetesChanges: []kubernetes.KubernetesChange{
					{
						FieldsChanged:    []string{"spec.replicas", "spec.template.spec.containers[0].image"},
						FieldChangeCount: 2,
					},
					{
						FieldsChanged:    []string{"metadata.labels", "spec.replicas"},
						FieldChangeCount: 2,
					},
				},
			}

			ctx := buildChangeContext(change)

			fieldsChanged := ctx["fields_changed"].([]string)
			Expect(fieldsChanged).To(HaveLen(3))    // Unique: replicas, image, labels
			Expect(ctx["field_count"]).To(Equal(4)) // Total: 2 + 2
		})
	})

	Describe("buildKubernetesContext", func() {
		It("should return empty context for nil k8s change", func() {
			ctx := buildKubernetesContext(nil)

			Expect(ctx).ToNot(BeNil())
			Expect(ctx["is_kubernetes"]).To(BeFalse())
			Expect(ctx["kind"]).To(Equal(""))
			Expect(ctx["namespace"]).To(Equal(""))
			Expect(ctx["version_upgrade"]).To(Equal(""))
			Expect(ctx["version_downgrade"]).To(Equal(""))
			Expect(ctx["has_sha_change"]).To(BeFalse())
			Expect(ctx["replica_delta"]).To(Equal(0))
		})

		It("should extract basic kubernetes info", func() {
			k8sChange := &kubernetes.KubernetesChange{
				KubernetesRef: kubernetes.KubernetesRef{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
					Namespace:  "production",
					Name:       "web-app",
				},
			}

			ctx := buildKubernetesContext(k8sChange)

			Expect(ctx["is_kubernetes"]).To(BeTrue())
			Expect(ctx["kind"]).To(Equal("Deployment"))
			Expect(ctx["api_version"]).To(Equal("apps/v1"))
			Expect(ctx["namespace"]).To(Equal("production"))
			Expect(ctx["name"]).To(Equal("web-app"))
		})

		It("should detect major version upgrade", func() {
			k8sChange := &kubernetes.KubernetesChange{
				VersionChanges: []kubernetes.VersionChange{
					{
						ChangeType: kubernetes.VersionChangeMajor,
						OldVersion: "1.0.0",
						NewVersion: "2.0.0",
					},
				},
			}

			ctx := buildKubernetesContext(k8sChange)

			Expect(ctx["version_upgrade"]).To(Equal("major"))
		})

		It("should detect minor version upgrade", func() {
			k8sChange := &kubernetes.KubernetesChange{
				VersionChanges: []kubernetes.VersionChange{
					{ChangeType: kubernetes.VersionChangeMinor},
				},
			}

			ctx := buildKubernetesContext(k8sChange)

			Expect(ctx["version_upgrade"]).To(Equal("minor"))
		})

		It("should detect patch version upgrade", func() {
			k8sChange := &kubernetes.KubernetesChange{
				VersionChanges: []kubernetes.VersionChange{
					{ChangeType: kubernetes.VersionChangePatch},
				},
			}

			ctx := buildKubernetesContext(k8sChange)

			Expect(ctx["version_upgrade"]).To(Equal("patch"))
		})

		It("should calculate replica delta", func() {
			oldReplicas := 3
			newReplicas := 10
			k8sChange := &kubernetes.KubernetesChange{
				Scaling: &kubernetes.Scaling{
					Replicas:    &oldReplicas,
					NewReplicas: &newReplicas,
				},
			}

			ctx := buildKubernetesContext(k8sChange)

			Expect(ctx["replica_delta"]).To(Equal(7))
		})

		It("should detect environment changes", func() {
			k8sChange := &kubernetes.KubernetesChange{
				EnvironmentChange: &kubernetes.EnvironmentChange{
					Old: map[string]string{"VAR1": "old"},
					New: map[string]string{"VAR1": "new", "VAR2": "value"},
				},
			}

			ctx := buildKubernetesContext(k8sChange)

			Expect(ctx["has_env_change"]).To(BeTrue())
		})

		It("should detect resource changes", func() {
			k8sChange := &kubernetes.KubernetesChange{
				Scaling: &kubernetes.Scaling{
					OldCPU:    "100m",
					NewCPU:    "500m",
					OldMemory: "256Mi",
					NewMemory: "512Mi",
				},
			}

			ctx := buildKubernetesContext(k8sChange)

			Expect(ctx["has_resource_change"]).To(BeTrue())
		})
	})

	Describe("buildFileContext", func() {
		It("should return empty context for nil change", func() {
			ctx := buildFileContext(nil)

			Expect(ctx).ToNot(BeNil())
			Expect(ctx["extension"]).To(Equal(""))
			Expect(ctx["directory"]).To(Equal(""))
			Expect(ctx["is_test"]).To(BeFalse())
			Expect(ctx["is_config"]).To(BeFalse())
		})

		It("should detect file extension and directory", func() {
			change := &models.CommitChange{
				File: "api/handler.go",
			}

			ctx := buildFileContext(change)

			Expect(ctx["extension"]).To(Equal(".go"))
			Expect(ctx["directory"]).To(Equal("api"))
		})

		It("should detect test files", func() {
			testCases := []string{
				"handler_test.go",
				"Button.spec.tsx",
				"utils.test.js",
			}

			for _, file := range testCases {
				change := &models.CommitChange{File: file}
				ctx := buildFileContext(change)
				Expect(ctx["is_test"]).To(BeTrue(), "File %s should be detected as test", file)
			}
		})

		It("should detect config files", func() {
			testCases := []string{
				"config.yaml",
				"settings.json",
				".env",
				".env.production",
			}

			for _, file := range testCases {
				change := &models.CommitChange{File: file}
				ctx := buildFileContext(change)
				Expect(ctx["is_config"]).To(BeTrue(), "File %s should be detected as config", file)
			}
		})

		It("should detect tech from change metadata", func() {
			change := &models.CommitChange{
				File: "deployment.yaml",
				Tech: []models.ScopeTechnology{models.Kubernetes},
			}

			ctx := buildFileContext(change)

			Expect(ctx["tech"]).To(Equal(string(models.Kubernetes)))
		})

		It("should detect tech from file extension", func() {
			testCases := map[string]string{
				"main.go":       "go",
				"script.py":     "python",
				"component.tsx": "typescript",
				"utils.js":      "javascript",
				"config.yaml":   "yaml",
				"data.json":     "json",
				"README.md":     "markdown",
				"deploy.sh":     "bash",
				"query.sql":     "sql",
			}

			for file, expectedTech := range testCases {
				change := &models.CommitChange{File: file}
				ctx := buildFileContext(change)
				Expect(ctx["tech"]).To(Equal(expectedTech), "File %s should detect tech as %s", file, expectedTech)
			}
		})
	})
})
