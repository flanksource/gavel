package jira

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func validateDOCXMedia(path string) (returnErr error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("read DOCX archive: %w", err)
	}
	defer func() { returnErr = errors.Join(returnErr, archive.Close()) }()
	relations, err := archive.Open("word/_rels/document.xml.rels")
	if err != nil {
		return fmt.Errorf("read DOCX relationships: %w", err)
	}
	defer func() { returnErr = errors.Join(returnErr, relations.Close()) }()
	var document struct {
		Relationships []struct {
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	if err := xml.NewDecoder(relations).Decode(&document); err != nil {
		return fmt.Errorf("decode DOCX relationships: %w", err)
	}
	for _, relationship := range document.Relationships {
		target := relationship.Target
		if strings.HasPrefix(target, "media/") && (strings.Contains(target, "\\") || !strings.HasPrefix(filepath.ToSlash(filepath.Clean(target)), "media/")) {
			return fmt.Errorf("unsafe DOCX image target %q", target)
		}
	}
	return nil
}
