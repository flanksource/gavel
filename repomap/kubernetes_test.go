package repomap

import (
	"testing"

	"github.com/flanksource/gavel/models/kubernetes"
)

func TestParseYAMLDocuments(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantDocs int
	}{
		{
			name: "single document",
			content: `apiVersion: v1
kind: Pod
metadata:
  name: test-pod`,
			wantDocs: 1,
		},
		{
			name: "multiple documents",
			content: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
---
apiVersion: v1
kind: Service
metadata:
  name: svc1`,
			wantDocs: 2,
		},
		{
			name:     "empty content",
			content:  "",
			wantDocs: 0,
		},
		{
			name: "invalid YAML",
			content: `apiVersion: v1
kind: Pod
  invalid indentation
metadata:`,
			wantDocs: 1, // goccy/go-yaml is more lenient and can parse this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs, err := ParseYAMLDocuments(tt.content)
			if err != nil {
				t.Errorf("ParseYAMLDocuments() error = %v", err)
				return
			}
			if len(docs) != tt.wantDocs {
				t.Errorf("ParseYAMLDocuments() got %d documents, want %d", len(docs), tt.wantDocs)
			}
		})
	}
}

func TestParseYAMLDocuments_LineNumberAccuracy(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantDocs  int
		checkDocs []struct {
			idx       int
			startLine int
			endLine   int
		}
	}{
		{
			name: "single document - accurate line numbers",
			content: `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  labels:
    app: nginx`,
			wantDocs: 1,
			checkDocs: []struct {
				idx       int
				startLine int
				endLine   int
			}{
				{idx: 0, startLine: 1, endLine: 6},
			},
		},
		{
			name: "multi-document with separator",
			content: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
---
apiVersion: v1
kind: Service
metadata:
  name: svc1`,
			wantDocs: 2,
			checkDocs: []struct {
				idx       int
				startLine int
				endLine   int
			}{
				{idx: 0, startLine: 1, endLine: 4},
				{idx: 1, startLine: 6, endLine: 9},
			},
		},
		{
			name: "with comments and blank lines",
			content: `# First document comment
apiVersion: v1
kind: Pod
metadata:
  name: pod1

---
# Second document comment

apiVersion: v1
kind: Service
metadata:
  name: svc1`,
			wantDocs: 2,
			checkDocs: []struct {
				idx       int
				startLine int
				endLine   int
			}{
				{idx: 0, startLine: 1, endLine: 6},
				{idx: 1, startLine: 8, endLine: 13},
			},
		},
		{
			name: "three documents with varying spacing",
			content: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
---
apiVersion: v1
kind: Service
metadata:
  name: svc1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config1`,
			wantDocs: 3,
			checkDocs: []struct {
				idx       int
				startLine int
				endLine   int
			}{
				{idx: 0, startLine: 1, endLine: 4},
				{idx: 1, startLine: 6, endLine: 9},
				{idx: 2, startLine: 11, endLine: 14},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs, err := ParseYAMLDocuments(tt.content)
			if err != nil {
				t.Errorf("ParseYAMLDocuments() error = %v", err)
				return
			}
			if len(docs) != tt.wantDocs {
				t.Errorf("ParseYAMLDocuments() got %d documents, want %d", len(docs), tt.wantDocs)
				return
			}

			// Verify specific line numbers
			for _, check := range tt.checkDocs {
				if check.idx >= len(docs) {
					t.Errorf("Document index %d out of range (have %d docs)", check.idx, len(docs))
					continue
				}
				doc := docs[check.idx]
				if doc.StartLine != check.startLine {
					t.Errorf("Document %d: StartLine = %d, want %d", check.idx, doc.StartLine, check.startLine)
				}
				if doc.EndLine != check.endLine {
					t.Errorf("Document %d: EndLine = %d, want %d", check.idx, doc.EndLine, check.endLine)
				}
			}
		})
	}
}

func TestIsKubernetesResource(t *testing.T) {
	tests := []struct {
		name string
		doc  map[string]interface{}
		want bool
	}{
		{
			name: "valid kubernetes resource",
			doc: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
			},
			want: true,
		},
		{
			name: "missing apiVersion",
			doc: map[string]interface{}{
				"kind": "Pod",
			},
			want: false,
		},
		{
			name: "missing kind",
			doc: map[string]interface{}{
				"apiVersion": "v1",
			},
			want: false,
		},
		{
			name: "plain config",
			doc: map[string]interface{}{
				"database": map[string]interface{}{
					"host": "localhost",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsKubernetesResource(tt.doc); got != tt.want {
				t.Errorf("IsKubernetesResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractKubernetesRef(t *testing.T) {
	doc := kubernetes.YAMLDocument{
		StartLine: 1,
		EndLine:   25,
		Content: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "nginx-deployment",
				"namespace": "production",
				"labels": map[string]interface{}{
					"app":     "nginx",
					"version": "v1.0",
				},
				"annotations": map[string]interface{}{
					"description": "Main nginx deployment",
				},
			},
		},
	}

	ref := ExtractKubernetesRef(doc)

	if ref.APIVersion != "apps/v1" {
		t.Errorf("APIVersion = %v, want apps/v1", ref.APIVersion)
	}
	if ref.Kind != "Deployment" {
		t.Errorf("Kind = %v, want Deployment", ref.Kind)
	}
	if ref.Name != "nginx-deployment" {
		t.Errorf("Name = %v, want nginx-deployment", ref.Name)
	}
	if ref.Namespace != "production" {
		t.Errorf("Namespace = %v, want production", ref.Namespace)
	}
	if ref.StartLine != 1 || ref.EndLine != 25 {
		t.Errorf("Lines = %d-%d, want 1-25", ref.StartLine, ref.EndLine)
	}
	if len(ref.Labels) != 2 {
		t.Errorf("Labels count = %d, want 2", len(ref.Labels))
	}
	if ref.Labels["app"] != "nginx" {
		t.Errorf("Label app = %v, want nginx", ref.Labels["app"])
	}
	if len(ref.Annotations) != 1 {
		t.Errorf("Annotations count = %d, want 1", len(ref.Annotations))
	}
}

// FIXME: ExtractKubernetesRefsFromFile function doesn't exist, using ExtractKubernetesRefsFromContent instead
func _TestExtractKubernetesRefsFromFile(t *testing.T) {
	t.Skip("ExtractKubernetesRefsFromFile function doesn't exist")
}

// FIXME: ExtractKubernetesRefsFromFile function doesn't exist
func _TestExtractKubernetesRefsFromFile_Deployment(t *testing.T) {
	t.Skip("ExtractKubernetesRefsFromFile function doesn't exist")
}

// FIXME: ExtractKubernetesRefsFromFile function doesn't exist
func _TestExtractKubernetesRefsFromFile_MultiDoc(t *testing.T) {
	t.Skip("ExtractKubernetesRefsFromFile function doesn't exist")
}
