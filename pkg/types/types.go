package types

// Analysis represents the complete analysis of a codebase
type Analysis struct {
	ProjectName   string            `json:"project_name"`
	RootPath      string            `json:"root_path"`
	TechStack     TechStack         `json:"tech_stack"`
	Structure     ProjectStructure  `json:"structure"`
	Conventions   []Convention      `json:"conventions"`
	Dependencies  []Dependency      `json:"dependencies"`
	Commands      []Command         `json:"commands"`
	KeyFiles      []KeyFile         `json:"key_files"`
	Endpoints     []Endpoint        `json:"endpoints,omitempty"`
	ReadmeContent *ReadmeContent    `json:"readme_content,omitempty"`
	MonorepoInfo  *MonorepoInfo     `json:"monorepo_info,omitempty"`
}

// ReadmeContent represents parsed README information
type ReadmeContent struct {
	Title        string   `json:"title,omitempty"`
	Description  string   `json:"description,omitempty"`
	Features     []string `json:"features,omitempty"`
	Installation string   `json:"installation,omitempty"`
	QuickStart   string   `json:"quick_start,omitempty"`
	Usage        string   `json:"usage,omitempty"`
}

// MonorepoInfo represents monorepo/workspace configuration
type MonorepoInfo struct {
	IsMonorepo     bool               `json:"is_monorepo"`
	Tool           string             `json:"tool,omitempty"`           // Turborepo, Lerna, Nx, etc.
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
	Languages   []Language  `json:"languages"`
	Frameworks  []Framework `json:"frameworks"`
	Databases   []string    `json:"databases"`
	Tools       []string    `json:"tools"`
}

// Language represents a programming language
type Language struct {
	Name       string `json:"name"`
	Version    string `json:"version,omitempty"`
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
	Path        string `json:"path"`
	Purpose     string `json:"purpose,omitempty"`
	FileCount   int    `json:"file_count"`
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
