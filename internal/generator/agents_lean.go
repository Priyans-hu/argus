package generator

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/Priyans-hu/argus/internal/detector"
	"github.com/Priyans-hu/argus/pkg/types"
)

// leanCommandGroups lists the command groups shown in the lean output, in order, with per-group caps.
var leanCommandGroups = []struct {
	name  string
	limit int
}{{"Setup", 2}, {"Build", 3}, {"Test", 3}, {"Lint", 3}, {"Format", 3}, {"Run", 3}, {"Generate", 3}, {"Database", 3}}

// leanConventionCategories are the non-custom convention categories shown in the lean output.
var leanConventionCategories = []string{"typescript", "imports", "testing", "logging", "structure"}

// commitStyleNames maps detected commit styles to display names.
var commitStyleNames = map[string]string{
	"conventional": "Conventional Commits",
	"angular":      "Angular style",
	"gitmoji":      "Gitmoji",
	"jira":         "Jira ticket prefix",
}

// renderLean renders the lean AGENTS.md body: a short summary with every section capped.
func renderLean(a *types.Analysis) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# %s\n\n", oneLine(a.ProjectName))
	if a.ReadmeContent != nil {
		if d := leanDescription(a.ReadmeContent.Description); d != "" {
			fmt.Fprintf(&buf, "%s\n\n", d)
		}
	}
	writeLeanSection(&buf, "Commands", leanCommands(a.Commands))
	writeLeanSection(&buf, "Stack", leanStack(a))
	writeLeanSection(&buf, "Conventions", leanConventions(a.Conventions))
	writeLeanSection(&buf, "Git", leanGit(a.GitConventions))
	writeLeanSection(&buf, "API", leanAPI(a.Endpoints))
	g := &ClaudeGenerator{}
	g.writeUsageInsights(&buf, a.UsageInsights)
	g.writeAIEnrichment(&buf, a.AIEnrichment)
	return append(bytes.TrimRight(buf.Bytes(), "\n"), '\n')
}

// writeLeanSection writes a "## title" section with the given lines, or nothing when there are none.
func writeLeanSection(buf *bytes.Buffer, title string, lines []string) {
	if len(lines) > 0 {
		fmt.Fprintf(buf, "## %s\n\n%s\n\n", title, strings.Join(lines, "\n"))
	}
}

// oneLine collapses all whitespace, including newlines, to single spaces.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// bullet appends "- label: value" to lines when value is non-empty.
func bullet(lines []string, label, value string) []string {
	if value == "" {
		return lines
	}
	return append(lines, "- "+label+": "+value)
}

// joinList joins the distinct non-empty items, each collapsed to one line and
// rendered with format, showing at most limit of them (0 means all) plus "and N more".
func joinList(items []string, limit int, format string) string {
	var out []string
	seen := map[string]bool{}
	for _, s := range items {
		if s = oneLine(s); s != "" && !seen[s] {
			seen[s] = true
			out = append(out, fmt.Sprintf(format, s))
		}
	}
	if limit > 0 && len(out) > limit {
		out = append(out[:limit], fmt.Sprintf("and %d more", len(out)-limit))
	}
	return strings.Join(out, ", ")
}

// leanDescription collapses s to one line and cuts it to at most 400 bytes at the
// last sentence end, else at the last space, else at the last rune boundary.
func leanDescription(s string) string {
	const limit = 400
	s = oneLine(s)
	if len(s) <= limit {
		return s
	}
	t := s[:limit+1] // the byte after the limit decides whether a sentence ends at it
	if i := max(strings.LastIndex(t, ". "), strings.LastIndex(t, "! "), strings.LastIndex(t, "? ")); i >= 0 {
		return s[:i+1]
	}
	if i := strings.LastIndexByte(t, ' '); i > 0 {
		return s[:i]
	}
	i := limit
	for i > 0 && !utf8.RuneStart(s[i]) {
		i--
	}
	return s[:i]
}

// leanCommands renders the de-duplicated commands as one bash block, grouped and capped.
func leanCommands(cmds []types.Command) []string {
	seen := map[string]bool{}
	var unique []types.Command
	for _, c := range cmds {
		if name := oneLine(c.Name); name != "" && !seen[name] {
			seen[name] = true
			unique = append(unique, c)
		}
	}
	groups := detector.GroupCommandsByCategory(unique)
	var lines []string
	total := 0
	for _, g := range leanCommandGroups {
		shown := groups[g.name][:min(len(groups[g.name]), g.limit, 24-total)]
		if len(shown) == 0 {
			continue
		}
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "# "+g.name)
		for _, c := range shown {
			lines = append(lines, oneLine(c.Name))
		}
		total += len(shown)
	}
	if len(lines) == 0 {
		return nil
	}
	return slices.Concat([]string{"```bash"}, lines, []string{"```"})
}

// leanStack renders the languages, frameworks, databases, entry point and monorepo lines.
func leanStack(a *types.Analysis) []string {
	langs := slices.Clone(a.TechStack.Languages)
	slices.SortFunc(langs, compareLanguages)
	var top, frameworks []string
	for _, l := range langs {
		if name := oneLine(l.Name); name != "" && l.Percentage >= 5 && len(top) < 3 {
			top = append(top, fmt.Sprintf("%s %.0f%%", name, l.Percentage))
		}
	}
	for _, f := range a.TechStack.Frameworks {
		frameworks = append(frameworks, f.Name)
	}
	lines := bullet(nil, "Languages", strings.Join(top, ", "))
	lines = bullet(lines, "Frameworks", joinList(frameworks, 0, "%s"))
	lines = bullet(lines, "Databases", joinList(a.TechStack.Databases, 0, "%s"))
	if a.ArchitectureInfo != nil {
		lines = bullet(lines, "Entry", oneLine(a.ArchitectureInfo.EntryPoint))
	}
	if m := a.MonorepoInfo; m != nil && m.IsMonorepo {
		ws := joinList(m.WorkspacePaths, 12, "%s")
		if ws != "" {
			ws = "workspaces: " + ws
		}
		lines = bullet(lines, "Monorepo", joinList([]string{m.Tool, m.PackageManager, ws}, 0, "%s"))
	}
	return lines
}

// leanConventions renders custom conventions first, then a few from the focus categories.
func leanConventions(convs []types.Convention) []string {
	var custom, other []string
	for _, c := range convs {
		d := oneLine(c.Description)
		if d == "" {
			continue
		}
		switch {
		case c.Category == "custom" && len(custom) < 20:
			custom = append(custom, "- "+d)
		case slices.Contains(leanConventionCategories, c.Category) && len(other) < 8:
			other = append(other, "- "+d)
		}
	}
	return append(custom, other...)
}

// leanGit renders the repository and the commit style. Commit types and
// scopes and branch prefixes are ranked by frequency and shift with new
// commits, so they are left out to keep the file stable.
func leanGit(git *types.GitConventions) []string {
	if git == nil {
		return nil
	}
	var lines []string
	if r := git.Repository; r != nil && oneLine(r.Owner) != "" && oneLine(r.Name) != "" {
		lines = bullet(lines, "Repository", oneLine(r.Owner)+"/"+oneLine(r.Name))
	}
	if cc := git.CommitConvention; cc != nil {
		style := oneLine(cc.Style)
		if name, ok := commitStyleNames[style]; ok {
			style = name
		}
		lines = bullet(lines, "Commit style", style)
		lines = bullet(lines, "Commit format", joinList([]string{cc.Format}, 0, "`%s`"))
	}
	return lines
}

// leanAPI renders the endpoint count, the defining files and the top-level resources.
func leanAPI(eps []types.Endpoint) []string {
	if len(eps) == 0 {
		return nil
	}
	var files, resources []string
	for _, e := range eps {
		files = append(files, oneLine(e.File))
		if seg, _, _ := strings.Cut(strings.Trim(oneLine(e.Path), "/"), "/"); seg != "" {
			resources = append(resources, "/"+seg)
		}
	}
	slices.Sort(files)
	slices.Sort(resources)
	lines := []string{fmt.Sprintf("- %d endpoints", len(eps))}
	lines = bullet(lines, "Defined in", joinList(files, 5, "%s"))
	return bullet(lines, "Resources", joinList(resources, 10, "%s"))
}
