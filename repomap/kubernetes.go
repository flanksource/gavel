package repomap

import (
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/models/kubernetes"
	"github.com/goccy/go-yaml"
)

func IsYaml(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}

// documentBoundary represents a YAML document's position in source
type documentBoundary struct {
	content   string
	startLine int
	endLine   int
}

// splitYAMLDocuments splits YAML content by --- separators while tracking actual source line numbers
func splitYAMLDocuments(content string) []documentBoundary {
	var boundaries []documentBoundary
	lines := strings.Split(content, "\n")

	var currentDoc []string
	docStartLine := 1

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Document separator
		if trimmed == "---" {
			// Save previous document if it has content
			if len(currentDoc) > 0 {
				boundaries = append(boundaries, documentBoundary{
					content:   strings.Join(currentDoc, "\n"),
					startLine: docStartLine,
					endLine:   i, // Line before separator
				})
			}
			// Start new document after separator
			currentDoc = []string{}
			docStartLine = lineNum + 1
		} else {
			currentDoc = append(currentDoc, line)
		}
	}

	// Add final document
	if len(currentDoc) > 0 {
		boundaries = append(boundaries, documentBoundary{
			content:   strings.Join(currentDoc, "\n"),
			startLine: docStartLine,
			endLine:   len(lines),
		})
	}

	return boundaries
}

// ParseYAMLDocuments parses a YAML file that may contain multiple documents
// and returns each document with its line boundaries from the actual source
func ParseYAMLDocuments(content string) ([]kubernetes.YAMLDocument, error) {
	boundaries := splitYAMLDocuments(content)
	var documents []kubernetes.YAMLDocument

	for _, boundary := range boundaries {
		var doc map[string]interface{}
		err := yaml.Unmarshal([]byte(boundary.content), &doc)
		if err != nil {
			// Skip malformed documents silently
			continue
		}

		if doc == nil {
			continue
		}

		documents = append(documents, kubernetes.YAMLDocument{
			StartLine: boundary.startLine,
			EndLine:   boundary.endLine,
			Content:   doc,
		})
	}

	return documents, nil
}

// IsKubernetesResource checks if a YAML document is a Kubernetes resource
func IsKubernetesResource(doc map[string]interface{}) bool {
	_, hasAPIVersion := doc["apiVersion"]
	_, hasKind := doc["kind"]
	return hasAPIVersion && hasKind
}

// ExtractKubernetesRef creates a KubernetesRef from a parsed YAML document
func ExtractKubernetesRef(doc kubernetes.YAMLDocument) kubernetes.KubernetesRef {
	ref := kubernetes.KubernetesRef{
		StartLine: doc.StartLine,
		EndLine:   doc.EndLine,
	}

	// Extract API version and kind
	if apiVersion, ok := doc.Content["apiVersion"].(string); ok {
		ref.APIVersion = apiVersion
	}
	if kind, ok := doc.Content["kind"].(string); ok {
		ref.Kind = kind
	}

	// Extract metadata
	if metadata, ok := doc.Content["metadata"].(map[string]interface{}); ok {
		if namespace, ok := metadata["namespace"].(string); ok {
			ref.Namespace = namespace
		}
		if name, ok := metadata["name"].(string); ok {
			ref.Name = name
		}

		// Extract labels
		if labels, ok := metadata["labels"].(map[string]interface{}); ok {
			ref.Labels = make(map[string]string)
			for k, v := range labels {
				if strVal, ok := v.(string); ok {
					ref.Labels[k] = strVal
				}
			}
		}

		// Extract annotations
		if annotations, ok := metadata["annotations"].(map[string]interface{}); ok {
			ref.Annotations = make(map[string]string)
			for k, v := range annotations {
				if strVal, ok := v.(string); ok {
					ref.Annotations[k] = strVal
				}
			}
		}
	}

	return ref
}

// ExtractKubernetesRefsFromFile reads a YAML file and extracts Kubernetes resource references
func ExtractKubernetesRefsFromContent(content string) ([]kubernetes.KubernetesRef, error) {

	// Parse YAML documents
	documents, err := ParseYAMLDocuments(string(content))
	if err != nil {
		return nil, err
	}

	// Extract Kubernetes refs from each document
	var refs []kubernetes.KubernetesRef
	for _, doc := range documents {
		// Only process Kubernetes resources (have apiVersion and kind)
		if !IsKubernetesResource(doc.Content) {
			continue
		}

		ref := ExtractKubernetesRef(doc)
		refs = append(refs, ref)
	}

	return refs, nil
}
