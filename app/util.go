package app

import "path/filepath"

func GetLanguage(filename string) string {
	extension := filepath.Ext(filename)
	switch extension {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".java":
		return "java"
	case ".cpp":
		return "cpp"
	case ".c":
		return "c"
	case ".rs":
		return "rust"
	case ".rb":
		return "ruby"
	}
	return "text"
}
