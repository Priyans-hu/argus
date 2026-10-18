package merger

import (
	"regexp"
	"strings"
)

// Section markers for identifying auto-generated vs custom content
const (
	// AutoStartMarker marks the beginning of auto-generated content
	AutoStartMarker = "<!-- ARGUS:AUTO -->"
	// AutoEndMarker marks the end of auto-generated content
	AutoEndMarker = "<!-- /ARGUS:AUTO -->"
	// CustomStartMarker marks the beginning of user's custom content
	CustomStartMarker = "<!-- ARGUS:CUSTOM -->"
	// CustomEndMarker marks the end of user's custom content
	CustomEndMarker = "<!-- /ARGUS:CUSTOM -->"
)

// Merger handles merging of generated content with existing files
type Merger struct {
	preserveCustom bool
}

// NewMerger creates a new merger
func NewMerger(preserveCustom bool) *Merger {
	return &Merger{
		preserveCustom: preserveCustom,
	}
}

// Outcome tells the caller what Merge did with the existing file.
type Outcome int

const (
	// Created means the existing file was missing or blank: the result is the AUTO block alone.
	Created Outcome = iota
	// Replaced means the single AUTO block was replaced in place.
	Replaced
	// Adopted means the file had no AUTO markers: the block was put on top of its text.
	Adopted
	// Malformed means the AUTO markers are unbalanced or repeated: existing is returned unchanged.
	Malformed
)

// Merge puts newContent, wrapped in AUTO markers, into existingContent. A
// single AUTO block is replaced in place and every byte outside it is kept;
// a file without AUTO markers gets the block on top of its text. With
// preserveCustom off the result is the wrapped block alone.
func (m *Merger) Merge(existingContent, newContent []byte) ([]byte, Outcome) {
	block := wrapWithMarkers(string(newContent))
	if !m.preserveCustom {
		return []byte(block), Replaced
	}

	existing := string(existingContent)
	if strings.TrimSpace(existing) == "" {
		return []byte(block), Created
	}

	starts := strings.Count(existing, AutoStartMarker)
	ends := strings.Count(existing, AutoEndMarker)
	if starts == 0 && ends == 0 {
		return []byte(block + "\n\n" + strings.TrimLeft(existing, "\r\n")), Adopted
	}
	start := strings.Index(existing, AutoStartMarker)
	end := strings.Index(existing, AutoEndMarker)
	if starts != 1 || ends != 1 || end < start {
		return existingContent, Malformed
	}
	return []byte(existing[:start] + block + existing[end+len(AutoEndMarker):]), Replaced
}

// AddCustomSectionPlaceholder adds a custom section placeholder to content
func AddCustomSectionPlaceholder(content string) string {
	placeholder := `

` + CustomStartMarker + `
## Custom Notes

Add your custom documentation here. This section will be preserved when regenerating.

` + CustomEndMarker

	return content + placeholder
}

// WrapContent wraps content with auto markers
func WrapContent(content string) string {
	return wrapWithMarkers(content)
}

// wrapWithMarkers wraps content with auto-generated markers
func wrapWithMarkers(content string) string {
	// Remove trailing whitespace from content
	content = strings.TrimRight(content, "\n\t ")

	return AutoStartMarker + "\n" + content + "\n" + AutoEndMarker
}

// extractCustomSections extracts all custom sections from content
func extractCustomSections(content string) []string {
	var sections []string

	// Regex to find custom sections
	re := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(CustomStartMarker) + `(.*?)` + regexp.QuoteMeta(CustomEndMarker))
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			// Reconstruct the full section with markers
			section := CustomStartMarker + match[1] + CustomEndMarker
			sections = append(sections, section)
		}
	}

	return sections
}

// HasCustomContent checks if existing content has custom sections worth preserving
func HasCustomContent(content string) bool {
	sections := extractCustomSections(content)
	for _, section := range sections {
		// Check if section has meaningful content (not just placeholder)
		inner := strings.TrimPrefix(section, CustomStartMarker)
		inner = strings.TrimSuffix(inner, CustomEndMarker)
		inner = strings.TrimSpace(inner)

		// Skip if it's just the default placeholder
		if strings.Contains(inner, "Add your custom documentation here") {
			continue
		}

		// Has real custom content
		if len(inner) > 0 {
			return true
		}
	}
	return false
}

// StripMarkers removes all Argus markers from content
func StripMarkers(content string) string {
	content = strings.ReplaceAll(content, AutoStartMarker, "")
	content = strings.ReplaceAll(content, AutoEndMarker, "")
	content = strings.ReplaceAll(content, CustomStartMarker, "")
	content = strings.ReplaceAll(content, CustomEndMarker, "")
	return strings.TrimSpace(content)
}
