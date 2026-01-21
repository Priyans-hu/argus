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

// Section represents a section in the document
type Section struct {
	Type    string // "auto", "custom", or "unknown"
	Content string
	Name    string // Optional name for the section
}

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

// Merge combines new auto-generated content with existing custom sections
func (m *Merger) Merge(existingContent, newContent []byte) []byte {
	if !m.preserveCustom {
		return newContent
	}

	existing := string(existingContent)
	new := string(newContent)

	// If existing file has no markers, preserve entire content as custom
	if !hasMarkers(existing) {
		// If existing file has content, treat it all as user's custom content
		if len(strings.TrimSpace(existing)) > 0 {
			result := wrapWithMarkers(new)
			result += "\n\n" + CustomStartMarker + "\n## Previous Content\n\n"
			result += "The following content was preserved from your original file:\n\n"
			result += strings.TrimSpace(existing) + "\n" + CustomEndMarker
			return []byte(result)
		}
		return []byte(wrapWithMarkers(new))
	}

	// Extract custom sections from existing content
	customSections := extractCustomSections(existing)

	// If no custom sections found, just return new content with markers
	if len(customSections) == 0 {
		return []byte(wrapWithMarkers(new))
	}

	// Wrap new content with auto markers and append custom sections
	result := wrapWithMarkers(new)

	// Append each custom section
	for _, section := range customSections {
		result += "\n\n" + section
	}

	return []byte(result)
}

// MergeWithSections performs a more sophisticated merge, preserving named sections
func (m *Merger) MergeWithSections(existingContent, newContent []byte) []byte {
	if !m.preserveCustom {
		return newContent
	}

	existing := string(existingContent)
	new := string(newContent)

	// Parse existing content into sections
	existingSections := parseSections(existing)

	// Find custom sections
	var customSections []Section
	for _, section := range existingSections {
		if section.Type == "custom" {
			customSections = append(customSections, section)
		}
	}

	// If no custom sections, just wrap new content
	if len(customSections) == 0 {
		return []byte(wrapWithMarkers(new))
	}

	// Build result with auto content and preserved custom sections
	result := wrapWithMarkers(new)

	for _, section := range customSections {
		result += "\n\n" + formatCustomSection(section)
	}

	return []byte(result)
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

// hasMarkers checks if content has any Argus markers
func hasMarkers(content string) bool {
	return strings.Contains(content, AutoStartMarker) ||
		strings.Contains(content, AutoEndMarker) ||
		strings.Contains(content, CustomStartMarker) ||
		strings.Contains(content, CustomEndMarker)
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

// parseSections parses content into sections
func parseSections(content string) []Section {
	var sections []Section

	// Find auto sections
	autoRe := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(AutoStartMarker) + `(.*?)` + regexp.QuoteMeta(AutoEndMarker))
	customRe := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(CustomStartMarker) + `(.*?)` + regexp.QuoteMeta(CustomEndMarker))

	// Extract auto sections
	autoMatches := autoRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range autoMatches {
		if len(match) >= 4 {
			sections = append(sections, Section{
				Type:    "auto",
				Content: content[match[2]:match[3]],
			})
		}
	}

	// Extract custom sections
	customMatches := customRe.FindAllStringSubmatch(content, -1)
	for _, match := range customMatches {
		if len(match) >= 2 {
			// Try to extract section name from first heading
			name := extractSectionName(match[1])
			sections = append(sections, Section{
				Type:    "custom",
				Content: match[1],
				Name:    name,
			})
		}
	}

	return sections
}

// extractSectionName tries to extract a section name from content
func extractSectionName(content string) string {
	// Look for first heading
	re := regexp.MustCompile(`(?m)^##\s+(.+)$`)
	match := re.FindStringSubmatch(content)
	if len(match) >= 2 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

// formatCustomSection formats a custom section for output
func formatCustomSection(section Section) string {
	return CustomStartMarker + section.Content + CustomEndMarker
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
