package app

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/quick"
)

func Highlight(content, language string) string {
	content = strings.TrimSpace(strings.Replace(content, "\t", "    ", -1))
	var b bytes.Buffer
	err := quick.Highlight(&b, content, language, "terminal16m", "dracula")
	if err != nil {
		return content
	}
	return b.String()
}
