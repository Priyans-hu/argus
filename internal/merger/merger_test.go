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

const (
	generated = "# demo\n\nnew body\n"
	block     = AutoStartMarker + "\n# demo\n\nnew body\n" + AutoEndMarker
	oldBlock  = AutoStartMarker + "\n# demo\n\nold body\n" + AutoEndMarker
	custom    = CustomStartMarker + "\n## Notes\nkeep me\n" + CustomEndMarker
	previous  = CustomStartMarker + "\n## Previous Content\n\nThe following content was preserved from your original file:\n\n# demo\nold body\n" + CustomEndMarker
)

func TestMerge(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		want     string
		outcome  Outcome
	}{
		{"missing", "", block, Created},
		{"whitespace only", "  \n\t\n", block, Created},
		{"auto only", oldBlock + "\n", block + "\n", Replaced},
		{
			"text above, between and below",
			"# Team rules\nKEEP-TOP\n\n" + oldBlock + "\n\nKEEP-MID\n\n" + custom + "\n\nKEEP-BOTTOM\n",
			"# Team rules\nKEEP-TOP\n\n" + block + "\n\nKEEP-MID\n\n" + custom + "\n\nKEEP-BOTTOM\n",
			Replaced,
		},
		{"auto moved below custom", custom + "\n\n" + oldBlock + "\n", custom + "\n\n" + block + "\n", Replaced},
		{"v0.4.0 previous content kept", oldBlock + "\n\n" + previous, block + "\n\n" + previous, Replaced},
		{"hand-written", "\n\n# Team rules\n- use the mock\n", block + "\n\n# Team rules\n- use the mock\n", Adopted},
		{"custom only", "# demo\nold body\n\n" + custom, block + "\n\n# demo\nold body\n\n" + custom, Adopted},
		{"BOM only", "\ufeff\n", block, Created},
		{"BOM and text", "\ufeff# Team rules\n", block + "\n\n# Team rules\n", Adopted},
		{"BOM and auto block", "\ufeff" + oldBlock + "\n", "\ufeff" + block + "\n", Replaced},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, outcome := NewMerger(true).Merge([]byte(tt.existing), []byte(generated))
			if string(got) != tt.want {
				t.Errorf("Merge() =\n%q\nwant\n%q", got, tt.want)
			}
			if outcome != tt.outcome {
				t.Errorf("outcome = %v, want %v", outcome, tt.outcome)
			}

			again, outcome := NewMerger(true).Merge(got, []byte(generated))
			if string(again) != string(got) || outcome != Replaced {
				t.Errorf("merging the result again = %q (%v), want it unchanged (Replaced)", again, outcome)
			}
		})
	}
}

func TestMerge_NilExisting(t *testing.T) {
	got, outcome := NewMerger(true).Merge(nil, []byte(generated))
	if string(got) != block || outcome != Created {
		t.Errorf("Merge(nil) = %q (%v), want %q (Created)", got, outcome, block)
	}
}

func TestMerge_Malformed(t *testing.T) {
	tests := []struct {
		name     string
		existing string
	}{
		{"two starts", oldBlock + "\n" + AutoStartMarker + "\nstray\n"},
		{"start only", AutoStartMarker + "\nonly a start\n"},
		{"end only", "text\n" + AutoEndMarker + "\n"},
		{"end before start", AutoEndMarker + "\nbody\n" + AutoStartMarker + "\n"},
		{"two pairs", oldBlock + "\n\n" + oldBlock + "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, outcome := NewMerger(true).Merge([]byte(tt.existing), []byte(generated))
			if string(got) != tt.existing {
				t.Errorf("Merge() changed a malformed file:\n%q", got)
			}
			if outcome != Malformed {
				t.Errorf("outcome = %v, want Malformed", outcome)
			}
		})
	}
}

func TestMerge_DefusesMarkerText(t *testing.T) {
	const (
		text = "- Replaces " + AutoStartMarker + " and " + AutoEndMarker + ", keeps " + CustomStartMarker + " and " + CustomEndMarker + "\n" +
			AutoStartMarker + "\n" + AutoEndMarker + "\n" + CustomStartMarker + "\n" + CustomEndMarker + "\n"
		wrapped = AutoStartMarker + "\n" +
			"- Replaces <!--ARGUS:AUTO --> and <!--/ARGUS:AUTO -->, keeps <!--ARGUS:CUSTOM --> and <!--/ARGUS:CUSTOM -->\n" +
			"<!--ARGUS:AUTO -->\n<!--/ARGUS:AUTO -->\n<!--ARGUS:CUSTOM -->\n<!--/ARGUS:CUSTOM -->\n" +
			AutoEndMarker
	)
	if got := WrapContent(text); got != wrapped {
		t.Errorf("WrapContent() = %q, want %q", got, wrapped)
	}

	tests := []struct {
		name     string
		existing []byte
		want     string
		outcome  Outcome
	}{
		{"nil", nil, wrapped, Created},
		{"unmarked", []byte("# Team rules\nKEEP\n"), wrapped + "\n\n# Team rules\nKEEP\n", Adopted},
		{"marked with custom", []byte(oldBlock + "\n\n" + custom + "\n"), wrapped + "\n\n" + custom + "\n", Replaced},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, outcome := NewMerger(true).Merge(tt.existing, []byte(text))
			if string(got) != tt.want || outcome != tt.outcome {
				t.Errorf("Merge() = %q (%v), want %q (%v)", got, outcome, tt.want, tt.outcome)
			}
			if strings.Count(string(got), AutoStartMarker) != 1 || strings.Count(string(got), AutoEndMarker) != 1 {
				t.Errorf("want exactly one AUTO pair:\n%s", got)
			}

			again, outcome := NewMerger(true).Merge(got, []byte(text))
			if string(again) != string(got) || outcome != Replaced {
				t.Errorf("merging the result again = %q (%v), want it unchanged (Replaced)", again, outcome)
			}
		})
	}
}

func TestMerge_KeepsCRLFOutsideBlock(t *testing.T) {
	existing := "# Team rules\r\nKEEP-TOP\r\n\r\n" + strings.ReplaceAll(oldBlock, "\n", "\r\n") + "\r\n\r\nKEEP-BOTTOM\r\n"
	want := "# Team rules\r\nKEEP-TOP\r\n\r\n" + block + "\r\n\r\nKEEP-BOTTOM\r\n"

	got, outcome := NewMerger(true).Merge([]byte(existing), []byte(generated))
	if string(got) != want || outcome != Replaced {
		t.Errorf("Merge() = %q (%v), want %q (Replaced)", got, outcome, want)
	}
}

func TestMerge_PreserveCustomDisabled(t *testing.T) {
	existing := "# Team rules\n\n" + oldBlock + "\n\n" + custom + "\n"

	got, outcome := NewMerger(false).Merge([]byte(existing), []byte(generated))
	if string(got) != block || outcome != Replaced {
		t.Errorf("Merge() = %q (%v), want %q (Replaced)", got, outcome, block)
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
