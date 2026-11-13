package fileread

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
)

var imageTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".avif": "image/avif",
	".bmp":  "image/bmp",
	".ico":  "image/x-icon",
}

// imageCSP stops a same-origin svg opened directly from running script.
const imageCSP = "default-src 'none'; style-src 'unsafe-inline'; sandbox"

// ContentType is the image MIME type for path's extension, and false when the
// extension is not an image.
func ContentType(path string) (string, bool) {
	contentType, ok := imageTypes[strings.ToLower(filepath.Ext(path))]
	return contentType, ok
}

// ImageExtensions lists the supported image extensions, with the leading dot, sorted.
func ImageExtensions() []string {
	exts := make([]string, 0, len(imageTypes))
	for ext := range imageTypes {
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	return exts
}

// WriteImage serves data as the image named by path, with headers that keep a
// browser from sniffing it as something else or running script inside it.
func WriteImage(w http.ResponseWriter, path string, data []byte) error {
	contentType, ok := ContentType(path)
	if !ok {
		return fmt.Errorf("%w: %q", ErrNotImage, path)
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", imageCSP)
	_, err := w.Write(data)
	return err
}
