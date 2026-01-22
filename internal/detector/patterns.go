package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// PatternDetector detects coding patterns, conventions, and practices
type PatternDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewPatternDetector creates a new pattern detector
func NewPatternDetector(rootPath string, files []types.FileInfo) *PatternDetector {
	return &PatternDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// Detect analyzes the codebase for patterns and conventions
func (d *PatternDetector) Detect() ([]types.Convention, error) {
	var conventions []types.Convention

	// Note: Git branch naming is now handled by GitDetector for better integration
	// with commit conventions output

	// Detect comment/documentation patterns
	conventions = append(conventions, d.detectCommentPatterns()...)

	// Detect logging patterns
	conventions = append(conventions, d.detectLoggingPatterns()...)

	// Detect error handling patterns
	conventions = append(conventions, d.detectErrorHandling()...)

	// Detect architectural patterns
	conventions = append(conventions, d.detectArchitecturalPatterns()...)

	return conventions, nil
}

// detectCommentPatterns analyzes documentation and comment styles
func (d *PatternDetector) detectCommentPatterns() []types.Convention {
	var conventions []types.Convention

	// Patterns to detect
	jsdocCount := 0
	javadocCount := 0
	pythonDocCount := 0
	goDocCount := 0
	xmlDocCount := 0
	todoCount := 0
	fixmeCount := 0

	// Regex patterns
	jsdocRegex := regexp.MustCompile(`/\*\*[\s\S]*?@(param|returns|type|example)`)
	javadocRegex := regexp.MustCompile(`/\*\*[\s\S]*?@(param|return|throws|see)`)
	pythonDocRegex := regexp.MustCompile(`"""[\s\S]*?(Args|Returns|Raises|Example):`)
	goDocRegex := regexp.MustCompile(`(?m)^// [A-Z][a-z]+ (is|returns|creates|handles)`)
	xmlDocRegex := regexp.MustCompile(`/// <(summary|param|returns)>`)
	todoRegex := regexp.MustCompile(`(?i)(TODO|FIXME|HACK|XXX)[\s:]+`)

	sampledFiles := 0
	maxSamples := 50

	for _, f := range d.files {
		if f.IsDir || sampledFiles >= maxSamples {
			continue
		}

		if !isDocumentableFile(f.Extension) {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 { // Skip large files
			continue
		}

		contentStr := string(content)

		// Check patterns based on file type
		switch f.Extension {
		case ".js", ".jsx", ".ts", ".tsx":
			if jsdocRegex.MatchString(contentStr) {
				jsdocCount++
			}
		case ".java", ".kt":
			if javadocRegex.MatchString(contentStr) {
				javadocCount++
			}
		case ".py":
			if pythonDocRegex.MatchString(contentStr) {
				pythonDocCount++
			}
		case ".go":
			if goDocRegex.MatchString(contentStr) {
				goDocCount++
			}
		case ".cs":
			if xmlDocRegex.MatchString(contentStr) {
				xmlDocCount++
			}
		}

		// Check for TODOs
		if todoRegex.MatchString(contentStr) {
			matches := todoRegex.FindAllString(contentStr, -1)
			todoCount += len(matches)
			for _, match := range matches {
				if strings.Contains(strings.ToUpper(match), "FIXME") {
					fixmeCount++
				}
			}
		}

		sampledFiles++
	}

	// Report detected patterns
	if jsdocCount >= 5 {
		conventions = append(conventions, types.Convention{
			Category:    "documentation",
			Description: "JSDoc comments for function documentation",
			Example:     "/** @param {string} name - User name */",
		})
	}

	if javadocCount >= 5 {
		conventions = append(conventions, types.Convention{
			Category:    "documentation",
			Description: "Javadoc comments for class and method documentation",
			Example:     "/** @param name the user name */",
		})
	}

	if pythonDocCount >= 5 {
		conventions = append(conventions, types.Convention{
			Category:    "documentation",
			Description: "Google-style Python docstrings",
			Example:     "\"\"\"Args:\\n    name: User name\\n\"\"\"",
		})
	}

	if goDocCount >= 5 {
		conventions = append(conventions, types.Convention{
			Category:    "documentation",
			Description: "Go doc comments (start with function name)",
			Example:     "// HandleRequest processes incoming HTTP requests",
		})
	}

	if xmlDocCount >= 5 {
		conventions = append(conventions, types.Convention{
			Category:    "documentation",
			Description: "XML documentation comments (C#)",
			Example:     "/// <summary>Handles the request</summary>",
		})
	}

	if todoCount >= 10 {
		conventions = append(conventions, types.Convention{
			Category:    "documentation",
			Description: "TODO/FIXME comments used for tracking work items",
		})
	}

	return conventions
}

// detectLoggingPatterns analyzes logging conventions
func (d *PatternDetector) detectLoggingPatterns() []types.Convention {
	var conventions []types.Convention

	// Logging patterns by language/framework with allowed extensions
	type logPattern struct {
		pattern    *regexp.Regexp
		extensions []string // allowed file extensions, empty means any
	}

	logPatterns := map[string]logPattern{
		"console":       {regexp.MustCompile(`console\.(log|info|warn|error|debug)\(`), []string{".js", ".ts", ".jsx", ".tsx"}},
		"winston":       {regexp.MustCompile(`(logger|log)\.(info|warn|error|debug)\(`), []string{".js", ".ts", ".jsx", ".tsx"}},
		"pino":          {regexp.MustCompile(`(logger|log)\.(info|warn|error|debug|fatal)\(`), []string{".js", ".ts", ".jsx", ".tsx"}},
		"log4j":         {regexp.MustCompile(`(logger|log)\.(info|warn|error|debug|trace)\(`), []string{".java"}},
		"slf4j":         {regexp.MustCompile(`(log|logger)\.(info|warn|error|debug)\(`), []string{".java"}},
		"python":        {regexp.MustCompile(`logging\.(info|warning|error|debug|critical)\(`), []string{".py"}},
		"python-logger": {regexp.MustCompile(`logger\.(info|warning|error|debug|critical)\(`), []string{".py"}},
		"go-log":        {regexp.MustCompile(`log\.(Print|Printf|Println|Fatal|Panic)\(`), []string{".go"}},
		"go-slog":       {regexp.MustCompile(`slog\.(Info|Warn|Error|Debug)\(`), []string{".go"}},
		"go-zap":        {regexp.MustCompile(`(logger|zap)\.(Info|Warn|Error|Debug)\(`), []string{".go"}},
		"go-zerolog":    {regexp.MustCompile(`(log|logger)\.(Info|Warn|Error|Debug)\(\)\.(Msg|Msgf)\(`), []string{".go"}},
		"rust-log":      {regexp.MustCompile(`(info|warn|error|debug|trace)!\(`), []string{".rs"}},
		"csharp":        {regexp.MustCompile(`(logger|_logger)\.(Log|LogInformation|LogWarning|LogError)\(`), []string{".cs"}},
		"ruby":          {regexp.MustCompile(`(logger|Rails\.logger)\.(info|warn|error|debug)\(`), []string{".rb"}},
	}

	counts := make(map[string]int)
	sampledFiles := 0
	maxSamples := 30

	for _, f := range d.files {
		if f.IsDir || sampledFiles >= maxSamples {
			continue
		}

		if !isSourceFile(f.Extension) {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)

		for name, logPat := range logPatterns {
			// Check if file extension matches allowed extensions
			if len(logPat.extensions) > 0 {
				matched := false
				for _, ext := range logPat.extensions {
					if f.Extension == ext {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}

			if logPat.pattern.MatchString(contentStr) {
				counts[name]++
			}
		}

		sampledFiles++
	}

	// Find dominant logging pattern
	var dominantLogger string
	maxCount := 0
	for name, count := range counts {
		if count > maxCount && count >= 3 {
			maxCount = count
			dominantLogger = name
		}
	}

	// Merge python-logger into python for reporting
	if dominantLogger == "python-logger" {
		dominantLogger = "python"
	}

	if dominantLogger != "" {
		loggerDescriptions := map[string]string{
			"console":    "console.log/warn/error for logging",
			"winston":    "Winston logger with structured logging",
			"pino":       "Pino logger (fast JSON logging)",
			"log4j":      "Log4j logging framework",
			"slf4j":      "SLF4J logging facade",
			"python":     "Python logging module",
			"go-log":     "Go standard library log package",
			"go-slog":    "Go structured logging (slog)",
			"go-zap":     "Uber's Zap logger",
			"go-zerolog": "Zerolog (zero-allocation JSON logging)",
			"rust-log":   "Rust log crate macros",
			"csharp":     "Microsoft.Extensions.Logging",
			"ruby":       "Ruby Logger / Rails.logger",
		}

		conventions = append(conventions, types.Convention{
			Category:    "logging",
			Description: loggerDescriptions[dominantLogger],
		})
	}

	return conventions
}

// detectErrorHandling analyzes error handling patterns
func (d *PatternDetector) detectErrorHandling() []types.Convention {
	var conventions []types.Convention

	// Error handling patterns
	tryCatchCount := 0
	goErrorCount := 0
	resultTypeCount := 0
	asyncAwaitCount := 0

	tryCatchRegex := regexp.MustCompile(`try\s*\{`)
	goErrorRegex := regexp.MustCompile(`if\s+err\s*!=\s*nil`)
	resultTypeRegex := regexp.MustCompile(`Result<|Result::`)
	asyncAwaitRegex := regexp.MustCompile(`async\s+(function|def|\(|=>)|await\s+`)

	sampledFiles := 0
	maxSamples := 30

	for _, f := range d.files {
		if f.IsDir || sampledFiles >= maxSamples {
			continue
		}

		if !isSourceFile(f.Extension) {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)

		if tryCatchRegex.MatchString(contentStr) {
			tryCatchCount++
		}
		if goErrorRegex.MatchString(contentStr) {
			goErrorCount++
		}
		if resultTypeRegex.MatchString(contentStr) {
			resultTypeCount++
		}
		if asyncAwaitRegex.MatchString(contentStr) {
			asyncAwaitCount++
		}

		sampledFiles++
	}

	if goErrorCount >= 5 {
		conventions = append(conventions, types.Convention{
			Category:    "error-handling",
			Description: "Go-style explicit error checking (if err != nil)",
			Example:     "if err != nil { return fmt.Errorf(\"context: %w\", err) }",
		})
	}

	if resultTypeCount >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "error-handling",
			Description: "Result/Option types for error handling (Rust-style)",
		})
	}

	if asyncAwaitCount >= 5 {
		conventions = append(conventions, types.Convention{
			Category:    "async",
			Description: "Async/await pattern for asynchronous operations",
		})
	}

	return conventions
}

// detectArchitecturalPatterns looks for architectural patterns
func (d *PatternDetector) detectArchitecturalPatterns() []types.Convention {
	var conventions []types.Convention

	// Check directory structure for patterns
	dirNames := make(map[string]bool)
	for _, f := range d.files {
		if f.IsDir {
			dirNames[strings.ToLower(f.Name)] = true
		}
	}

	// MVC pattern
	if dirNames["models"] && dirNames["views"] && dirNames["controllers"] {
		conventions = append(conventions, types.Convention{
			Category:    "architecture",
			Description: "MVC (Model-View-Controller) architecture",
		})
	}

	// Clean/Hexagonal architecture
	if dirNames["domain"] && (dirNames["infrastructure"] || dirNames["adapters"]) {
		conventions = append(conventions, types.Convention{
			Category:    "architecture",
			Description: "Clean/Hexagonal architecture (domain separation)",
		})
	}

	// Feature-based/Module-based
	if dirNames["features"] || dirNames["modules"] {
		conventions = append(conventions, types.Convention{
			Category:    "architecture",
			Description: "Feature/Module-based architecture",
		})
	}

	// Repository pattern
	if dirNames["repositories"] || dirNames["repository"] {
		conventions = append(conventions, types.Convention{
			Category:    "architecture",
			Description: "Repository pattern for data access",
		})
	}

	// Service layer
	if dirNames["services"] || dirNames["service"] {
		conventions = append(conventions, types.Convention{
			Category:    "architecture",
			Description: "Service layer for business logic",
		})
	}

	return conventions
}

// Helper function to check if file should be analyzed for documentation
func isDocumentableFile(ext string) bool {
	documentableExts := map[string]bool{
		".js": true, ".jsx": true, ".ts": true, ".tsx": true,
		".java": true, ".kt": true, ".scala": true,
		".py":  true,
		".go":  true,
		".rs":  true,
		".rb":  true,
		".cs":  true,
		".cpp": true, ".c": true, ".h": true, ".hpp": true,
		".swift": true,
		".php":   true,
	}
	return documentableExts[ext]
}
