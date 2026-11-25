// Package linecounts identifies files whose changes have meaningful line counts.
package linecounts

import (
	"path/filepath"
	"strings"
)

// Supported excludes assets even when Git happens to treat their bytes as text.
// Unknown extensions remain eligible, as do documentation and configuration.
// Binary detection is separate: a source extension does not guarantee text.
func Supported(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".icns", ".webp", ".avif", ".tif", ".tiff", ".svg", ".svgz", ".psd":
	case ".pdf", ".ps", ".eps", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".odt", ".ods", ".odp":
	case ".zip", ".gz", ".bz2", ".xz", ".zst", ".tar", ".tgz", ".7z", ".rar", ".jar", ".war":
	case ".woff", ".woff2", ".ttf", ".otf", ".eot":
	case ".mp3", ".mp4", ".m4a", ".mov", ".avi", ".webm", ".wav", ".ogg", ".flac":
	case ".exe", ".dll", ".so", ".dylib", ".a", ".o", ".obj", ".class", ".pyc", ".wasm", ".bin":
	case ".db", ".sqlite", ".sqlite3":
		return false
	default:
		return true
	}
	return false
}
