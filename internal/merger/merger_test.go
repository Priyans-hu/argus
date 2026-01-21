package merger

import (
	"strings"
	"testing"
)

func TestNewMerger(t *testing.T) {
	tests := []struct {
		name           string
		preserveCustom bool
	}{
		{"with preserve custom enabled", true},
		{"with preserve custom disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMerger(tt.preserveCustom)
			if m == nil {
				t.Fatal("NewMerger returned nil")
			}
			if m.preserveCustom != tt.preserveCustom {
				t.Errorf("preserveCustom = %v, want %v", m.preserveCustom, tt.preserveCustom)
			}
		})
	}
}

func TestMerge_PreserveCustomDisabled(t *testing.T) {
	m := NewMerger(false)
	existing := []byte("old content")
	new := []byte("new content")

	result := m.Merge(existing, new)

	if string(result) != "new content" {
		t.Errorf("Merge() = %q, want %q", string(result), "new content")
	}
}

func TestMerge_NoMarkersInExisting(t *testing.T) {
	m := NewMerger(true)
	existing := []byte("existing content without markers")
	new := []byte("new generated content")

	result := m.Merge(existing, new)
	resultStr := string(result)

	// Should wrap new content with auto markers
	if !strings.Contains(resultStr, AutoStartMarker) {
		t.Error("Result should contain AutoStartMarker")
	}
	if !strings.Contains(resultStr, AutoEndMarker) {
		t.Error("Result should contain AutoEndMarker")
	}
	if !strings.Contains(resultStr, "new generated content") {
		t.Error("Result should contain new content")
	}

	// Should preserve existing content in custom section
	if !strings.Contains(resultStr, CustomStartMarker) {
		t.Error("Result should contain CustomStartMarker for preserved content")
	}
	if !strings.Contains(resultStr, "existing content without markers") {
		t.Error("Result should preserve existing content")
	}
	if !strings.Contains(resultStr, "Previous Content") {
		t.Error("Result should have 'Previous Content' heading for preserved content")
	}
}

func TestMerge_EmptyExistingNoMarkers(t *testing.T) {
	m := NewMerger(true)
	existing := []byte("   \n\t  ") // whitespace only
	new := []byte("new generated content")

	result := m.Merge(existing, new)
	resultStr := string(result)

	// Should wrap new content with markers
	if !strings.Contains(resultStr, AutoStartMarker) {
		t.Error("Result should contain AutoStartMarker")
	}
	if !strings.Contains(resultStr, "new generated content") {
		t.Error("Result should contain new content")
	}

	// Should NOT add custom section for empty/whitespace content
	if strings.Contains(resultStr, CustomStartMarker) {
		t.Error("Result should not contain CustomStartMarker for empty content")
	}
}

func TestMerge_WithCustomSections(t *testing.T) {
	m := NewMerger(true)

	existing := AutoStartMarker + "\nold auto content\n" + AutoEndMarker + "\n\n" +
		CustomStartMarker + "\n## My Custom Notes\nImportant info here\n" + CustomEndMarker

	new := []byte("new auto content")

	result := m.Merge([]byte(existing), new)
	resultStr := string(result)

	// Should contain new auto content
	if !strings.Contains(resultStr, "new auto content") {
		t.Error("Result should contain new auto content")
	}

	// Should preserve custom section
	if !strings.Contains(resultStr, "My Custom Notes") {
		t.Error("Result should preserve custom section content")
	}
	if !strings.Contains(resultStr, "Important info here") {
		t.Error("Result should preserve custom section details")
	}
}

func TestMerge_MultipleCustomSections(t *testing.T) {
	m := NewMerger(true)

	existing := AutoStartMarker + "\nauto\n" + AutoEndMarker + "\n\n" +
		CustomStartMarker + "\n## Section 1\n" + CustomEndMarker + "\n\n" +
		CustomStartMarker + "\n## Section 2\n" + CustomEndMarker

	new := []byte("new content")

	result := m.Merge([]byte(existing), new)
	resultStr := string(result)

	// Should preserve both custom sections
	if strings.Count(resultStr, CustomStartMarker) != 2 {
		t.Errorf("Expected 2 custom start markers, got %d", strings.Count(resultStr, CustomStartMarker))
	}
	if !strings.Contains(resultStr, "Section 1") {
		t.Error("Result should preserve Section 1")
	}
	if !strings.Contains(resultStr, "Section 2") {
		t.Error("Result should preserve Section 2")
	}
}

func TestMergeWithSections_PreserveCustomDisabled(t *testing.T) {
	m := NewMerger(false)
	existing := []byte("old content")
	new := []byte("new content")

	result := m.MergeWithSections(existing, new)

	if string(result) != "new content" {
		t.Errorf("MergeWithSections() = %q, want %q", string(result), "new content")
	}
}

func TestMergeWithSections_WithCustomSections(t *testing.T) {
	m := NewMerger(true)

	existing := AutoStartMarker + "\nold\n" + AutoEndMarker + "\n\n" +
		CustomStartMarker + "\n## Custom Notes\nMy notes\n" + CustomEndMarker

	result := m.MergeWithSections([]byte(existing), []byte("new content"))
	resultStr := string(result)

	if !strings.Contains(resultStr, "new content") {
		t.Error("Result should contain new content")
	}
	if !strings.Contains(resultStr, "Custom Notes") {
		t.Error("Result should preserve custom section")
	}
}

func TestMergeWithSections_NoCustomSections(t *testing.T) {
	m := NewMerger(true)

	existing := AutoStartMarker + "\nold auto only\n" + AutoEndMarker

	result := m.MergeWithSections([]byte(existing), []byte("new content"))
	resultStr := string(result)

	if !strings.Contains(resultStr, "new content") {
		t.Error("Result should contain new content")
	}
	if !strings.Contains(resultStr, AutoStartMarker) {
		t.Error("Result should have auto markers")
	}
}

func TestAddCustomSectionPlaceholder(t *testing.T) {
	content := "# My Document"
	result := AddCustomSectionPlaceholder(content)

	if !strings.Contains(result, CustomStartMarker) {
		t.Error("Result should contain CustomStartMarker")
	}
	if !strings.Contains(result, CustomEndMarker) {
		t.Error("Result should contain CustomEndMarker")
	}
	if !strings.Contains(result, "Custom Notes") {
		t.Error("Result should contain placeholder heading")
	}
	if !strings.Contains(result, "Add your custom documentation here") {
		t.Error("Result should contain placeholder text")
	}
}

func TestWrapContent(t *testing.T) {
	content := "my content"
	result := WrapContent(content)

	if !strings.HasPrefix(result, AutoStartMarker) {
		t.Error("Result should start with AutoStartMarker")
	}
	if !strings.HasSuffix(result, AutoEndMarker) {
		t.Error("Result should end with AutoEndMarker")
	}
	if !strings.Contains(result, "my content") {
		t.Error("Result should contain original content")
	}
}

func TestHasMarkers(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"no markers", "plain content", false},
		{"auto start marker", AutoStartMarker + " content", true},
		{"auto end marker", "content " + AutoEndMarker, true},
		{"custom start marker", CustomStartMarker + " content", true},
		{"custom end marker", "content " + CustomEndMarker, true},
		{"full auto section", AutoStartMarker + "\ncontent\n" + AutoEndMarker, true},
		{"full custom section", CustomStartMarker + "\ncontent\n" + CustomEndMarker, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasMarkers(tt.content); got != tt.expected {
				t.Errorf("hasMarkers() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractCustomSections(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedCount int
	}{
		{"no custom sections", "plain content", 0},
		{"one custom section", CustomStartMarker + "\ncontent\n" + CustomEndMarker, 1},
		{
			"two custom sections",
			CustomStartMarker + "\nfirst\n" + CustomEndMarker + "\n\n" +
				CustomStartMarker + "\nsecond\n" + CustomEndMarker,
			2,
		},
		{"mixed content", AutoStartMarker + "\nauto\n" + AutoEndMarker + "\n" +
			CustomStartMarker + "\ncustom\n" + CustomEndMarker, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := extractCustomSections(tt.content)
			if len(sections) != tt.expectedCount {
				t.Errorf("extractCustomSections() returned %d sections, want %d", len(sections), tt.expectedCount)
			}
		})
	}
}

func TestParseSections(t *testing.T) {
	content := AutoStartMarker + "\nauto content\n" + AutoEndMarker + "\n\n" +
		CustomStartMarker + "\n## Custom Heading\ncustom content\n" + CustomEndMarker

	sections := parseSections(content)

	autoCount := 0
	customCount := 0
	for _, s := range sections {
		if s.Type == "auto" {
			autoCount++
		}
		if s.Type == "custom" {
			customCount++
		}
	}

	if autoCount != 1 {
		t.Errorf("Expected 1 auto section, got %d", autoCount)
	}
	if customCount != 1 {
		t.Errorf("Expected 1 custom section, got %d", customCount)
	}
}

func TestExtractSectionName(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"with heading", "## My Section\nsome content", "My Section"},
		{"with heading multiline", "\n## Spaced Heading\nmore", "Spaced Heading"},
		{"no heading", "just content\nno heading here", ""},
		{"h1 heading (not h2)", "# H1 Heading\ncontent", ""},
		{"h3 heading (not h2)", "### H3 Heading\ncontent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractSectionName(tt.content); got != tt.expected {
				t.Errorf("extractSectionName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestHasCustomContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"no custom sections", "plain content", false},
		{
			"placeholder only",
			CustomStartMarker + "\n## Custom Notes\nAdd your custom documentation here.\n" + CustomEndMarker,
			false,
		},
		{
			"real custom content",
			CustomStartMarker + "\n## My Notes\nActual user content here\n" + CustomEndMarker,
			true,
		},
		{"empty custom section", CustomStartMarker + "\n\n" + CustomEndMarker, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasCustomContent(tt.content); got != tt.expected {
				t.Errorf("HasCustomContent() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStripMarkers(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"no markers", "plain content", "plain content"},
		{
			"auto markers",
			AutoStartMarker + "\ncontent\n" + AutoEndMarker,
			"content",
		},
		{
			"custom markers",
			CustomStartMarker + "\ncontent\n" + CustomEndMarker,
			"content",
		},
		{
			"all markers",
			AutoStartMarker + "\nauto\n" + AutoEndMarker + "\n" +
				CustomStartMarker + "\ncustom\n" + CustomEndMarker,
			"auto\n\n\ncustom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripMarkers(tt.content); got != tt.expected {
				t.Errorf("StripMarkers() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestWrapWithMarkers_TrimsTrailingWhitespace(t *testing.T) {
	content := "content with trailing whitespace   \n\n\t"
	result := wrapWithMarkers(content)

	// Should not have trailing whitespace before end marker
	expected := AutoStartMarker + "\ncontent with trailing whitespace\n" + AutoEndMarker
	if result != expected {
		t.Errorf("wrapWithMarkers() = %q, want %q", result, expected)
	}
}

func TestFormatCustomSection(t *testing.T) {
	section := Section{
		Type:    "custom",
		Content: "\n## My Section\nContent here\n",
		Name:    "My Section",
	}

	result := formatCustomSection(section)

	if !strings.HasPrefix(result, CustomStartMarker) {
		t.Error("Result should start with CustomStartMarker")
	}
	if !strings.HasSuffix(result, CustomEndMarker) {
		t.Error("Result should end with CustomEndMarker")
	}
	if !strings.Contains(result, "My Section") {
		t.Error("Result should contain section name")
	}
}
