package editor

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strings"

	jsonlint "github.com/flanksource/gavel/linters/json"
)

var dataDocument = regexp.MustCompile(`(?i)\.(json|ya?ml)$`)

// previewComponentExtensions are the pages Facet renders a data file through.
var previewComponentExtensions = []string{".tsx", ".jsx", ".mdx"}

// PreviewComponent is the component that renders a JSON or YAML document's
// source, as a path in the document's preview root: the file's own `# gavel:
// preview=` header or `$preview` key, else its schema's x-clicky-preview. It is
// empty when neither names one.
func (s *Server) PreviewComponent(ctx context.Context, id, source string) (string, error) {
	doc, err := s.Store.Read(id)
	if err != nil {
		return "", err
	}
	if !dataDocument.MatchString(doc.Path) {
		return "", fmt.Errorf("%s is not a JSON or YAML document; only data files preview through a component", doc.Path)
	}
	ref, err := jsonlint.PreviewDeclaration([]byte(source), doc.Path)
	if err != nil {
		return "", err
	}
	if ref == "" {
		schema, err := s.editorSchema(ctx, doc, source)
		if err != nil {
			return "", err
		}
		if ref, err = jsonlint.SchemaPreview(schema); err != nil {
			return "", err
		}
	}
	if ref == "" {
		return "", nil
	}
	root, _, err := s.Store.previewRoot(id)
	if err != nil {
		return "", err
	}
	preview, err := s.Store.Preview(id)
	if err != nil {
		return "", err
	}
	component, err := previewSourcePath(ref, root, preview.Paths, "preview component")
	if err != nil {
		return "", err
	}
	if !slices.Contains(previewComponentExtensions, strings.ToLower(path.Ext(component))) {
		return "", fmt.Errorf("preview component %s must be one of %s", component, strings.Join(previewComponentExtensions, ", "))
	}
	return component, nil
}

// previewComponent serves POST /api/file/{id}/preview/component {source}.
func (s *Server) previewComponent(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Source string `json:"source"`
	}
	if !decode(w, r, &input) {
		return
	}
	component, err := s.PreviewComponent(r.Context(), r.PathValue("id"), input.Source)
	respond(w, struct {
		Component string `json:"component"`
	}{component}, err)
}
