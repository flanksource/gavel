package jsonlint

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	gavelModeline   = regexp.MustCompile(`^#\s*gavel:`)
	modelinePreview = regexp.MustCompile(`\bpreview\s*=\s*("[^"]+"|'[^']+'|\S+)`)
	yamlPath        = regexp.MustCompile(`(?i)\.ya?ml$`)
)

// PreviewDeclaration is the component a data file names to render it: a YAML
// `# gavel: preview=<path>` header, else a top-level `$preview` key, resolved
// against the file's folder to a URL. It is empty when the file names none.
func PreviewDeclaration(data []byte, path string) (string, error) {
	reference := ""
	if yamlPath.MatchString(path) {
		header, err := previewHeader(data)
		if err != nil {
			return "", err
		}
		reference = header
	}
	if reference == "" {
		var value any
		if err := yaml.Unmarshal(data, &value); err != nil {
			return "", fmt.Errorf("read the $preview declaration of %s: %w", filepath.Base(path), yamlSourceError(err, data, path))
		}
		object, _ := value.(map[string]any)
		if declared, ok := object["$preview"]; ok {
			text, isString := declared.(string)
			if !isString || strings.TrimSpace(text) == "" {
				return "", fmt.Errorf("$preview must be a nonempty path to a preview component")
			}
			reference = text
		}
	}
	if reference == "" {
		return "", nil
	}
	parsed, err := url.Parse(reference)
	if err != nil {
		return "", fmt.Errorf("preview %s: %w", reference, err)
	}
	base := &url.URL{Scheme: "file", Path: filepath.ToSlash(filepath.Dir(path)) + "/"}
	return base.ResolveReference(parsed).String(), nil
}

// previewHeader is the preview= of the first `# gavel:` line among the
// leading comments.
func previewHeader(data []byte) (string, error) {
	for _, line := range bytes.Split(data, []byte("\n")) {
		text := strings.TrimSpace(strings.TrimPrefix(string(line), "\ufeff"))
		if text == "" {
			continue
		}
		if !strings.HasPrefix(text, "#") {
			return "", nil
		}
		if !gavelModeline.MatchString(text) {
			continue
		}
		match := modelinePreview.FindStringSubmatch(text)
		if match == nil || strings.TrimSpace(strings.Trim(match[1], "\"'")) == "" {
			return "", fmt.Errorf("the `# gavel:` header requires a nonempty preview=<path>")
		}
		return strings.Trim(match[1], "\"'"), nil
	}
	return "", nil
}

// SchemaPreview is the x-clicky-preview ref of the root schema of an
// EditorSchema bundle, which EditorSchema resolved against the schema that
// declares it. It is empty for no bundle or no declaration.
func SchemaPreview(bundle map[string]any) (string, error) {
	if bundle == nil {
		return "", nil
	}
	ref, _ := bundle["$ref"].(string)
	name, ok := strings.CutPrefix(ref, "#/$defs/")
	if !ok || strings.Contains(name, "/") {
		return "", nil
	}
	defs, _ := bundle["$defs"].(map[string]any)
	root, _ := defs[name].(map[string]any)
	declared, ok := root["x-clicky-preview"]
	if !ok {
		return "", nil
	}
	source, _ := declared.(map[string]any)
	component, isString := source["ref"].(string)
	if !isString || strings.TrimSpace(component) == "" {
		return "", fmt.Errorf("x-clicky-preview must be { ref } naming a preview component")
	}
	return component, nil
}

// resolvePreviewRef makes a schema's x-clicky-preview ref absolute against the
// schema that declares it, as resolveExtensionRefs does for form extensions.
func resolvePreviewRef(value any, base *url.URL) error {
	source, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("x-clicky-preview must be { ref } naming a preview component")
	}
	ref, ok := source["ref"].(string)
	if !ok {
		return fmt.Errorf("x-clicky-preview must be { ref } naming a preview component")
	}
	parsed, err := url.Parse(ref)
	if err != nil {
		return fmt.Errorf("x-clicky-preview.ref: %w", err)
	}
	source["ref"] = base.ResolveReference(parsed).String()
	return nil
}
