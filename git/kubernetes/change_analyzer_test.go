package kubernetes_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gkubernetes "github.com/flanksource/gavel/git/kubernetes"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/models/kubernetes"
	"github.com/flanksource/gavel/repomap"
)

type ExpectedResult struct {
	ExpectedCount int              `json:"expectedCount"`
	Changes       []ExpectedChange `json:"changes"`
}

type ExpectedChange struct {
	Kind             string                          `json:"kind"`
	Name             string                          `json:"name"`
	Namespace        string                          `json:"namespace"`
	ChangeType       kubernetes.SourceChangeType     `json:"changeType"`
	Severity         kubernetes.ChangeSeverity       `json:"severity"`
	SourceType       kubernetes.KubernetesSourceType `json:"sourceType"`
	Summary          string                          `json:"summary,omitempty"`
	FieldsChanged    []string                        `json:"fieldsChanged,omitempty"`
	HasScaling       bool                            `json:"hasScaling,omitempty"`
	Scaling          *ExpectedScaling                `json:"scaling,omitempty"`
	HasVersionChange bool                            `json:"hasVersionChange,omitempty"`
	VersionChange    *ExpectedVersionChange          `json:"versionChange,omitempty"`
}

type ExpectedScaling struct {
	OldReplicas int `json:"oldReplicas,omitempty"`
	NewReplicas int `json:"newReplicas,omitempty"`
}

type ExpectedVersionChange struct {
	OldVersion string `json:"oldVersion,omitempty"`
	NewVersion string `json:"newVersion,omitempty"`
}

var _ = Describe("AnalyzeKubernetesChanges", func() {
	fixturesDir := "fixtures"

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		panic("Failed to read fixtures directory: " + err.Error())
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		fixtureName := entry.Name()
		fixturePath := filepath.Join(fixturesDir, fixtureName)

		It("should correctly analyze "+fixtureName, func() {
			beforeYAML := loadFile(filepath.Join(fixturePath, "before.yaml"))
			afterYAML := loadFile(filepath.Join(fixturePath, "after.yaml"))
			patchDiff := loadFile(filepath.Join(fixturePath, "patch.diff"))
			expected := loadExpectedResult(filepath.Join(fixturePath, "expected.json"))

			var beforeDocs []kubernetes.YAMLDocument
			var afterDocs []kubernetes.YAMLDocument
			var err error

			if beforeYAML != "" && beforeYAML != "# Empty file - service doesn't exist yet\n" && beforeYAML != "# File deleted - deployment removed\n" {
				beforeDocs, err = repomap.ParseYAMLDocuments(beforeYAML)
				Expect(err).ToNot(HaveOccurred())
			}

			if afterYAML != "" && afterYAML != "# Empty file - service doesn't exist yet\n" && afterYAML != "# File deleted - deployment removed\n" {
				afterDocs, err = repomap.ParseYAMLDocuments(afterYAML)
				Expect(err).ToNot(HaveOccurred())
			}

			changedLineInfo := gkubernetes.ExtractChangedLines(patchDiff)

			isFileDeletion := strings.Contains(patchDiff, "deleted file mode")
			isPartialDeletion := len(beforeDocs) > len(afterDocs) && len(afterDocs) > 0
			if !isFileDeletion && !isPartialDeletion {
				Expect(len(changedLineInfo.AddedLines)+len(changedLineInfo.DeletedLines)).To(BeNumerically(">", 0), "Should extract changed lines from patch")
			}

			var affectedIndices []int
			if len(afterDocs) > 0 {
				affectedIndices = gkubernetes.FindAffectedDocuments(afterDocs, changedLineInfo.AddedLines)
			}

			var actualChanges []kubernetes.KubernetesChange

			for _, idx := range affectedIndices {
				if idx >= len(afterDocs) {
					continue
				}

				afterDoc := afterDocs[idx]
				if !repomap.IsKubernetesResource(afterDoc.Content) {
					continue
				}

				afterRef := repomap.ExtractKubernetesRef(afterDoc)
				var beforeDoc *kubernetes.YAMLDocument
				for i := range beforeDocs {
					if !repomap.IsKubernetesResource(beforeDocs[i].Content) {
						continue
					}
					beforeRef := repomap.ExtractKubernetesRef(beforeDocs[i])

					if beforeRef.Kind == afterRef.Kind &&
						beforeRef.Name == afterRef.Name &&
						beforeRef.Namespace == afterRef.Namespace {
						beforeDoc = &beforeDocs[i]
						break
					}
				}

				k8sChange := createChange(beforeDoc, afterDoc)
				actualChanges = append(actualChanges, k8sChange)
			}

			if len(beforeDocs) > 0 && len(afterDocs) == 0 {
				for _, beforeDoc := range beforeDocs {
					if !repomap.IsKubernetesResource(beforeDoc.Content) {
						continue
					}

					k8sChange := createChange(&beforeDoc, kubernetes.YAMLDocument{})
					actualChanges = append(actualChanges, k8sChange)
				}
			} else if len(beforeDocs) > len(afterDocs) {
				for _, beforeDoc := range beforeDocs {
					if !repomap.IsKubernetesResource(beforeDoc.Content) {
						continue
					}

					beforeRef := repomap.ExtractKubernetesRef(beforeDoc)
					found := false

					for _, afterDoc := range afterDocs {
						if !repomap.IsKubernetesResource(afterDoc.Content) {
							continue
						}
						afterRef := repomap.ExtractKubernetesRef(afterDoc)

						if beforeRef.Kind == afterRef.Kind &&
							beforeRef.Name == afterRef.Name &&
							beforeRef.Namespace == afterRef.Namespace {
							found = true
							break
						}
					}

					if !found {
						k8sChange := createChange(&beforeDoc, kubernetes.YAMLDocument{})
						actualChanges = append(actualChanges, k8sChange)
					}
				}
			}

			Expect(actualChanges).To(HaveLen(expected.ExpectedCount),
				"Expected %d changes but got %d", expected.ExpectedCount, len(actualChanges))

			for i, expectedChange := range expected.Changes {
				Expect(i).To(BeNumerically("<", len(actualChanges)),
					"Not enough actual changes to compare")

				actual := actualChanges[i]

				Expect(actual.Kind).To(Equal(expectedChange.Kind),
					"Kind mismatch for change %d", i)
				Expect(actual.Name).To(Equal(expectedChange.Name),
					"Name mismatch for change %d", i)
				Expect(actual.Namespace).To(Equal(expectedChange.Namespace),
					"Namespace mismatch for change %d", i)
				Expect(actual.ChangeType).To(Equal(expectedChange.ChangeType),
					"ChangeType mismatch for change %d", i)
				Expect(actual.Severity).To(Equal(expectedChange.Severity),
					"Severity mismatch for change %d", i)
				Expect(actual.SourceType).To(Equal(expectedChange.SourceType),
					"SourceType mismatch for change %d", i)

				if len(expectedChange.FieldsChanged) > 0 {
					Expect(actual.FieldsChanged).To(ConsistOf(expectedChange.FieldsChanged),
						"FieldsChanged mismatch for change %d", i)
				}

				if expectedChange.HasScaling {
					Expect(actual.Scaling).ToNot(BeNil(), "Expected scaling change for %d", i)
					if expectedChange.Scaling != nil {
						if expectedChange.Scaling.OldReplicas > 0 {
							Expect(actual.Scaling.Replicas).ToNot(BeNil())
							Expect(*actual.Scaling.Replicas).To(Equal(expectedChange.Scaling.OldReplicas))
						}
						if expectedChange.Scaling.NewReplicas > 0 {
							Expect(actual.Scaling.NewReplicas).ToNot(BeNil())
							Expect(*actual.Scaling.NewReplicas).To(Equal(expectedChange.Scaling.NewReplicas))
						}
					}
				}

				if expectedChange.HasVersionChange {
					Expect(len(actual.VersionChanges)).To(BeNumerically(">", 0), "Expected version changes for %d", i)
					if expectedChange.VersionChange != nil && len(actual.VersionChanges) > 0 {
						Expect(actual.VersionChanges[0].OldVersion).To(Equal(expectedChange.VersionChange.OldVersion))
						Expect(actual.VersionChanges[0].NewVersion).To(Equal(expectedChange.VersionChange.NewVersion))
					}
				}

				Expect(actual.StartLine).To(BeNumerically(">", 0), "StartLine should be set for change %d", i)
				Expect(actual.EndLine).To(BeNumerically(">=", actual.StartLine), "EndLine should be >= StartLine for change %d", i)
			}
		})
	}
})

var _ = Describe("Line Range Formatting", func() {
	It("should format single line", func() {
		result := models.NewLineRanges([]int{5})
		Expect(result.String()).To(Equal("5"))
	})

	It("should format contiguous range", func() {
		result := models.NewLineRanges([]int{1, 2, 3, 4, 5})
		Expect(result.String()).To(Equal("1-5"))
	})

	It("should format multiple ranges", func() {
		result := models.NewLineRanges([]int{1, 2, 3, 7, 8, 9, 12})
		Expect(result.String()).To(Equal("1-3,7-9,12"))
	})

	It("should sort unsorted input", func() {
		result := models.NewLineRanges([]int{9, 1, 3, 2, 8, 7})
		Expect(result.String()).To(Equal("1-3,7-9"))
	})

	It("should handle empty input", func() {
		result := models.NewLineRanges([]int{})
		Expect(result.String()).To(Equal(""))
	})

	It("should handle single element range", func() {
		result := models.NewLineRanges([]int{5, 6})
		Expect(result.String()).To(Equal("5-6"))
	})

	It("should handle mixed single lines and ranges", func() {
		result := models.NewLineRanges([]int{1, 3, 4, 5, 7, 10, 11, 15})
		Expect(result.String()).To(Equal("1,3-5,7,10-11,15"))
	})
})

var _ = Describe("Line Tracking", func() {
	It("should track line ranges for multi-document YAML", func() {
		yaml := `apiVersion: v1
kind: Namespace
metadata:
  name: production
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: production
data:
  env: "production"
---
apiVersion: v1
kind: Service
metadata:
  name: app-service
  namespace: production
spec:
  selector:
    app: myapp
`

		docs, err := repomap.ParseYAMLDocuments(yaml)
		Expect(err).ToNot(HaveOccurred())
		Expect(docs).To(HaveLen(3))

		Expect(docs[0].StartLine).To(Equal(1))
		Expect(docs[0].EndLine).To(BeNumerically(">", 1))

		Expect(docs[1].StartLine).To(BeNumerically(">", docs[0].EndLine))
		Expect(docs[1].EndLine).To(BeNumerically(">", docs[1].StartLine))

		Expect(docs[2].StartLine).To(BeNumerically(">", docs[1].EndLine))
		Expect(docs[2].EndLine).To(BeNumerically(">", docs[2].StartLine))

		for i, doc := range docs {
			ref := repomap.ExtractKubernetesRef(doc)
			GinkgoWriter.Printf("Doc %d: %s/%s (lines %d-%d)\n", i, ref.Kind, ref.Name, doc.StartLine, doc.EndLine)
		}
	})

	It("should include line ranges in KubernetesChange", func() {
		beforeYAML := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  key: "value"
`

		afterYAML := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  key: "newvalue"
  newkey: "test"
`

		beforeDocs, err := repomap.ParseYAMLDocuments(beforeYAML)
		Expect(err).ToNot(HaveOccurred())
		Expect(beforeDocs).To(HaveLen(1))

		afterDocs, err := repomap.ParseYAMLDocuments(afterYAML)
		Expect(err).ToNot(HaveOccurred())
		Expect(afterDocs).To(HaveLen(1))

		beforeDoc := beforeDocs[0]
		afterDoc := afterDocs[0]

		change := createChange(&beforeDoc, afterDoc)

		Expect(change.StartLine).To(Equal(afterDoc.StartLine))
		Expect(change.EndLine).To(Equal(afterDoc.EndLine))
		Expect(change.StartLine).To(BeNumerically(">", 0))
		Expect(change.EndLine).To(BeNumerically(">=", change.StartLine))

		GinkgoWriter.Printf("Change for %s/%s covers lines %d-%d\n",
			change.Kind, change.Name, change.StartLine, change.EndLine)
	})
})

func loadFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func loadExpectedResult(path string) ExpectedResult {
	content, err := os.ReadFile(path)
	Expect(err).ToNot(HaveOccurred(), "Failed to read expected.json from "+path)

	var result ExpectedResult
	err = json.Unmarshal(content, &result)
	Expect(err).ToNot(HaveOccurred(), "Failed to parse expected.json from "+path)

	return result
}

func createChange(beforeDoc *kubernetes.YAMLDocument, afterDoc kubernetes.YAMLDocument) kubernetes.KubernetesChange {
	refDoc := afterDoc
	if afterDoc.Content == nil && beforeDoc != nil {
		refDoc = *beforeDoc
	}

	ref := repomap.ExtractKubernetesRef(refDoc)

	var changeType kubernetes.SourceChangeType
	var beforeContent, afterContent map[string]interface{}

	if beforeDoc == nil || beforeDoc.Content == nil {
		changeType = kubernetes.SourceChangeTypeAdded
		afterContent = afterDoc.Content
	} else if afterDoc.Content == nil {
		changeType = kubernetes.SourceChangeTypeDeleted
		beforeContent = beforeDoc.Content
	} else {
		changeType = kubernetes.SourceChangeTypeModified
		beforeContent = beforeDoc.Content
		afterContent = afterDoc.Content
	}

	var patches []kubernetes.ExtendedPatch
	if changeType == kubernetes.SourceChangeTypeModified {
		patches, _ = gkubernetes.GenerateJSONPatches(beforeContent, afterContent)
	}

	fieldPaths := gkubernetes.ExtractFieldPaths(patches)

	sourceType := gkubernetes.DetermineSourceType("test.yaml", afterContent)

	var scaling *kubernetes.Scaling
	var versionChanges []kubernetes.VersionChange
	var envChange *kubernetes.EnvironmentChange

	if changeType == kubernetes.SourceChangeTypeModified {
		scaling = gkubernetes.ExtractScalingChanges(patches, beforeContent, afterContent)
		versionChanges = gkubernetes.ExtractAllVersionChanges(patches, beforeContent, afterContent, nil)
		envChange = gkubernetes.ExtractEnvironmentChanges(patches, beforeContent, afterContent)
	}

	severity := gkubernetes.DetermineChangeSeverity(changeType, patches, versionChanges)

	return kubernetes.KubernetesChange{
		KubernetesRef:     ref,
		ChangeType:        changeType,
		SourceType:        sourceType,
		Severity:          severity,
		Patches:           patches,
		Scaling:           scaling,
		VersionChanges:    versionChanges,
		EnvironmentChange: envChange,
		FieldsChanged:     fieldPaths,
		FieldChangeCount:  len(patches),
	}
}
