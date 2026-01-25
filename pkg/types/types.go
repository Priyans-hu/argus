package types

// Analysis represents the complete analysis of a codebase
type Analysis struct {
	ProjectName      string            `json:"project_name"`
	RootPath         string            `json:"root_path"`
	TechStack        TechStack         `json:"tech_stack"`
	Structure        ProjectStructure  `json:"structure"`
	Conventions      []Convention      `json:"conventions"`
	Dependencies     []Dependency      `json:"dependencies"`
	Commands         []Command         `json:"commands"`
	KeyFiles         []KeyFile         `json:"key_files"`
	Endpoints        []Endpoint        `json:"endpoints,omitempty"`
	ReadmeContent    *ReadmeContent    `json:"readme_content,omitempty"`
	MonorepoInfo     *MonorepoInfo     `json:"monorepo_info,omitempty"`
	CodePatterns     *CodePatterns     `json:"code_patterns,omitempty"`
	GitConventions   *GitConventions   `json:"git_conventions,omitempty"`
	ArchitectureInfo *ArchitectureInfo `json:"architecture_info,omitempty"`
	DevelopmentInfo  *DevelopmentInfo  `json:"development_info,omitempty"`
	ConfigFiles      []ConfigFileInfo  `json:"config_files,omitempty"`
	CLIInfo          *CLIInfo          `json:"cli_info,omitempty"`
	ProjectTools     []ProjectTool     `json:"project_tools,omitempty"`
}

// ReadmeContent represents parsed README information
type ReadmeContent struct {
	Title         string            `json:"title,omitempty"`
	Description   string            `json:"description,omitempty"`
	Features      []string          `json:"features,omitempty"`
	Installation  string            `json:"installation,omitempty"`
	QuickStart    string            `json:"quick_start,omitempty"`
	Usage         string            `json:"usage,omitempty"`
	Prerequisites []string          `json:"prerequisites,omitempty"` // Required tools/dependencies
	KeyCommands   []string          `json:"key_commands,omitempty"`  // Important commands from code blocks
	ModelSpecs    map[string]string `json:"model_specs,omitempty"`   // ML model specifications
	ProjectType   string            `json:"project_type,omitempty"`  // docs, ml, cli, library, app
}

// MonorepoInfo represents monorepo/workspace configuration
type MonorepoInfo struct {
	IsMonorepo     bool               `json:"is_monorepo"`
	Tool           string             `json:"tool,omitempty"`            // Turborepo, Lerna, Nx, etc.
	PackageManager string             `json:"package_manager,omitempty"` // npm, yarn, pnpm, bun
	WorkspacePaths []string           `json:"workspace_paths,omitempty"`
	Packages       []WorkspacePackage `json:"packages,omitempty"`
}

// WorkspacePackage represents a package in a monorepo
type WorkspacePackage struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Description string   `json:"description,omitempty"`
	SubPackages []string `json:"sub_packages,omitempty"`
}

// CodePatterns represents detected code patterns from deep analysis
type CodePatterns struct {
	StateManagement []PatternInfo `json:"state_management,omitempty"`
	DataFetching    []PatternInfo `json:"data_fetching,omitempty"`
	Routing         []PatternInfo `json:"routing,omitempty"`
	Forms           []PatternInfo `json:"forms,omitempty"`
	Testing         []PatternInfo `json:"testing,omitempty"`
	Styling         []PatternInfo `json:"styling,omitempty"`
	Authentication  []PatternInfo `json:"authentication,omitempty"`
	APIPatterns     []PatternInfo `json:"api_patterns,omitempty"`
	DatabaseORM     []PatternInfo `json:"database_orm,omitempty"`
	Utilities       []PatternInfo `json:"utilities,omitempty"`
	GoPatterns      []PatternInfo `json:"go_patterns,omitempty"`
	RustPatterns    []PatternInfo `json:"rust_patterns,omitempty"`
	PythonPatterns  []PatternInfo `json:"python_patterns,omitempty"`
	MLPatterns      []PatternInfo `json:"ml_patterns,omitempty"`
}

// PatternInfo represents a detected pattern
type PatternInfo struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	FileCount   int      `json:"file_count"`
	Examples    []string `json:"examples,omitempty"`
	Usage       string   `json:"usage,omitempty"`
}

// Endpoint represents an API endpoint
type Endpoint struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Handler     string `json:"handler,omitempty"`
	File        string `json:"file"`
	Line        int    `json:"line,omitempty"`
	Auth        string `json:"auth,omitempty"`
	Description string `json:"description,omitempty"`
}

// TechStack represents detected technologies
type TechStack struct {
	Languages  []Language  `json:"languages"`
	Frameworks []Framework `json:"frameworks"`
	Databases  []string    `json:"databases"`
	Tools      []string    `json:"tools"`
}

// Language represents a programming language
type Language struct {
	Name       string  `json:"name"`
	Version    string  `json:"version,omitempty"`
	Percentage float64 `json:"percentage,omitempty"`
}

// Framework represents a framework or library
type Framework struct {
	Name     string `json:"name"`
	Version  string `json:"version,omitempty"`
	Category string `json:"category"` // frontend, backend, testing, etc.
}

// ProjectStructure represents the directory layout
type ProjectStructure struct {
	Directories []Directory `json:"directories"`
	RootFiles   []string    `json:"root_files"`
}

// Directory represents a directory in the project
type Directory struct {
	Path      string `json:"path"`
	Purpose   string `json:"purpose,omitempty"`
	FileCount int    `json:"file_count"`
}

// Convention represents a detected coding convention
type Convention struct {
	Category    string `json:"category"` // naming, imports, structure, etc.
	Description string `json:"description"`
	Example     string `json:"example,omitempty"`
}

// Dependency represents a project dependency
type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"` // runtime, dev, peer, etc.
}

// Command represents an available command/script
type Command struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	Description string `json:"description,omitempty"`
}

// ProjectTool represents a project-specific CLI tool that deserves a skill
type ProjectTool struct {
	Name              string   `json:"name"`                         // Tool name (e.g., "grepai", "argus")
	BinaryPath        string   `json:"binary_path,omitempty"`        // Path to binary if built by project
	Description       string   `json:"description"`                  // What the tool does
	UsageExamples     []string `json:"usage_examples,omitempty"`     // Example commands
	WhenToUse         string   `json:"when_to_use,omitempty"`        // When Claude should use this tool
	ReplacesTool      string   `json:"replaces_tool,omitempty"`      // Tool it replaces (e.g., "Grep", "Glob")
	RequiresSetup     bool     `json:"requires_setup,omitempty"`     // Whether tool needs installation/setup
	SetupInstructions string   `json:"setup_instructions,omitempty"` // How to set it up
}

// KeyFile represents an important file in the project
type KeyFile struct {
	Path        string `json:"path"`
	Purpose     string `json:"purpose"`
	Description string `json:"description,omitempty"`
}

// Config represents Argus configuration
type Config struct {
	Output            []string          `yaml:"output"`
	Ignore            []string          `yaml:"ignore"`
	CustomConventions []string          `yaml:"custom_conventions"`
	Overrides         map[string]string `yaml:"overrides"`
}

// Detector interface for all detection modules
type Detector interface {
	Name() string
	Detect(ctx *AnalysisContext) error
}

// Generator interface for all output generators
type Generator interface {
	Name() string
	OutputFile() string
	Generate(analysis *Analysis) ([]byte, error)
}

// AnalysisContext holds context during analysis
type AnalysisContext struct {
	RootPath string
	Config   *Config
	Analysis *Analysis
	Files    []FileInfo
}

// FileInfo represents a file in the project
type FileInfo struct {
	Path      string
	Name      string
	Extension string
	Size      int64
	IsDir     bool
}

// GitRepository represents git repository information
type GitRepository struct {
	RemoteURL string `json:"remote_url,omitempty"` // e.g., https://github.com/user/repo.git
	Owner     string `json:"owner,omitempty"`      // e.g., user or organization
	Name      string `json:"name,omitempty"`       // e.g., repo name
	Platform  string `json:"platform,omitempty"`   // e.g., github, gitlab, bitbucket
}

// GitCommit represents a git commit
type GitCommit struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Author  string `json:"author,omitempty"`
	Date    string `json:"date,omitempty"`
}

// GitConventions represents detected git conventions
type GitConventions struct {
	CommitConvention *CommitConvention `json:"commit_convention,omitempty"`
	BranchConvention *BranchConvention `json:"branch_convention,omitempty"`
	Repository       *GitRepository    `json:"repository,omitempty"`
	RecentCommits    []GitCommit       `json:"recent_commits,omitempty"`
}

// CommitConvention represents detected commit message conventions
type CommitConvention struct {
	Style   string   `json:"style"`            // conventional, gitmoji, jira, angular
	Format  string   `json:"format"`           // e.g., "<type>(<scope>): <description>"
	Types   []string `json:"types,omitempty"`  // feat, fix, docs, etc.
	Scopes  []string `json:"scopes,omitempty"` // api, ui, core, etc.
	Example string   `json:"example"`
}

// BranchConvention represents detected branch naming conventions
type BranchConvention struct {
	Prefixes []string `json:"prefixes"` // feat, fix, chore, etc.
	Format   string   `json:"format"`   // e.g., "<prefix>/<description>"
	Examples []string `json:"examples,omitempty"`
}

// ArchitectureInfo represents detected architecture
type ArchitectureInfo struct {
	Style      string              `json:"style,omitempty"`       // layered, modular, clean, etc.
	Layers     []ArchitectureLayer `json:"layers,omitempty"`      // detected layers
	EntryPoint string              `json:"entry_point,omitempty"` // main entry point
	Diagram    string              `json:"diagram,omitempty"`     // text-based diagram
}

// ArchitectureLayer represents an architectural layer
type ArchitectureLayer struct {
	Name      string   `json:"name"`
	Purpose   string   `json:"purpose,omitempty"`
	Packages  []string `json:"packages,omitempty"`
	DependsOn []string `json:"depends_on,omitempty"`
}

// GeneratedFile represents a file to be written by a multi-file generator
type GeneratedFile struct {
	Path    string // Relative path, e.g., ".claude/agents/go-reviewer.md"
	Content []byte
}

// MultiFileGenerator generates multiple output files
type MultiFileGenerator interface {
	Name() string
	Generate(analysis *Analysis) ([]GeneratedFile, error)
}

// DevelopmentInfo captures development environment setup information
type DevelopmentInfo struct {
	Prerequisites []Prerequisite `json:"prerequisites,omitempty"`
	SetupSteps    []SetupStep    `json:"setup_steps,omitempty"`
	GitHooks      []GitHook      `json:"git_hooks,omitempty"`
}

// Prerequisite represents a required tool/runtime
type Prerequisite struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// SetupStep represents a setup instruction
type SetupStep struct {
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
}

// GitHook represents a git hook configuration
type GitHook struct {
	Name    string   `json:"name"`
	Actions []string `json:"actions,omitempty"`
}

// ConfigFileInfo represents a configuration file in the project
type ConfigFileInfo struct {
	Path    string `json:"path"`
	Type    string `json:"type"`
	Purpose string `json:"purpose"`
}

// CLIInfo represents CLI tool information
type CLIInfo struct {
	VerboseFlag string      `json:"verbose_flag,omitempty"`
	DryRunFlag  string      `json:"dry_run_flag,omitempty"`
	Indicators  []Indicator `json:"indicators,omitempty"`
}

// Indicator represents a CLI output indicator
type Indicator struct {
	Symbol  string `json:"symbol"`
	Meaning string `json:"meaning"`
}

// ClaudeCodeConfig controls what Claude Code configs to generate
type ClaudeCodeConfig struct {
	Agents   bool `yaml:"agents"`
	Commands bool `yaml:"commands"`
	Rules    bool `yaml:"rules"`
}
