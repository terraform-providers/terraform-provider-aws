package check

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	FileExtensionErb          = `.erb`
	FileExtensionHtmlMarkdown = `.html.markdown`
	FileExtensionHtmlMd       = `.html.md`
	FileExtensionMarkdown     = `.markdown`
	FileExtensionMd           = `.md`
)

var ValidLegacyFileExtensions = []string{
	FileExtensionHtmlMarkdown,
	FileExtensionHtmlMd,
	FileExtensionMarkdown,
	FileExtensionMd,
}

var ValidRegistryFileExtensions = []string{
	FileExtensionMd,
}

func LegacyFileExtensionCheck(path string) error {
	if !FilePathEndsWithExtensionFrom(path, ValidLegacyFileExtensions) {
		return fmt.Errorf("file does not end with a valid extension, valid extensions: %v", ValidLegacyFileExtensions)
	}

	return nil
}

func RegistryFileExtensionCheck(path string) error {
	if !FilePathEndsWithExtensionFrom(path, ValidRegistryFileExtensions) {
		return fmt.Errorf("file does not end with a valid extension, valid extensions: %v", ValidLegacyFileExtensions)
	}

	return nil
}

func FilePathEndsWithExtensionFrom(path string, validExtensions []string) bool {
	for _, validExtension := range validExtensions {
		if strings.HasSuffix(path, validExtension) {
			return true
		}
	}

	return false
}

// TrimFileExtension removes file extensions including those with multiple periods.
func TrimFileExtension(path string) string {
	filename := filepath.Base(path)

	if filename == "." {
		return ""
	}

	dotIndex := strings.IndexByte(filename, '.')

	if dotIndex > 0 {
		return filename[:dotIndex]
	}

	return filename
}
