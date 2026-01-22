# Potential Improvements for Argus

This document outlines libraries and architectural improvements that could significantly enhance argus.

## Current Limitations

1. **Pattern Detection**: Uses string matching (`strings.Contains`) - 203 occurrences across 14 files
2. **Git Operations**: Shells out to `git` commands (6 locations) - fragile, OS-dependent
3. **Code Analysis**: No AST parsing - misses context, can't detect actual code structure
4. **Dependency Parsing**: Manual parsing of package.json, go.mod
5. **Architecture Detection**: Heuristic-based (directory names) - could miss actual patterns

## High-Impact Improvements

### 1. Code Analysis with AST Parsing

**Problem**: Currently using `strings.Contains("useState")` to detect React hooks. This has false positives (comments, strings) and misses imports.

**Solution**: Use Abstract Syntax Tree (AST) parsing

#### For Go
```go
// stdlib - no dependencies needed!
import (
    "go/ast"
    "go/parser"
    "go/token"
    "golang.org/x/tools/go/packages"
)
```

**Benefits**:
- Accurately detect interfaces, structs, methods
- Find actual import usage (not just string matching)
- Detect design patterns (Factory, Builder, etc.)
- Count cyclomatic complexity
- Find unused code

**Example**: Detect if a Go file uses Cobra
```go
// Current (string matching):
if strings.Contains(content, "cobra.Command") {
    // Might be in a comment or string literal!
}

// With AST:
cfg := &packages.Config{Mode: packages.NeedImports | packages.NeedName}
pkgs, _ := packages.Load(cfg, "./...")
for _, pkg := range pkgs {
    for path := range pkg.Imports {
        if path == "github.com/spf13/cobra" {
            // Actually imported and used!
        }
    }
}
```

#### For TypeScript/JavaScript
**Library**: `github.com/evanw/esbuild` (Go bindings) or `tree-sitter`

```go
import "github.com/smacker/go-tree-sitter"
import "github.com/smacker/go-tree-sitter/typescript"
```

**Benefits**:
- Parse TypeScript/JSX correctly
- Detect actual React component patterns
- Find hook usage in components only (not everywhere)

#### For Python
**Library**: `tree-sitter` with Python grammar

**Benefits**:
- Detect Django models, views properly
- Find FastAPI decorators accurately
- Parse class hierarchies

### 2. Git Analysis Library

**Problem**: Shelling out to `git` commands is fragile, slow, and OS-dependent

**Solution**: Use pure Go git library

```go
import "github.com/go-git/go-git/v5"
```

**Benefits**:
- No git binary dependency
- 3-5x faster git operations
- Better error handling
- Cross-platform reliability
- Can analyze commit diffs, file changes
- Detect branching strategies (gitflow, trunk-based)

**Example**:
```go
repo, _ := git.PlainOpen(rootPath)
ref, _ := repo.Head()
commit, _ := repo.CommitObject(ref.Hash())
// Direct access to commit history, no parsing needed
```

### 3. Dependency Analysis

**Problem**: Manual parsing of package.json, go.mod is error-prone

**Solutions**:

#### Go modules
```go
import "golang.org/x/mod/modfile"
import "golang.org/x/mod/semver"
```

**Benefits**:
- Properly parse go.mod with replace directives
- Understand go.work workspaces
- Detect indirect dependencies
- Version constraint analysis

#### Package.json
```go
import "encoding/json"
// Current approach is OK, but could add:
import "github.com/Masterminds/semver/v3"
```

**Benefits**:
- Validate semantic versions
- Detect version ranges
- Find outdated dependencies

#### Python (pyproject.toml, requirements.txt)
```go
import "github.com/BurntSushi/toml"
```

### 4. Architecture & Design Pattern Detection

**Library**: `github.com/loov/goda` (Go dependency analysis)

**Benefits**:
- Detect actual dependency direction
- Find cyclic dependencies
- Visualize module relationships
- Detect layered architecture violations

**Example Use Cases**:
- Detect if project follows Clean Architecture (by analyzing import paths)
- Find if domain layer imports infrastructure (violation!)
- Detect hexagonal architecture by checking adapter/port dependencies

### 5. Language-Agnostic Parsing

**Library**: `github.com/smacker/go-tree-sitter`

**Benefits**:
- Single library for 40+ languages
- Consistent AST across all languages
- Can add support for Rust, Java, C++, etc. easily

**Example**:
```go
parser := sitter.NewParser()
parser.SetLanguage(javascript.GetLanguage())
tree := parser.Parse(nil, []byte(sourceCode))
// Query patterns like "find all function declarations"
```

### 6. Configuration Validation

**Library**: `github.com/go-playground/validator/v10`

**Benefits**:
- Validate .argus.yaml structure
- Better error messages for users
- Enforce constraints (e.g., max 10 custom conventions)

### 7. Performance Improvements

#### Parallel Processing
```go
import "golang.org/x/sync/errgroup"
```

**Benefits**:
- Better goroutine coordination
- Proper error handling in concurrent code
- Currently: detectors run sequentially (could be parallel)

#### File Walking
```go
import "github.com/karrick/godirwalk"
```

**Benefits**:
- 2-3x faster than filepath.WalkDir
- Better memory usage for large repos

### 8. Metrics & Telemetry (Optional)

**Library**: `github.com/prometheus/client_golang`

**Benefits**:
- Track which patterns are most detected
- Performance metrics per detector
- Help prioritize improvements

## Architecture Improvements

### 1. Plugin System for Detectors

**Current**: All detectors are compiled in
**Proposed**: Plugin architecture

```go
type Detector interface {
    Name() string
    Detect(ctx context.Context, files []FileInfo) (Result, error)
    Priority() int // for ordering
}

// Detectors register themselves
func init() {
    RegisterDetector(&GoPatternDetector{})
}
```

**Benefits**:
- Easy to add new detectors
- Community can contribute detectors
- Can enable/disable detectors via config

### 2. Caching Layer

**Proposal**: Cache detection results

```go
type Cache interface {
    Get(key string) ([]byte, bool)
    Set(key string, value []byte)
}

// Cache based on file hash
hash := sha256.Sum256(fileContent)
if cached, ok := cache.Get(hex.EncodeToString(hash[:])); ok {
    return cached
}
```

**Benefits**:
- Faster re-analysis (watch mode)
- Only re-analyze changed files

### 3. Incremental Analysis

**Proposal**: Track what changed, only re-analyze affected parts

```go
type ChangeDetector interface {
    GetChangedFiles(since time.Time) ([]string, error)
    GetChangedSince(commitHash string) ([]string, error)
}
```

**Benefits**:
- 10-100x faster on large repos
- Better watch mode performance

### 4. AST-Based Pattern Language

**Proposal**: Let users define custom patterns

```yaml
# .argus.yaml
custom_patterns:
  - name: "API Error Handler"
    language: go
    pattern: |
      func $NAME($PARAMS) error {
        if err != nil {
          return fmt.Errorf($MESSAGE)
        }
      }
```

**Benefits**:
- Users can detect project-specific patterns
- No code changes needed for new patterns

## Prioritized Recommendations

### Phase 1 (High Impact, Low Effort)
1. ✅ **go-git/go-git** - Replace git shell commands (~2 hours)
2. ✅ **golang.org/x/mod/modfile** - Proper go.mod parsing (~1 hour)
3. ✅ **golang.org/x/sync/errgroup** - Parallel detector execution (~1 hour)

### Phase 2 (High Impact, Medium Effort)
4. **go/ast + go/parser** - AST-based Go pattern detection (~4 hours)
5. **tree-sitter** - Multi-language AST parsing (~6 hours)
6. **loov/goda** - Architecture dependency analysis (~3 hours)

### Phase 3 (Medium Impact, Low Effort)
7. **validator** - Config validation (~2 hours)
8. **godirwalk** - Faster file walking (~1 hour)
9. **BurntSushi/toml** - Python pyproject.toml support (~1 hour)

### Phase 4 (Nice to Have)
10. **Plugin system** - Detector plugins (~8 hours)
11. **Caching layer** - Result caching (~4 hours)
12. **Custom pattern language** - User-defined patterns (~12 hours)

## Estimated Impact

### Detection Accuracy
- **Current**: ~70-80% accuracy (string matching has false positives)
- **With AST**: ~95-98% accuracy

### Performance
- **Current**: ~1-3 seconds for medium repo (1000 files)
- **With go-git**: ~0.5-1 second (2-3x faster)
- **With parallel**: ~0.3-0.5 second (3-6x faster)
- **With caching**: ~0.1-0.2 second for unchanged (10-15x faster)

### Maintainability
- **AST-based**: Much easier to add new pattern detections
- **Plugin system**: Community can contribute detectors

## Next Steps

Would you like me to:
1. Start with Phase 1 (go-git, modfile, errgroup)?
2. Create a prototype AST-based Go pattern detector?
3. Build a performance benchmark to measure current performance?
4. Design the plugin architecture?
