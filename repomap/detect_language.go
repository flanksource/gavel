package repomap

func detectLanguage(filePath string) string {
	if len(filePath) < 3 {
		return "unknown"
	}

	switch {
	case filePath[len(filePath)-3:] == ".go":
		return "go"
	case filePath[len(filePath)-3:] == ".py" || (len(filePath) >= 4 && filePath[len(filePath)-4:] == ".pyi"):
		return "python"
	case len(filePath) >= 5 && filePath[len(filePath)-5:] == ".java":
		return "java"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".tsx":
		return "typescript"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".ts":
		return "typescript"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".jsx":
		return "javascript"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".js":
		return "javascript"
	case len(filePath) >= 4 && (filePath[len(filePath)-4:] == ".mjs" || filePath[len(filePath)-4:] == ".cjs"):
		return "javascript"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".rs":
		return "rust"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".rb":
		return "ruby"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".md":
		return "markdown"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".mdx":
		return "markdown"
	case len(filePath) >= 9 && filePath[len(filePath)-9:] == ".markdown":
		return "markdown"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".xml":
		return "xml"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".xsd":
		return "xml"
	case len(filePath) >= 5 && filePath[len(filePath)-5:] == ".xslt":
		return "xml"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".xsl":
		return "xml"
	default:
		return "unknown"
	}
}
