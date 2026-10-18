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
	return &pipeline{root: root, cfg: cfg, files: files, deps: detector.ReadDependencies(root, files)}, nil
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
	conventions, err := detector.NewConventionDetector(p.root, p.files).Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect conventions: %w", err)
	}
	patterns, err := detector.NewPatternDetector(p.root, p.files).Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect patterns: %w", err)
	}
	framework, err := detector.NewFrameworkDetector(p.root, p.files).Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect framework patterns: %w", err)
	}
	return slices.Concat(conventions, patterns, framework), nil
}

// codePatterns runs the keyword, ML and AST detectors; React hooks and the
// Python AST results join the keyword results by name.
func (p *pipeline) codePatterns() *types.CodePatterns {
	patterns := detector.NewCodePatternDetector(p.root, p.files).Detect()
	patterns.MLPatterns = detector.NewMLDetector(p.root, p.files).GetMLPatterns()
	hooks := filterByCategory(detector.NewJSASTDetector(p.root, p.files).Detect(), "React Hooks")
	patterns.StateManagement = mergePatterns(patterns.StateManagement, hooks)
	patterns.PythonPatterns = mergePatterns(patterns.PythonPatterns, detector.NewPythonASTDetector(p.root, p.files).Detect())
	return patterns
}

// commands returns the script, Makefile and manifest commands, pyproject and
// Cargo included.
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
	return commands
}

// projectTools reads the tech stack, README and CLI info of a.
func (p *pipeline) projectTools(a *types.Analysis) []types.ProjectTool {
	return detector.NewProjectToolsDetector(p.root, p.files, a.ProjectName, &a.TechStack, a.ReadmeContent, a.CLIInfo).Detect()
}
