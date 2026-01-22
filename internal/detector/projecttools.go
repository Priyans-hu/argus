package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// ProjectToolsDetector detects project-specific CLI tools that deserve skills
type ProjectToolsDetector struct {
	rootPath      string
	files         []types.FileInfo
	projectName   string
	techStack     *types.TechStack
	readmeContent *types.ReadmeContent
	cliInfo       *types.CLIInfo
}

// NewProjectToolsDetector creates a new project tools detector
func NewProjectToolsDetector(
	rootPath string,
	files []types.FileInfo,
	projectName string,
	techStack *types.TechStack,
	readmeContent *types.ReadmeContent,
	cliInfo *types.CLIInfo,
) *ProjectToolsDetector {
	return &ProjectToolsDetector{
		rootPath:      rootPath,
		files:         files,
		projectName:   projectName,
		techStack:     techStack,
		readmeContent: readmeContent,
		cliInfo:       cliInfo,
	}
}

// Detect identifies project-specific tools
func (d *ProjectToolsDetector) Detect() []types.ProjectTool {
	var tools []types.ProjectTool

	// 1. Check if the project itself is a CLI tool
	if selfTool := d.detectSelfTool(); selfTool != nil {
		tools = append(tools, *selfTool)
	}

	// 2. Detect custom scripts that are project-specific
	customScripts := d.detectCustomScripts()
	tools = append(tools, customScripts...)

	return tools
}

// detectSelfTool checks if this project is itself a CLI tool worth creating a skill for
func (d *ProjectToolsDetector) detectSelfTool() *types.ProjectTool {
	// Only consider if it's a CLI project
	if d.cliInfo == nil {
		return nil
	}

	// Check for binary in bin/, dist/, or built by Makefile/package.json
	binaryPath := d.findProjectBinary()
	if binaryPath == "" {
		return nil
	}

	// Extract description and usage from README
	description := d.extractToolDescription()
	if description == "" {
		description = "Project-specific CLI tool"
	}

	// Build usage examples from README or commands
	usageExamples := d.extractUsageExamples()

	// Determine when to use
	whenToUse := d.determineWhenToUse()

	return &types.ProjectTool{
		Name:          d.projectName,
		BinaryPath:    binaryPath,
		Description:   description,
		UsageExamples: usageExamples,
		WhenToUse:     whenToUse,
	}
}

// findProjectBinary looks for a binary built by this project
func (d *ProjectToolsDetector) findProjectBinary() string {
	// Common binary locations
	binaryLocations := []string{
		filepath.Join("bin", d.projectName),
		filepath.Join("dist", d.projectName),
		filepath.Join("build", d.projectName),
		d.projectName, // Root directory
	}

	for _, loc := range binaryLocations {
		fullPath := filepath.Join(d.rootPath, loc)
		if _, err := os.Stat(fullPath); err == nil {
			return loc
		}
	}

	// Check if Makefile builds a binary
	if d.checkMakefileBuildsBinary() {
		return "bin/" + d.projectName
	}

	// Check if go.mod suggests it's a Go CLI
	if d.isGoCLI() {
		return "bin/" + d.projectName
	}

	return ""
}

// checkMakefileBuildsBinary checks if Makefile has a build target
func (d *ProjectToolsDetector) checkMakefileBuildsBinary() bool {
	makefilePath := filepath.Join(d.rootPath, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return false
	}

	contentStr := string(content)

	// Look for build commands that output to bin/
	buildPatterns := []string{
		`bin/` + d.projectName,
		`-o\s+bin/`,
		`go\s+build.*-o`,
	}

	for _, pattern := range buildPatterns {
		if matched, _ := regexp.MatchString(pattern, contentStr); matched {
			return true
		}
	}

	return false
}

// isGoCLI checks if this is a Go CLI project
func (d *ProjectToolsDetector) isGoCLI() bool {
	// Check for cmd/ directory with main.go
	for _, f := range d.files {
		if strings.HasPrefix(f.Path, "cmd/"+d.projectName+"/") && f.Name == "main.go" {
			return true
		}
	}

	// Check for main.go in root
	mainPath := filepath.Join(d.rootPath, "main.go")
	if _, err := os.Stat(mainPath); err == nil {
		return true
	}

	return false
}

// extractToolDescription extracts tool description from README
func (d *ProjectToolsDetector) extractToolDescription() string {
	if d.readmeContent == nil {
		return ""
	}

	// Use README description if available
	if d.readmeContent.Description != "" {
		// Limit to first sentence or 200 chars
		desc := d.readmeContent.Description
		if len(desc) > 200 {
			desc = desc[:200]
			// Cut at last period
			if idx := strings.LastIndex(desc, "."); idx > 0 {
				desc = desc[:idx+1]
			}
		}
		return desc
	}

	return ""
}

// extractUsageExamples extracts usage examples from README
func (d *ProjectToolsDetector) extractUsageExamples() []string {
	var examples []string

	if d.readmeContent == nil {
		return examples
	}

	// Extract from QuickStart or Usage sections
	usageText := d.readmeContent.QuickStart
	if usageText == "" {
		usageText = d.readmeContent.Usage
	}

	if usageText == "" {
		return examples
	}

	// Look for code blocks with project name
	codeBlockRegex := regexp.MustCompile("```(?:bash|sh)?\\s*\n([^`]+)```")
	matches := codeBlockRegex.FindAllStringSubmatch(usageText, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			lines := strings.Split(match[1], "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				// Only include lines that use the project tool
				if strings.HasPrefix(line, d.projectName+" ") || strings.HasPrefix(line, "./"+d.projectName) {
					examples = append(examples, line)
					if len(examples) >= 5 {
						return examples
					}
				}
			}
		}
	}

	return examples
}

// determineWhenToUse determines when Claude should use this tool
func (d *ProjectToolsDetector) determineWhenToUse() string {
	// Check README for semantic search, AI, or special capabilities
	if d.readmeContent != nil {
		desc := strings.ToLower(d.readmeContent.Description)
		title := strings.ToLower(d.readmeContent.Title)
		combined := desc + " " + title

		// Semantic search tool
		if strings.Contains(combined, "semantic") && strings.Contains(combined, "search") {
			return "Use for semantic code search when you need to find code by intent or meaning, not just text matching. Invoke before using Grep/Glob for intent-based searches."
		}

		// Code analysis tool
		if strings.Contains(combined, "analysis") || strings.Contains(combined, "analyzer") {
			return "Use when you need to analyze code structure, patterns, or quality. Invoke when understanding codebase architecture or detecting issues."
		}

		// Call graph tool
		if strings.Contains(combined, "call graph") || strings.Contains(combined, "trace") {
			return "Use to understand function relationships and call hierarchies. Invoke when analyzing code flow or impact of changes."
		}
	}

	return "Use when working with this codebase for project-specific operations."
}

// detectCustomScripts detects custom scripts that might need skills
func (d *ProjectToolsDetector) detectCustomScripts() []types.ProjectTool {
	var tools []types.ProjectTool

	// For now, we'll focus on the self-tool detection
	// Custom scripts can be added later if needed

	return tools
}
