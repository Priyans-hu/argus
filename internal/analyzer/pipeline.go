package analyzer

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/Priyans-hu/argus/internal/detector"
	"github.com/Priyans-hu/argus/pkg/types"
)

// pipeline holds the inputs of one analysis and the steps the sequential,
// parallel and incremental analyzers share, so they differ only in
// scheduling.
type pipeline struct {
	root  string // absolute
	cfg   *types.Config
	files []types.FileInfo
	src   []types.FileInfo // project source: the files pattern detectors read
	deps  []types.Dependency
}

// newPipeline resolves the root, walks it and reads the declared
// dependencies, which detectors take as evidence.
func newPipeline(ctx context.Context, rootPath string, cfg *types.Config) (*pipeline, error) {
	root, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	files, err := NewWalker(root, cfg).Walk(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}
	return &pipeline{root: root, cfg: cfg, files: files, src: detector.ProjectSourceFiles(files), deps: detector.ReadDependencies(root, files)}, nil
}

// newAnalysis returns the fields known before any detector runs.
func (p *pipeline) newAnalysis() *types.Analysis {
	return &types.Analysis{ProjectName: p.projectName(), RootPath: p.root, Dependencies: p.deps}
}

func (p *pipeline) projectName() string {
	return filepath.Base(p.root)
}

func (p *pipeline) techStack() (types.TechStack, error) {
	stack, err := detector.NewTechStackDetector(p.root, p.files, p.deps).Detect()
	if err != nil {
		return types.TechStack{}, fmt.Errorf("failed to detect tech stack: %w", err)
	}
	return *stack, nil
}

// conventions returns the conventions, then the patterns, then the
// framework patterns.
func (p *pipeline) conventions() ([]types.Convention, error) {
	conventions, err := detector.NewConventionDetector(p.root, p.src).Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect conventions: %w", err)
	}
	patterns, err := detector.NewPatternDetector(p.root, p.src, p.deps).Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect patterns: %w", err)
	}
	framework, err := detector.NewFrameworkDetector(p.root, p.src).Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect framework patterns: %w", err)
	}
	return slices.Concat(conventions, patterns, framework), nil
}

// codePatterns runs the keyword, ML and AST detectors on project source;
// React hooks, reported only with React or Preact declared, and the Python
// and Go AST results join the keyword results by name. The JS and Python AST
// detectors skip test files, which feed only testing patterns.
func (p *pipeline) codePatterns() *types.CodePatterns {
	patterns := detector.NewCodePatternDetector(p.root, p.src, p.deps).Detect()
	patterns.MLPatterns = detector.NewMLDetector(p.root, p.src).GetMLPatterns()
	nonTest := slices.DeleteFunc(slices.Clone(p.src), func(f types.FileInfo) bool { return detector.IsTestFile(f.Path) })
	if detector.NewDepSet(p.deps).Has("react", "preact") {
		hooks := filterByCategory(detector.NewJSASTDetector(p.root, nonTest).Detect(), "React Hooks")
		patterns.StateManagement = mergePatterns(patterns.StateManagement, hooks)
	}
	patterns.PythonPatterns = mergePatterns(patterns.PythonPatterns, detector.NewPythonASTDetector(p.root, nonTest).Detect())
	mergeGoAST(patterns, detector.NewGoASTDetector(p.root, p.src).Detect())
	return patterns
}

// mergeGoAST files the Go AST results by name: web libraries in Routing,
// database libraries in Database & ORM, testing libraries in Testing, the
// rest in Go Patterns under their go-* category.
func mergeGoAST(patterns *types.CodePatterns, found []types.PatternInfo) {
	for _, p := range found {
		switch p.Category {
		case "go-web":
			p.Category = "routing"
			patterns.Routing = mergePatterns(patterns.Routing, []types.PatternInfo{p})
		case "go-database":
			p.Category = "database"
			patterns.DatabaseORM = mergePatterns(patterns.DatabaseORM, []types.PatternInfo{p})
		case "go-testing":
			p.Category = "testing"
			patterns.Testing = mergePatterns(patterns.Testing, []types.PatternInfo{p})
		default:
			patterns.GoPatterns = mergePatterns(patterns.GoPatterns, []types.PatternInfo{p})
		}
	}
}

// commands returns the script, Makefile and manifest commands, pyproject and
// Cargo included, then Gradle and the crates one directory deep.
func (p *pipeline) commands() []types.Command {
	commands := detector.DetectCommands(p.root)
	if pyInfo := detector.NewPyProjectDetector(p.root).Detect(); pyInfo != nil && pyInfo.HasPyProject {
		commands = append(commands, detectPyProjectCommands(pyInfo)...)
	}
	cargo := detector.NewCargoDetector(p.root)
	if info := cargo.Detect(); info != nil && info.HasCargo {
		// the Cargo detector's commands replace the basic cargo ones
		commands = append(filterNonCargoCommands(commands), cargo.DetectCargoCommands()...)
	}
	commands = append(commands, detector.GradleCommands(p.root, p.files)...)
	return append(commands, detector.SubdirCargoCommands(p.root, p.files)...)
}

// projectTools reads the tech stack, README and CLI info of a.
func (p *pipeline) projectTools(a *types.Analysis) []types.ProjectTool {
	return detector.NewProjectToolsDetector(p.root, p.files, a.ProjectName, &a.TechStack, a.ReadmeContent, a.CLIInfo).Detect()
}
