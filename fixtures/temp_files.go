package fixtures

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TempFileInfo holds information about a temporary file
type TempFileInfo struct {
	Path      string
	Content   string
	Extension string
	Detected  string // Detected file type (would use libmagic)
}

// GetTemplateData returns data for gomplate templates
func (t *TempFileInfo) GetTemplateData() map[string]interface{} {
	return map[string]interface{}{
		"path":     t.Path,
		"content":  t.Content,
		"ext":      t.Extension,
		"detected": t.Detected,
	}
}

// GetCELData returns data for CEL evaluation
func (t *TempFileInfo) GetCELData() map[string]interface{} {
	data := t.GetTemplateData()

	// Try to parse as JSON if it looks like JSON
	if strings.HasPrefix(strings.TrimSpace(t.Content), "{") || strings.HasPrefix(strings.TrimSpace(t.Content), "[") {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(t.Content), &jsonData); err == nil {
			data["json"] = jsonData
		}
	}

	return data
}

// createTempFile creates a temporary file with the given content
func createTempFile(name, content string) (*TempFileInfo, error) {
	// Determine extension from name
	ext := filepath.Ext(name)
	if ext == "" {
		ext = ".tmp"
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("fixture-%s-*%s", name, ext))
	if err != nil {
		return nil, err
	}

	// Write content
	if _, err := tmpFile.WriteString(content); err != nil {
		_ = tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, err
	}
	_ = tmpFile.Close()

	// Detect file type (simplified - would use libmagic in production)
	detected := "text"
	if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
		detected = "json"
	} else if strings.HasPrefix(content, "<?xml") {
		detected = "xml"
	} else if strings.HasPrefix(content, "---\n") {
		detected = "yaml"
	}

	return &TempFileInfo{
		Path:      tmpFile.Name(),
		Content:   content,
		Extension: ext,
		Detected:  detected,
	}, nil
}
