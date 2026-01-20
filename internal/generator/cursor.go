package generator

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// CursorGenerator generates .cursorrules files
type CursorGenerator struct{}

// NewCursorGenerator creates a new Cursor generator
func NewCursorGenerator() *CursorGenerator {
	return &CursorGenerator{}
}

// Name returns the generator name
func (g *CursorGenerator) Name() string {
	return "cursor"
}

// OutputFile returns the output filename
func (g *CursorGenerator) OutputFile() string {
	return ".cursorrules"
}

// Generate creates the .cursorrules content
func (g *CursorGenerator) Generate(analysis *types.Analysis) ([]byte, error) {
	var buf bytes.Buffer

	// Header section
	g.writeHeader(&buf, analysis)

	// Tech stack context
	g.writeTechContext(&buf, &analysis.TechStack)

	// Project structure
	g.writeProjectStructure(&buf, &analysis.Structure)

	// Conventions as rules
	g.writeRules(&buf, analysis.Conventions)

	// Code style guidelines
	g.writeCodeStyle(&buf, &analysis.TechStack, analysis.Conventions)

	// Key files reference
	g.writeKeyFilesReference(&buf, analysis.KeyFiles)

	return buf.Bytes(), nil
}

// writeHeader writes the header section
func (g *CursorGenerator) writeHeader(buf *bytes.Buffer, analysis *types.Analysis) {
	buf.WriteString("# Project: " + analysis.ProjectName + "\n\n")
	buf.WriteString("You are an expert developer working on this codebase. Follow these rules and conventions.\n\n")
}

// writeTechContext writes the technology context
func (g *CursorGenerator) writeTechContext(buf *bytes.Buffer, stack *types.TechStack) {
	buf.WriteString("## Technology Stack\n\n")

	// Primary language
	if len(stack.Languages) > 0 {
		// Sort by percentage
		langs := make([]types.Language, len(stack.Languages))
		copy(langs, stack.Languages)
		sort.Slice(langs, func(i, j int) bool {
			return langs[i].Percentage > langs[j].Percentage
		})

		primary := langs[0]
		if primary.Version != "" {
			fmt.Fprintf(buf, "- Primary Language: %s %s\n", primary.Name, primary.Version)
		} else {
			fmt.Fprintf(buf, "- Primary Language: %s\n", primary.Name)
		}

		// Other languages
		if len(langs) > 1 {
			others := make([]string, 0)
			for _, l := range langs[1:] {
				if l.Percentage > 5 { // Only significant languages
					others = append(others, l.Name)
				}
			}
			if len(others) > 0 {
				fmt.Fprintf(buf, "- Other Languages: %s\n", strings.Join(others, ", "))
			}
		}
	}

	// Frameworks
	if len(stack.Frameworks) > 0 {
		frameworks := make([]string, 0)
		for _, fw := range stack.Frameworks {
			if fw.Version != "" {
				frameworks = append(frameworks, fmt.Sprintf("%s %s", fw.Name, fw.Version))
			} else {
				frameworks = append(frameworks, fw.Name)
			}
		}
		fmt.Fprintf(buf, "- Frameworks: %s\n", strings.Join(frameworks, ", "))
	}

	// Databases
	if len(stack.Databases) > 0 {
		fmt.Fprintf(buf, "- Databases: %s\n", strings.Join(stack.Databases, ", "))
	}

	// Tools
	if len(stack.Tools) > 0 {
		fmt.Fprintf(buf, "- Tools: %s\n", strings.Join(stack.Tools, ", "))
	}

	buf.WriteString("\n")
}

// writeProjectStructure writes the project structure overview
func (g *CursorGenerator) writeProjectStructure(buf *bytes.Buffer, structure *types.ProjectStructure) {
	if len(structure.Directories) == 0 {
		return
	}

	buf.WriteString("## Project Structure\n\n")

	// Sort directories
	dirs := make([]types.Directory, len(structure.Directories))
	copy(dirs, structure.Directories)
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Path < dirs[j].Path
	})

	for _, dir := range dirs {
		if dir.Purpose != "" {
			fmt.Fprintf(buf, "- `%s/` - %s\n", dir.Path, dir.Purpose)
		} else {
			fmt.Fprintf(buf, "- `%s/`\n", dir.Path)
		}
	}
	buf.WriteString("\n")
}

// writeRules writes conventions as rules
func (g *CursorGenerator) writeRules(buf *bytes.Buffer, conventions []types.Convention) {
	if len(conventions) == 0 {
		return
	}

	buf.WriteString("## Rules\n\n")

	// Group by category
	byCategory := make(map[string][]types.Convention)
	categoryOrder := []string{}

	for _, conv := range conventions {
		cat := conv.Category
		if cat == "" {
			cat = "general"
		}
		if _, exists := byCategory[cat]; !exists {
			categoryOrder = append(categoryOrder, cat)
		}
		byCategory[cat] = append(byCategory[cat], conv)
	}

	for _, cat := range categoryOrder {
		convs := byCategory[cat]
		fmt.Fprintf(buf, "### %s\n\n", titleCase(cat))
		for _, conv := range convs {
			fmt.Fprintf(buf, "- %s\n", conv.Description)
		}
		buf.WriteString("\n")
	}
}

// writeCodeStyle writes code style guidelines based on tech stack
func (g *CursorGenerator) writeCodeStyle(buf *bytes.Buffer, stack *types.TechStack, conventions []types.Convention) {
	buf.WriteString("## Code Style Guidelines\n\n")

	// Language-specific guidelines
	if len(stack.Languages) > 0 {
		primary := stack.Languages[0].Name

		switch strings.ToLower(primary) {
		case "typescript", "javascript":
			g.writeJSStyleGuide(buf, stack)
		case "go":
			g.writeGoStyleGuide(buf)
		case "python":
			g.writePythonStyleGuide(buf)
		case "java":
			g.writeJavaStyleGuide(buf)
		case "rust":
			g.writeRustStyleGuide(buf)
		case "c#":
			g.writeCSharpStyleGuide(buf)
		case "ruby":
			g.writeRubyStyleGuide(buf)
		default:
			g.writeGenericStyleGuide(buf, primary)
		}
	}
}

func (g *CursorGenerator) writeJSStyleGuide(buf *bytes.Buffer, stack *types.TechStack) {
	buf.WriteString("When writing JavaScript/TypeScript code:\n\n")
	buf.WriteString("- Use consistent naming: camelCase for variables/functions, PascalCase for classes/components\n")
	buf.WriteString("- Prefer const over let, avoid var\n")
	buf.WriteString("- Use async/await over raw promises when possible\n")
	buf.WriteString("- Handle errors properly in try/catch blocks\n")

	// Check for React
	for _, fw := range stack.Frameworks {
		if strings.Contains(strings.ToLower(fw.Name), "react") {
			buf.WriteString("- React: Use functional components with hooks\n")
			buf.WriteString("- React: Extract reusable logic into custom hooks\n")
			break
		}
	}

	buf.WriteString("\n")
}

func (g *CursorGenerator) writeGoStyleGuide(buf *bytes.Buffer) {
	buf.WriteString("When writing Go code:\n\n")
	buf.WriteString("- Follow effective Go guidelines\n")
	buf.WriteString("- Use gofmt/goimports for formatting\n")
	buf.WriteString("- Handle all errors explicitly (if err != nil)\n")
	buf.WriteString("- Use meaningful variable names, short names for short scopes\n")
	buf.WriteString("- Document exported functions with comments starting with function name\n")
	buf.WriteString("- Prefer composition over inheritance\n")
	buf.WriteString("\n")
}

func (g *CursorGenerator) writePythonStyleGuide(buf *bytes.Buffer) {
	buf.WriteString("When writing Python code:\n\n")
	buf.WriteString("- Follow PEP 8 style guide\n")
	buf.WriteString("- Use snake_case for variables and functions\n")
	buf.WriteString("- Use PascalCase for classes\n")
	buf.WriteString("- Add type hints to function signatures\n")
	buf.WriteString("- Write docstrings for functions and classes\n")
	buf.WriteString("- Use context managers (with statements) for resource management\n")
	buf.WriteString("\n")
}

func (g *CursorGenerator) writeJavaStyleGuide(buf *bytes.Buffer) {
	buf.WriteString("When writing Java code:\n\n")
	buf.WriteString("- Follow Java naming conventions\n")
	buf.WriteString("- Use camelCase for methods and variables\n")
	buf.WriteString("- Use PascalCase for classes\n")
	buf.WriteString("- Use UPPER_SNAKE_CASE for constants\n")
	buf.WriteString("- Add Javadoc comments to public methods\n")
	buf.WriteString("- Prefer interfaces over concrete types\n")
	buf.WriteString("\n")
}

func (g *CursorGenerator) writeRustStyleGuide(buf *bytes.Buffer) {
	buf.WriteString("When writing Rust code:\n\n")
	buf.WriteString("- Follow Rust API guidelines\n")
	buf.WriteString("- Use snake_case for functions and variables\n")
	buf.WriteString("- Use PascalCase for types and traits\n")
	buf.WriteString("- Handle Result/Option types properly, avoid unwrap in production code\n")
	buf.WriteString("- Prefer borrowing over cloning when possible\n")
	buf.WriteString("- Write doc comments with /// for public items\n")
	buf.WriteString("\n")
}

func (g *CursorGenerator) writeCSharpStyleGuide(buf *bytes.Buffer) {
	buf.WriteString("When writing C# code:\n\n")
	buf.WriteString("- Follow .NET naming conventions\n")
	buf.WriteString("- Use PascalCase for public members\n")
	buf.WriteString("- Use camelCase for private fields (with _ prefix optional)\n")
	buf.WriteString("- Add XML documentation comments to public APIs\n")
	buf.WriteString("- Use async/await for asynchronous operations\n")
	buf.WriteString("- Prefer LINQ for collection operations\n")
	buf.WriteString("\n")
}

func (g *CursorGenerator) writeRubyStyleGuide(buf *bytes.Buffer) {
	buf.WriteString("When writing Ruby code:\n\n")
	buf.WriteString("- Follow Ruby style guide\n")
	buf.WriteString("- Use snake_case for methods and variables\n")
	buf.WriteString("- Use PascalCase for classes and modules\n")
	buf.WriteString("- Use SCREAMING_SNAKE_CASE for constants\n")
	buf.WriteString("- Prefer blocks over explicit procs\n")
	buf.WriteString("- Use meaningful method names that read like English\n")
	buf.WriteString("\n")
}

func (g *CursorGenerator) writeGenericStyleGuide(buf *bytes.Buffer, language string) {
	fmt.Fprintf(buf, "When writing %s code:\n\n", language)
	buf.WriteString("- Follow language-specific conventions\n")
	buf.WriteString("- Use consistent naming throughout the codebase\n")
	buf.WriteString("- Write clear, descriptive variable and function names\n")
	buf.WriteString("- Add comments for complex logic\n")
	buf.WriteString("- Keep functions focused and single-purpose\n")
	buf.WriteString("\n")
}

// writeKeyFilesReference writes key files for reference
func (g *CursorGenerator) writeKeyFilesReference(buf *bytes.Buffer, keyFiles []types.KeyFile) {
	if len(keyFiles) == 0 {
		return
	}

	buf.WriteString("## Key Files Reference\n\n")

	for _, kf := range keyFiles {
		if kf.Description != "" {
			fmt.Fprintf(buf, "- `%s` - %s\n", kf.Path, kf.Description)
		} else if kf.Purpose != "" {
			fmt.Fprintf(buf, "- `%s` - %s\n", kf.Path, kf.Purpose)
		} else {
			fmt.Fprintf(buf, "- `%s`\n", kf.Path)
		}
	}
	buf.WriteString("\n")
}
