package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Priyans-hu/argus/pkg/types"
)

// CargoToml represents Cargo.toml structure
type CargoToml struct {
	Package           CargoPackage            `toml:"package"`
	Lib               *CargoLib               `toml:"lib"`
	Bin               []CargoBin              `toml:"bin"`
	Workspace         *CargoWorkspace         `toml:"workspace"`
	Dependencies      map[string]interface{}  `toml:"dependencies"`
	DevDependencies   map[string]interface{}  `toml:"dev-dependencies"`
	BuildDependencies map[string]interface{}  `toml:"build-dependencies"`
	Features          map[string][]string     `toml:"features"`
	Profile           map[string]CargoProfile `toml:"profile"`
}

// CargoPackage represents [package] section
type CargoPackage struct {
	Name          string   `toml:"name"`
	Version       string   `toml:"version"`
	Edition       string   `toml:"edition"`
	Authors       []string `toml:"authors"`
	Description   string   `toml:"description"`
	License       string   `toml:"license"`
	Repository    string   `toml:"repository"`
	Documentation string   `toml:"documentation"`
	Homepage      string   `toml:"homepage"`
	Readme        string   `toml:"readme"`
	Keywords      []string `toml:"keywords"`
	Categories    []string `toml:"categories"`
	RustVersion   string   `toml:"rust-version"`
}

// CargoLib represents [lib] section
type CargoLib struct {
	Name      string   `toml:"name"`
	Path      string   `toml:"path"`
	CrateType []string `toml:"crate-type"`
}

// CargoBin represents [[bin]] section
type CargoBin struct {
	Name string `toml:"name"`
	Path string `toml:"path"`
}

// CargoWorkspace represents [workspace] section
type CargoWorkspace struct {
	Members  []string `toml:"members"`
	Exclude  []string `toml:"exclude"`
	Resolver string   `toml:"resolver"`
}

// CargoProfile represents [profile.*] section
type CargoProfile struct {
	OptLevel    interface{} `toml:"opt-level"`
	Debug       interface{} `toml:"debug"`
	LTO         interface{} `toml:"lto"`
	Codegen     int         `toml:"codegen-units"`
	Panic       string      `toml:"panic"`
	Incremental bool        `toml:"incremental"`
	Overflow    string      `toml:"overflow-checks"`
}

// CargoDetector detects Rust project info from Cargo.toml
type CargoDetector struct {
	rootPath string
}

// NewCargoDetector creates a new Cargo.toml detector
func NewCargoDetector(rootPath string) *CargoDetector {
	return &CargoDetector{rootPath: rootPath}
}

// CargoInfo holds parsed Cargo.toml information
type CargoInfo struct {
	HasCargo         bool
	Name             string
	Version          string
	Edition          string
	RustVersion      string
	Description      string
	IsWorkspace      bool
	WorkspaceMembers []string
	IsBinary         bool
	IsLibrary        bool
	Binaries         []string
	Dependencies     []string
	DevDependencies  []string
	Features         []string
}

// Detect parses Cargo.toml and returns project info
func (d *CargoDetector) Detect() *CargoInfo {
	cargoPath := filepath.Join(d.rootPath, "Cargo.toml")

	content, err := os.ReadFile(cargoPath)
	if err != nil {
		return nil
	}

	var cargo CargoToml
	if _, err := toml.Decode(string(content), &cargo); err != nil {
		return nil
	}

	info := &CargoInfo{
		HasCargo:    true,
		Name:        cargo.Package.Name,
		Version:     cargo.Package.Version,
		Edition:     cargo.Package.Edition,
		RustVersion: cargo.Package.RustVersion,
		Description: cargo.Package.Description,
	}

	// Check for workspace
	if cargo.Workspace != nil {
		info.IsWorkspace = true
		info.WorkspaceMembers = cargo.Workspace.Members
	}

	// Check for library
	if cargo.Lib != nil {
		info.IsLibrary = true
	}

	// Check for binaries
	if len(cargo.Bin) > 0 {
		info.IsBinary = true
		for _, bin := range cargo.Bin {
			info.Binaries = append(info.Binaries, bin.Name)
		}
	} else if cargo.Package.Name != "" {
		// Default binary with package name
		srcMain := filepath.Join(d.rootPath, "src", "main.rs")
		if _, err := os.Stat(srcMain); err == nil {
			info.IsBinary = true
			info.Binaries = append(info.Binaries, cargo.Package.Name)
		}
	}

	// Extract dependencies
	info.Dependencies = extractCargoDeps(cargo.Dependencies)
	info.DevDependencies = extractCargoDeps(cargo.DevDependencies)

	// Extract features
	for feature := range cargo.Features {
		info.Features = append(info.Features, feature)
	}

	return info
}

// extractCargoDeps extracts dependency names from Cargo dependency format
func extractCargoDeps(deps map[string]interface{}) []string {
	var result []string
	for name := range deps {
		result = append(result, name)
	}
	return result
}

// DetectRustPatterns detects Rust patterns from Cargo.toml
func (d *CargoDetector) DetectRustPatterns() []types.PatternInfo {
	info := d.Detect()
	if info == nil {
		return nil
	}

	var patterns []types.PatternInfo

	// Project type
	if info.IsWorkspace {
		patterns = append(patterns, types.PatternInfo{
			Name:        "Cargo Workspace",
			Category:    "rust-structure",
			Description: "Multi-crate Rust workspace",
			FileCount:   len(info.WorkspaceMembers),
			Examples:    info.WorkspaceMembers,
		})
	}

	if info.IsBinary && info.IsLibrary {
		patterns = append(patterns, types.PatternInfo{
			Name:        "Binary + Library",
			Category:    "rust-structure",
			Description: "Rust project with both binary and library targets",
			FileCount:   1,
		})
	} else if info.IsBinary {
		patterns = append(patterns, types.PatternInfo{
			Name:        "Binary Crate",
			Category:    "rust-structure",
			Description: "Rust executable binary",
			FileCount:   len(info.Binaries),
			Examples:    info.Binaries,
		})
	} else if info.IsLibrary {
		patterns = append(patterns, types.PatternInfo{
			Name:        "Library Crate",
			Category:    "rust-structure",
			Description: "Rust library crate",
			FileCount:   1,
		})
	}

	// Detect major dependencies
	frameworkDeps := map[string]string{
		"tokio":      "Tokio async runtime",
		"async-std":  "async-std runtime",
		"actix-web":  "Actix-web framework",
		"axum":       "Axum web framework",
		"rocket":     "Rocket web framework",
		"warp":       "Warp web framework",
		"hyper":      "Hyper HTTP library",
		"reqwest":    "Reqwest HTTP client",
		"serde":      "Serde serialization",
		"serde_json": "Serde JSON",
		"clap":       "Clap CLI framework",
		"structopt":  "StructOpt CLI (deprecated)",
		"diesel":     "Diesel ORM",
		"sqlx":       "SQLx async SQL",
		"sea-orm":    "SeaORM async ORM",
		"mongodb":    "MongoDB driver",
		"redis":      "Redis client",
		"tracing":    "Tracing framework",
		"log":        "Log facade",
		"anyhow":     "Anyhow error handling",
		"thiserror":  "thiserror derive",
		"rayon":      "Rayon parallelism",
		"crossbeam":  "Crossbeam concurrency",
		"regex":      "Regex library",
		"chrono":     "Chrono date/time",
		"uuid":       "UUID generation",
	}

	allDeps := append(info.Dependencies, info.DevDependencies...)
	for _, dep := range allDeps {
		// Normalize dependency name (remove underscores/hyphens)
		depNorm := strings.ReplaceAll(strings.ReplaceAll(dep, "-", ""), "_", "")
		for frameworkDep, desc := range frameworkDeps {
			frameworkNorm := strings.ReplaceAll(strings.ReplaceAll(frameworkDep, "-", ""), "_", "")
			if depNorm == frameworkNorm || dep == frameworkDep {
				patterns = append(patterns, types.PatternInfo{
					Name:        dep,
					Category:    "rust-dependency",
					Description: desc,
					FileCount:   1,
					Examples:    []string{"Cargo.toml"},
				})
				break
			}
		}
	}

	return patterns
}

// DetectCargoCommands returns Rust-specific commands based on Cargo.toml
func (d *CargoDetector) DetectCargoCommands() []types.Command {
	info := d.Detect()
	if info == nil {
		return nil
	}

	var commands []types.Command

	// Basic commands
	commands = append(commands, types.Command{
		Name:        "cargo build",
		Description: "Build the project in debug mode",
	})
	commands = append(commands, types.Command{
		Name:        "cargo build --release",
		Description: "Build optimized release binary",
	})
	commands = append(commands, types.Command{
		Name:        "cargo run",
		Description: "Build and run the project",
	})
	commands = append(commands, types.Command{
		Name:        "cargo test",
		Description: "Run all tests",
	})
	commands = append(commands, types.Command{
		Name:        "cargo test -- --nocapture",
		Description: "Run tests with output",
	})
	commands = append(commands, types.Command{
		Name:        "cargo fmt",
		Description: "Format code with rustfmt",
	})
	commands = append(commands, types.Command{
		Name:        "cargo clippy",
		Description: "Run Clippy linter",
	})
	commands = append(commands, types.Command{
		Name:        "cargo doc --open",
		Description: "Generate and open documentation",
	})

	// Check for specific binaries
	if len(info.Binaries) > 1 {
		for _, bin := range info.Binaries {
			commands = append(commands, types.Command{
				Name:        "cargo run --bin " + bin,
				Description: "Run " + bin + " binary",
			})
		}
	}

	// Workspace commands
	if info.IsWorkspace {
		commands = append(commands, types.Command{
			Name:        "cargo build --workspace",
			Description: "Build all workspace members",
		})
		commands = append(commands, types.Command{
			Name:        "cargo test --workspace",
			Description: "Test all workspace members",
		})
	}

	// Check for common dev tools in dependencies
	for _, dep := range info.DevDependencies {
		switch dep {
		case "cargo-watch":
			commands = append(commands, types.Command{
				Name:        "cargo watch -x run",
				Description: "Watch and run on changes",
			})
		case "cargo-nextest":
			commands = append(commands, types.Command{
				Name:        "cargo nextest run",
				Description: "Run tests with nextest",
			})
		}
	}

	// Check for features
	if len(info.Features) > 0 {
		featuresStr := strings.Join(info.Features[:min(3, len(info.Features))], ",")
		commands = append(commands, types.Command{
			Name:        "cargo build --features " + featuresStr,
			Description: "Build with specific features",
		})
	}

	return commands
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// DetectRustVersion tries to detect the Rust version from various sources
func (d *CargoDetector) DetectRustVersion() string {
	info := d.Detect()
	if info != nil && info.RustVersion != "" {
		return info.RustVersion
	}

	// Check rust-toolchain.toml
	toolchainPath := filepath.Join(d.rootPath, "rust-toolchain.toml")
	if data, err := os.ReadFile(toolchainPath); err == nil {
		re := regexp.MustCompile(`channel\s*=\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(string(data)); len(matches) > 1 {
			return matches[1]
		}
	}

	// Check rust-toolchain file
	toolchainFilePath := filepath.Join(d.rootPath, "rust-toolchain")
	if data, err := os.ReadFile(toolchainFilePath); err == nil {
		return strings.TrimSpace(string(data))
	}

	return ""
}
