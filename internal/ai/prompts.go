package ai

import (
	"fmt"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// buildProjectSummaryPrompt creates a prompt for generating a project summary
func buildProjectSummaryPrompt(analysis *types.Analysis) string {
	var sb strings.Builder

	sb.WriteString("You are a technical writer. Based on the following codebase analysis, write a concise 2-3 sentence project summary.\n\n")
	sb.WriteString(fmt.Sprintf("Project: %s\n", analysis.ProjectName))

	if len(analysis.TechStack.Languages) > 0 {
		var langs []string
		for _, l := range analysis.TechStack.Languages {
			langs = append(langs, l.Name)
		}
		sb.WriteString(fmt.Sprintf("Languages: %s\n", strings.Join(langs, ", ")))
	}

	if len(analysis.TechStack.Frameworks) > 0 {
		var fws []string
		for _, f := range analysis.TechStack.Frameworks {
			fws = append(fws, f.Name)
		}
		sb.WriteString(fmt.Sprintf("Frameworks: %s\n", strings.Join(fws, ", ")))
	}

	if analysis.ReadmeContent != nil && analysis.ReadmeContent.Description != "" {
		sb.WriteString(fmt.Sprintf("README description: %s\n", analysis.ReadmeContent.Description))
	}

	if len(analysis.Structure.Directories) > 0 {
		var dirs []string
		for _, d := range analysis.Structure.Directories {
			if d.Purpose != "" {
				dirs = append(dirs, fmt.Sprintf("%s (%s)", d.Path, d.Purpose))
			}
		}
		if len(dirs) > 0 {
			sb.WriteString(fmt.Sprintf("Key directories: %s\n", strings.Join(dirs, ", ")))
		}
	}

	sb.WriteString("\nRespond with ONLY the summary text, no headers or formatting.\n")
	return sb.String()
}

// buildConventionsPrompt creates a prompt for enriching detected conventions
func buildConventionsPrompt(analysis *types.Analysis) string {
	var sb strings.Builder

	sb.WriteString("You are a senior developer. Based on this codebase analysis, suggest 3-5 additional coding conventions that would benefit this project.\n\n")
	sb.WriteString(fmt.Sprintf("Project: %s\n", analysis.ProjectName))

	if len(analysis.TechStack.Languages) > 0 {
		sb.WriteString(fmt.Sprintf("Primary language: %s\n", analysis.TechStack.Languages[0].Name))
	}

	if len(analysis.Conventions) > 0 {
		sb.WriteString("Already detected conventions:\n")
		limit := min(len(analysis.Conventions), 10)
		for _, c := range analysis.Conventions[:limit] {
			sb.WriteString(fmt.Sprintf("- [%s] %s\n", c.Category, c.Description))
		}
	}

	sb.WriteString("\nRespond with a JSON array of objects, each with \"title\" and \"description\" fields. Example:\n")
	sb.WriteString(`[{"title":"Error wrapping","description":"Always wrap errors with context using fmt.Errorf"}]`)
	sb.WriteString("\nRespond with ONLY the JSON array.\n")
	return sb.String()
}

// buildArchitecturePrompt creates a prompt for architecture insights
func buildArchitecturePrompt(analysis *types.Analysis) string {
	var sb strings.Builder

	sb.WriteString("You are a software architect. Based on this codebase analysis, provide 2-4 architectural insights or recommendations.\n\n")
	sb.WriteString(fmt.Sprintf("Project: %s\n", analysis.ProjectName))

	if analysis.ArchitectureInfo != nil {
		if analysis.ArchitectureInfo.Style != "" {
			sb.WriteString(fmt.Sprintf("Architecture style: %s\n", analysis.ArchitectureInfo.Style))
		}
		if len(analysis.ArchitectureInfo.Layers) > 0 {
			sb.WriteString("Layers:\n")
			for _, l := range analysis.ArchitectureInfo.Layers {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", l.Name, l.Purpose))
			}
		}
	}

	if len(analysis.Structure.Directories) > 0 {
		sb.WriteString("Directory structure:\n")
		for _, d := range analysis.Structure.Directories {
			if d.Purpose != "" {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", d.Path, d.Purpose))
			}
		}
	}

	sb.WriteString("\nRespond with a JSON array of objects, each with \"title\" and \"description\" fields.\n")
	sb.WriteString("Respond with ONLY the JSON array.\n")
	return sb.String()
}

// buildBestPracticesPrompt creates a prompt for best practices
func buildBestPracticesPrompt(analysis *types.Analysis) string {
	var sb strings.Builder

	sb.WriteString("You are a senior developer. Based on this project's tech stack, suggest 3-5 best practices specific to this project.\n\n")
	sb.WriteString(fmt.Sprintf("Project: %s\n", analysis.ProjectName))

	if len(analysis.TechStack.Languages) > 0 {
		var langs []string
		for _, l := range analysis.TechStack.Languages {
			langs = append(langs, l.Name)
		}
		sb.WriteString(fmt.Sprintf("Languages: %s\n", strings.Join(langs, ", ")))
	}

	if len(analysis.TechStack.Frameworks) > 0 {
		var fws []string
		for _, f := range analysis.TechStack.Frameworks {
			fws = append(fws, f.Name)
		}
		sb.WriteString(fmt.Sprintf("Frameworks: %s\n", strings.Join(fws, ", ")))
	}

	if len(analysis.TechStack.Databases) > 0 {
		sb.WriteString(fmt.Sprintf("Databases: %s\n", strings.Join(analysis.TechStack.Databases, ", ")))
	}

	sb.WriteString("\nRespond with a JSON array of objects, each with \"title\" and \"description\" fields.\n")
	sb.WriteString("Respond with ONLY the JSON array.\n")
	return sb.String()
}

// buildPatternsPrompt creates a prompt for pattern analysis
func buildPatternsPrompt(analysis *types.Analysis) string {
	var sb strings.Builder

	sb.WriteString("You are a code reviewer. Based on the detected code patterns, suggest 2-4 insights about how this project uses patterns.\n\n")
	sb.WriteString(fmt.Sprintf("Project: %s\n", analysis.ProjectName))

	if analysis.CodePatterns != nil {
		writePatternSummary := func(name string, patterns []types.PatternInfo) {
			if len(patterns) == 0 {
				return
			}
			var names []string
			for _, p := range patterns {
				names = append(names, p.Name)
			}
			sb.WriteString(fmt.Sprintf("%s: %s\n", name, strings.Join(names, ", ")))
		}

		writePatternSummary("Testing", analysis.CodePatterns.Testing)
		writePatternSummary("Data Fetching", analysis.CodePatterns.DataFetching)
		writePatternSummary("API Patterns", analysis.CodePatterns.APIPatterns)
		writePatternSummary("Database", analysis.CodePatterns.DatabaseORM)
		writePatternSummary("Go Patterns", analysis.CodePatterns.GoPatterns)
		writePatternSummary("Authentication", analysis.CodePatterns.Authentication)
	}

	sb.WriteString("\nRespond with a JSON array of objects, each with \"title\" and \"description\" fields.\n")
	sb.WriteString("Respond with ONLY the JSON array.\n")
	return sb.String()
}
