package jira

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

func validateDocumentSVG(data []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	depth, roots := 0, 0
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			if depth != 0 || roots != 1 {
				return fmt.Errorf("SVG requires one complete root element")
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("invalid SVG XML: %w", err)
		}
		switch element := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if element.Name.Local != "svg" || element.Name.Space != "http://www.w3.org/2000/svg" {
					return fmt.Errorf("invalid SVG root %v", element.Name)
				}
			}
			depth++
			if element.Name.Local == "script" || element.Name.Local == "foreignObject" {
				return fmt.Errorf("unsupported SVG element %q", element.Name.Local)
			}
			for _, attr := range element.Attr {
				name := strings.ToLower(attr.Name.Local)
				if strings.HasPrefix(name, "on") || (name == "href" && !strings.HasPrefix(attr.Value, "#") && !strings.HasPrefix(attr.Value, "data:image/")) {
					return fmt.Errorf("unsupported SVG attribute %q", attr.Name.Local)
				}
			}
		case xml.EndElement:
			depth--
		case xml.Directive:
			return fmt.Errorf("unsupported SVG XML directive")
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(element)) != "" {
				return fmt.Errorf("text outside SVG root")
			}
		}
	}
}
