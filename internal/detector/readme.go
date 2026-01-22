package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// ReadmeDetector parses README files for project information
type ReadmeDetector struct {
	rootPath string
}

// NewReadmeDetector creates a new README detector
func NewReadmeDetector(rootPath string) *ReadmeDetector {
	return &ReadmeDetector{rootPath: rootPath}
}

// Detect parses README and extracts project information
func (d *ReadmeDetector) Detect() *types.ReadmeContent {
	// Try common README filenames
	readmeNames := []string{
		"README.md",
		"readme.md",
		"Readme.md",
		"README.MD",
		"README",
		"readme",
	}

	var content []byte
	var err error
	for _, name := range readmeNames {
		content, err = os.ReadFile(filepath.Join(d.rootPath, name))
		if err == nil {
			break
		}
	}

	if content == nil {
		return nil
	}

	return d.parseReadme(string(content))
}

// parseReadme extracts structured information from README content
func (d *ReadmeDetector) parseReadme(content string) *types.ReadmeContent {
	result := &types.ReadmeContent{}

	lines := strings.Split(content, "\n")

	// Extract title (first # heading)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			result.Title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Extract description (first non-empty paragraph after title)
	result.Description = d.extractDescription(lines)

	// Extract sections
	sections := d.parseSections(content)

	// Features section
	if features, ok := sections["features"]; ok {
		result.Features = d.extractBulletPoints(features)
	}
	// Also try "what it does", "highlights"
	if len(result.Features) == 0 {
		if features, ok := sections["what it does"]; ok {
			result.Features = d.extractBulletPoints(features)
		}
	}
	if len(result.Features) == 0 {
		if features, ok := sections["highlights"]; ok {
			result.Features = d.extractBulletPoints(features)
		}
	}

	// Installation section
	if install, ok := sections["installation"]; ok {
		result.Installation = d.cleanSection(install)
	}
	if result.Installation == "" {
		if install, ok := sections["install"]; ok {
			result.Installation = d.cleanSection(install)
		}
	}
	if result.Installation == "" {
		if install, ok := sections["setup"]; ok {
			result.Installation = d.cleanSection(install)
		}
	}

	// Quick start section
	if qs, ok := sections["quick start"]; ok {
		result.QuickStart = d.cleanSection(qs)
	}
	if result.QuickStart == "" {
		if qs, ok := sections["quickstart"]; ok {
			result.QuickStart = d.cleanSection(qs)
		}
	}
	if result.QuickStart == "" {
		if qs, ok := sections["getting started"]; ok {
			result.QuickStart = d.cleanSection(qs)
		}
	}

	// Usage section
	if usage, ok := sections["usage"]; ok {
		result.Usage = d.cleanSection(usage)
	}
	if result.Usage == "" {
		if usage, ok := sections["how to use"]; ok {
			result.Usage = d.cleanSection(usage)
		}
	}

	return result
}

// extractDescription gets the first paragraph after the title
func (d *ReadmeDetector) extractDescription(lines []string) string {
	foundTitle := false
	var descLines []string
	inDescription := false
	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		original := line

		// Skip until we find the title (markdown or HTML)
		if !foundTitle {
			if strings.HasPrefix(trimmed, "# ") ||
				strings.Contains(trimmed, "<h1") {
				foundTitle = true
			}
			continue
		}

		// Track code blocks (``` delimiters)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// Skip content inside code blocks
		if inCodeBlock {
			continue
		}

		// Skip indented code blocks (4 spaces or 1 tab)
		if len(original) > 0 && (strings.HasPrefix(original, "    ") || strings.HasPrefix(original, "\t")) {
			continue
		}

		// Skip badges, empty lines at start
		if !inDescription {
			if trimmed == "" {
				continue
			}
			// Skip HTML tags that are not description
			if strings.HasPrefix(trimmed, "<p align=") || strings.HasPrefix(trimmed, "<img") ||
				strings.HasPrefix(trimmed, "<a href=") || strings.HasPrefix(trimmed, "</") ||
				trimmed == "</p>" || trimmed == "</a>" {
				// Check if this line contains actual text content (not just HTML)
				textContent := d.extractTextFromHTML(trimmed)
				if textContent != "" && !strings.Contains(trimmed, "shields.io") && !strings.Contains(trimmed, "<img") {
					descLines = append(descLines, textContent)
					inDescription = true
				}
				continue
			}
			// Skip badge lines (contain shields.io, badge, etc.)
			if strings.Contains(trimmed, "shields.io") ||
				strings.Contains(trimmed, "badge") ||
				strings.Contains(trimmed, "![") {
				continue
			}
			// Skip HTML comments
			if strings.HasPrefix(trimmed, "<!--") {
				continue
			}
			// Skip blockquotes
			if strings.HasPrefix(trimmed, ">") {
				continue
			}
			// Skip next heading
			if strings.HasPrefix(trimmed, "#") {
				break
			}
			// Skip horizontal rules
			if trimmed == "---" || trimmed == "***" || trimmed == "___" {
				continue
			}
			inDescription = true
		}

		// Stop at next heading or empty line after content
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "##") {
			break
		}
		if trimmed == "" && len(descLines) > 0 {
			break
		}

		// Only add non-empty lines
		if trimmed != "" {
			descLines = append(descLines, trimmed)
		}
	}

	desc := strings.Join(descLines, " ")
	// Clean up markdown and HTML formatting
	desc = d.cleanMarkdown(desc)
	desc = d.cleanHTML(desc)
	// Limit length
	if len(desc) > 500 {
		desc = desc[:497] + "..."
	}
	return desc
}

// extractTextFromHTML extracts text content from HTML tags
func (d *ReadmeDetector) extractTextFromHTML(html string) string {
	// Remove HTML tags
	htmlPattern := regexp.MustCompile(`<[^>]+>`)
	text := htmlPattern.ReplaceAllString(html, "")
	return strings.TrimSpace(text)
}

// cleanHTML removes HTML tags and entities
func (d *ReadmeDetector) cleanHTML(text string) string {
	// Remove HTML tags
	htmlPattern := regexp.MustCompile(`<[^>]+>`)
	text = htmlPattern.ReplaceAllString(text, "")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")

	return strings.TrimSpace(text)
}

// parseSections extracts all markdown sections into a map
func (d *ReadmeDetector) parseSections(content string) map[string]string {
	sections := make(map[string]string)

	// Match ## headings (level 2 and 3)
	headerPattern := regexp.MustCompile(`(?m)^#{2,3}\s+(.+)$`)
	matches := headerPattern.FindAllStringSubmatchIndex(content, -1)

	for i, match := range matches {
		if len(match) < 4 {
			continue
		}
		headerName := strings.ToLower(strings.TrimSpace(content[match[2]:match[3]]))
		headerName = d.cleanMarkdown(headerName)

		// Get content until next header or end
		startIdx := match[1]
		endIdx := len(content)
		if i+1 < len(matches) {
			endIdx = matches[i+1][0]
		}

		sectionContent := content[startIdx:endIdx]
		sections[headerName] = sectionContent
	}

	return sections
}

// extractBulletPoints extracts bullet points from a section
func (d *ReadmeDetector) extractBulletPoints(section string) []string {
	var points []string
	lines := strings.Split(section, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Match - or * bullet points
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			point := strings.TrimPrefix(trimmed, "- ")
			point = strings.TrimPrefix(point, "* ")
			point = d.cleanMarkdown(point)
			if point != "" && len(point) < 200 {
				points = append(points, point)
			}
		}
	}

	// Limit to first 10 features
	if len(points) > 10 {
		points = points[:10]
	}
	return points
}

// cleanSection removes markdown formatting and limits length
func (d *ReadmeDetector) cleanSection(section string) string {
	// Remove code blocks for summary (keep for display)
	section = strings.TrimSpace(section)
	if len(section) > 1000 {
		section = section[:997] + "..."
	}
	return section
}

// cleanMarkdown removes common markdown formatting
func (d *ReadmeDetector) cleanMarkdown(text string) string {
	// Remove bold **text**
	boldPattern := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	text = boldPattern.ReplaceAllString(text, "$1")

	// Remove italic *text* or _text_
	italicPattern := regexp.MustCompile(`[*_]([^*_]+)[*_]`)
	text = italicPattern.ReplaceAllString(text, "$1")

	// Remove inline code `text`
	codePattern := regexp.MustCompile("`([^`]+)`")
	text = codePattern.ReplaceAllString(text, "$1")

	// Remove links [text](url) -> text
	linkPattern := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	text = linkPattern.ReplaceAllString(text, "$1")

	return strings.TrimSpace(text)
}
