package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// EndpointDetector detects API endpoints from various frameworks
type EndpointDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewEndpointDetector creates a new endpoint detector
func NewEndpointDetector(rootPath string, files []types.FileInfo) *EndpointDetector {
	return &EndpointDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// Detect finds all API endpoints in the codebase
func (d *EndpointDetector) Detect() ([]types.Endpoint, error) {
	var endpoints []types.Endpoint

	// Detect based on framework
	endpoints = append(endpoints, d.detectExpressEndpoints()...)
	endpoints = append(endpoints, d.detectFastifyEndpoints()...)
	endpoints = append(endpoints, d.detectNextJSEndpoints()...)
	endpoints = append(endpoints, d.detectFastAPIEndpoints()...)
	endpoints = append(endpoints, d.detectFlaskEndpoints()...)
	endpoints = append(endpoints, d.detectDjangoEndpoints()...)
	endpoints = append(endpoints, d.detectSpringEndpoints()...)
	endpoints = append(endpoints, d.detectGinEndpoints()...)
	endpoints = append(endpoints, d.detectEchoEndpoints()...)
	endpoints = append(endpoints, d.detectFiberEndpoints()...)
	endpoints = append(endpoints, d.detectChiEndpoints()...)

	// Sort endpoints by path
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Path == endpoints[j].Path {
			return methodOrder(endpoints[i].Method) < methodOrder(endpoints[j].Method)
		}
		return endpoints[i].Path < endpoints[j].Path
	})

	return endpoints, nil
}

func methodOrder(method string) int {
	order := map[string]int{"GET": 1, "POST": 2, "PUT": 3, "PATCH": 4, "DELETE": 5}
	if o, ok := order[method]; ok {
		return o
	}
	return 99
}

// detectExpressEndpoints detects Express.js endpoints
func (d *EndpointDetector) detectExpressEndpoints() []types.Endpoint {
	var endpoints []types.Endpoint

	// Check if Express is used
	if !d.hasPackage("express") {
		return endpoints
	}

	// Patterns for Express routes
	// app.get('/path', handler)
	// router.get('/path', handler)
	routeRegex := regexp.MustCompile(`(?:app|router)\.(get|post|put|patch|delete)\s*\(\s*['"\x60]([^'"\x60]+)['"\x60]`)

	// Check for auth middleware
	authMiddlewareRegex := regexp.MustCompile(`(?:authenticate|auth|protect|requireAuth|isAuthenticated|verifyToken|jwt)`)

	for _, f := range d.files {
		if f.IsDir {
			continue
		}
		if f.Extension != ".js" && f.Extension != ".ts" {
			continue
		}
		// Skip test files and node_modules
		if strings.Contains(f.Path, "node_modules") || strings.Contains(f.Path, ".test.") || strings.Contains(f.Path, ".spec.") {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		for lineNum, line := range lines {
			matches := routeRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					method := strings.ToUpper(match[1])
					path := match[2]

					// Check if line has auth middleware
					auth := ""
					if authMiddlewareRegex.MatchString(line) {
						auth = "Required"
					}

					endpoints = append(endpoints, types.Endpoint{
						Method:  method,
						Path:    path,
						File:    f.Path,
						Line:    lineNum + 1,
						Auth:    auth,
						Handler: extractHandlerName(line),
					})
				}
			}
		}
	}

	return endpoints
}

// detectFastifyEndpoints detects Fastify endpoints
func (d *EndpointDetector) detectFastifyEndpoints() []types.Endpoint {
	var endpoints []types.Endpoint

	if !d.hasPackage("fastify") {
		return endpoints
	}

	routeRegex := regexp.MustCompile(`(?:fastify|app|server)\.(get|post|put|patch|delete)\s*\(\s*['"\x60]([^'"\x60]+)['"\x60]`)

	for _, f := range d.files {
		if f.IsDir || (f.Extension != ".js" && f.Extension != ".ts") {
			continue
		}
		if strings.Contains(f.Path, "node_modules") {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		for lineNum, line := range lines {
			matches := routeRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					endpoints = append(endpoints, types.Endpoint{
						Method: strings.ToUpper(match[1]),
						Path:   match[2],
						File:   f.Path,
						Line:   lineNum + 1,
					})
				}
			}
		}
	}

	return endpoints
}

// detectNextJSEndpoints detects Next.js API routes
func (d *EndpointDetector) detectNextJSEndpoints() []types.Endpoint {
	var endpoints []types.Endpoint

	if !d.hasPackage("next") {
		return endpoints
	}

	for _, f := range d.files {
		if f.IsDir {
			continue
		}

		// Pages Router: pages/api/**/*.ts
		if strings.Contains(f.Path, "pages/api/") || strings.Contains(f.Path, "src/pages/api/") {
			path := extractNextPagesAPIPath(f.Path)
			endpoints = append(endpoints, types.Endpoint{
				Method: "ALL",
				Path:   path,
				File:   f.Path,
			})
		}

		// App Router: app/**/route.ts
		if (f.Name == "route.ts" || f.Name == "route.js") &&
			(strings.Contains(f.Path, "/app/") || strings.Contains(f.Path, "src/app/")) {
			path := extractNextAppAPIPath(f.Path)
			methods := d.detectNextAppRouterMethods(filepath.Join(d.rootPath, f.Path))
			for _, method := range methods {
				endpoints = append(endpoints, types.Endpoint{
					Method: method,
					Path:   path,
					File:   f.Path,
				})
			}
		}
	}

	return endpoints
}

func extractNextPagesAPIPath(filePath string) string {
	// pages/api/users/[id].ts -> /api/users/[id]
	path := filePath
	path = strings.TrimPrefix(path, "src/")
	path = strings.TrimPrefix(path, "pages")
	path = strings.TrimSuffix(path, ".ts")
	path = strings.TrimSuffix(path, ".js")
	path = strings.TrimSuffix(path, "/index")
	if path == "" {
		path = "/"
	}
	return path
}

func extractNextAppAPIPath(filePath string) string {
	// app/api/users/[id]/route.ts -> /api/users/[id]
	path := filePath
	path = strings.TrimPrefix(path, "src/")
	path = strings.TrimPrefix(path, "app")
	path = strings.TrimSuffix(path, "/route.ts")
	path = strings.TrimSuffix(path, "/route.js")
	if path == "" {
		path = "/"
	}
	return path
}

func (d *EndpointDetector) detectNextAppRouterMethods(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{"ALL"}
	}

	contentStr := string(content)
	var methods []string

	methodPatterns := map[string]*regexp.Regexp{
		"GET":    regexp.MustCompile(`export\s+(?:async\s+)?function\s+GET`),
		"POST":   regexp.MustCompile(`export\s+(?:async\s+)?function\s+POST`),
		"PUT":    regexp.MustCompile(`export\s+(?:async\s+)?function\s+PUT`),
		"PATCH":  regexp.MustCompile(`export\s+(?:async\s+)?function\s+PATCH`),
		"DELETE": regexp.MustCompile(`export\s+(?:async\s+)?function\s+DELETE`),
	}

	for method, pattern := range methodPatterns {
		if pattern.MatchString(contentStr) {
			methods = append(methods, method)
		}
	}

	if len(methods) == 0 {
		return []string{"ALL"}
	}

	sort.Strings(methods)
	return methods
}

// detectFastAPIEndpoints detects FastAPI endpoints
func (d *EndpointDetector) detectFastAPIEndpoints() []types.Endpoint {
	var endpoints []types.Endpoint

	if !d.hasPythonPackage("fastapi") {
		return endpoints
	}

	// @app.get("/path") or @router.get("/path")
	routeRegex := regexp.MustCompile(`@(?:app|router)\.(get|post|put|patch|delete)\s*\(\s*["']([^"']+)["']`)

	for _, f := range d.files {
		if f.IsDir || f.Extension != ".py" {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		for lineNum, line := range lines {
			matches := routeRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					endpoints = append(endpoints, types.Endpoint{
						Method: strings.ToUpper(match[1]),
						Path:   match[2],
						File:   f.Path,
						Line:   lineNum + 1,
					})
				}
			}
		}
	}

	return endpoints
}

// detectFlaskEndpoints detects Flask endpoints
func (d *EndpointDetector) detectFlaskEndpoints() []types.Endpoint {
	var endpoints []types.Endpoint

	if !d.hasPythonPackage("flask") {
		return endpoints
	}

	// @app.route("/path", methods=["GET", "POST"])
	routeRegex := regexp.MustCompile(`@(?:app|blueprint|bp)\.route\s*\(\s*["']([^"']+)["'](?:.*methods\s*=\s*\[([^\]]+)\])?`)

	for _, f := range d.files {
		if f.IsDir || f.Extension != ".py" {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		for lineNum, line := range lines {
			matches := routeRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 2 {
					path := match[1]
					methods := []string{"GET"} // Default

					if len(match) >= 3 && match[2] != "" {
						// Parse methods from ["GET", "POST"]
						methodStr := strings.ReplaceAll(match[2], "'", "")
						methodStr = strings.ReplaceAll(methodStr, "\"", "")
						methodStr = strings.ReplaceAll(methodStr, " ", "")
						methods = strings.Split(methodStr, ",")
					}

					for _, method := range methods {
						endpoints = append(endpoints, types.Endpoint{
							Method: strings.ToUpper(strings.TrimSpace(method)),
							Path:   path,
							File:   f.Path,
							Line:   lineNum + 1,
						})
					}
				}
			}
		}
	}

	return endpoints
}

// detectDjangoEndpoints detects Django/DRF endpoints
func (d *EndpointDetector) detectDjangoEndpoints() []types.Endpoint {
	var endpoints []types.Endpoint

	if !d.hasPythonPackage("django") {
		return endpoints
	}

	// Look for urls.py files
	urlPathRegex := regexp.MustCompile(`path\s*\(\s*["']([^"']+)["']`)

	for _, f := range d.files {
		if f.IsDir || f.Name != "urls.py" {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		for lineNum, line := range lines {
			matches := urlPathRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 2 {
					path := "/" + match[1]
					endpoints = append(endpoints, types.Endpoint{
						Method: "ALL",
						Path:   path,
						File:   f.Path,
						Line:   lineNum + 1,
					})
				}
			}
		}
	}

	return endpoints
}

// detectSpringEndpoints detects Spring Boot endpoints
func (d *EndpointDetector) detectSpringEndpoints() []types.Endpoint {
	var endpoints []types.Endpoint

	if !d.hasSpring() {
		return endpoints
	}

	// @GetMapping("/path"), @PostMapping, etc.
	mappingRegex := regexp.MustCompile(`@(Get|Post|Put|Patch|Delete)Mapping\s*\(\s*(?:value\s*=\s*)?["']?([^"'\)]+)["']?\s*\)`)
	requestMappingRegex := regexp.MustCompile(`@RequestMapping\s*\(\s*(?:value\s*=\s*)?["']([^"']+)["']`)
	classRequestMapping := regexp.MustCompile(`@RequestMapping\s*\(\s*["']([^"']+)["']\s*\)`)

	for _, f := range d.files {
		if f.IsDir || (f.Extension != ".java" && f.Extension != ".kt") {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		// Find class-level @RequestMapping
		basePath := ""
		for _, line := range lines {
			if match := classRequestMapping.FindStringSubmatch(line); len(match) >= 2 {
				basePath = match[1]
				break
			}
		}

		for lineNum, line := range lines {
			// Check @GetMapping, @PostMapping, etc.
			matches := mappingRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					method := strings.ToUpper(match[1])
					path := basePath + match[2]
					if !strings.HasPrefix(path, "/") {
						path = "/" + path
					}
					endpoints = append(endpoints, types.Endpoint{
						Method: method,
						Path:   path,
						File:   f.Path,
						Line:   lineNum + 1,
					})
				}
			}

			// Check @RequestMapping (method-level)
			if strings.Contains(line, "@RequestMapping") && !strings.Contains(line, "class") {
				if match := requestMappingRegex.FindStringSubmatch(line); len(match) >= 2 {
					path := basePath + match[1]
					if !strings.HasPrefix(path, "/") {
						path = "/" + path
					}
					endpoints = append(endpoints, types.Endpoint{
						Method: "ALL",
						Path:   path,
						File:   f.Path,
						Line:   lineNum + 1,
					})
				}
			}
		}
	}

	return endpoints
}

// detectGinEndpoints detects Gin (Go) endpoints
func (d *EndpointDetector) detectGinEndpoints() []types.Endpoint {
	var endpoints []types.Endpoint

	if !d.hasGoPackage("github.com/gin-gonic/gin") {
		return endpoints
	}

	// r.GET("/path", handler) or group.GET("/path", handler)
	routeRegex := regexp.MustCompile(`\.(GET|POST|PUT|PATCH|DELETE)\s*\(\s*["']([^"']+)["']`)

	for _, f := range d.files {
		if f.IsDir || f.Extension != ".go" {
			continue
		}
		if strings.HasSuffix(f.Name, "_test.go") {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "gin") {
			continue
		}

		lines := strings.Split(contentStr, "\n")

		for lineNum, line := range lines {
			matches := routeRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					endpoints = append(endpoints, types.Endpoint{
						Method: match[1],
						Path:   match[2],
						File:   f.Path,
						Line:   lineNum + 1,
					})
				}
			}
		}
	}

	return endpoints
}

// detectEchoEndpoints detects Echo (Go) endpoints
func (d *EndpointDetector) detectEchoEndpoints() []types.Endpoint {
	var endpoints []types.Endpoint

	if !d.hasGoPackage("github.com/labstack/echo") {
		return endpoints
	}

	routeRegex := regexp.MustCompile(`e\.(GET|POST|PUT|PATCH|DELETE)\s*\(\s*["']([^"']+)["']`)

	for _, f := range d.files {
		if f.IsDir || f.Extension != ".go" || strings.HasSuffix(f.Name, "_test.go") {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		for lineNum, line := range lines {
			matches := routeRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					endpoints = append(endpoints, types.Endpoint{
						Method: match[1],
						Path:   match[2],
						File:   f.Path,
						Line:   lineNum + 1,
					})
				}
			}
		}
	}

	return endpoints
}

// detectFiberEndpoints detects Fiber (Go) endpoints
func (d *EndpointDetector) detectFiberEndpoints() []types.Endpoint {
	var endpoints []types.Endpoint

	if !d.hasGoPackage("github.com/gofiber/fiber") {
		return endpoints
	}

	routeRegex := regexp.MustCompile(`\.(Get|Post|Put|Patch|Delete)\s*\(\s*["']([^"']+)["']`)

	for _, f := range d.files {
		if f.IsDir || f.Extension != ".go" || strings.HasSuffix(f.Name, "_test.go") {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		for lineNum, line := range lines {
			matches := routeRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					endpoints = append(endpoints, types.Endpoint{
						Method: strings.ToUpper(match[1]),
						Path:   match[2],
						File:   f.Path,
						Line:   lineNum + 1,
					})
				}
			}
		}
	}

	return endpoints
}

// detectChiEndpoints detects Chi (Go) endpoints
func (d *EndpointDetector) detectChiEndpoints() []types.Endpoint {
	var endpoints []types.Endpoint

	if !d.hasGoPackage("github.com/go-chi/chi") {
		return endpoints
	}

	routeRegex := regexp.MustCompile(`r\.(Get|Post|Put|Patch|Delete)\s*\(\s*["']([^"']+)["']`)

	for _, f := range d.files {
		if f.IsDir || f.Extension != ".go" || strings.HasSuffix(f.Name, "_test.go") {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		for lineNum, line := range lines {
			matches := routeRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					endpoints = append(endpoints, types.Endpoint{
						Method: strings.ToUpper(match[1]),
						Path:   match[2],
						File:   f.Path,
						Line:   lineNum + 1,
					})
				}
			}
		}
	}

	return endpoints
}

// Helper functions

func (d *EndpointDetector) hasPackage(name string) bool {
	pkgPath := filepath.Join(d.rootPath, "package.json")
	content, err := os.ReadFile(pkgPath)
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(content)), "\""+name+"\"")
}

func (d *EndpointDetector) hasPythonPackage(name string) bool {
	// Check requirements.txt
	reqPath := filepath.Join(d.rootPath, "requirements.txt")
	if content, err := os.ReadFile(reqPath); err == nil {
		if strings.Contains(strings.ToLower(string(content)), name) {
			return true
		}
	}

	// Check pyproject.toml
	pyprojectPath := filepath.Join(d.rootPath, "pyproject.toml")
	if content, err := os.ReadFile(pyprojectPath); err == nil {
		if strings.Contains(strings.ToLower(string(content)), name) {
			return true
		}
	}

	return false
}

func (d *EndpointDetector) hasGoPackage(name string) bool {
	modPath := filepath.Join(d.rootPath, "go.mod")
	content, err := os.ReadFile(modPath)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), name)
}

func (d *EndpointDetector) hasSpring() bool {
	// Check pom.xml
	pomPath := filepath.Join(d.rootPath, "pom.xml")
	if content, err := os.ReadFile(pomPath); err == nil {
		if strings.Contains(string(content), "spring-boot") || strings.Contains(string(content), "springframework") {
			return true
		}
	}

	// Check build.gradle
	gradlePath := filepath.Join(d.rootPath, "build.gradle")
	if content, err := os.ReadFile(gradlePath); err == nil {
		if strings.Contains(string(content), "spring-boot") || strings.Contains(string(content), "springframework") {
			return true
		}
	}

	return false
}

func extractHandlerName(line string) string {
	// Try to extract handler function name from the line
	// app.get('/path', handlerName)
	// app.get('/path', (req, res) => { ... })
	handlerRegex := regexp.MustCompile(`,\s*(\w+)\s*\)`)
	if match := handlerRegex.FindStringSubmatch(line); len(match) >= 2 {
		return match[1]
	}
	return ""
}
