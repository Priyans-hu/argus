package detector

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// CodePatternDetector performs deep analysis of code for framework-specific patterns
type CodePatternDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewCodePatternDetector creates a new code pattern detector
func NewCodePatternDetector(rootPath string, files []types.FileInfo) *CodePatternDetector {
	return &CodePatternDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// PatternCategory represents a category of detected patterns
type PatternCategory struct {
	Name        string
	Patterns    []DetectedPattern
	Description string
}

// DetectedPattern represents a pattern found in the codebase
type DetectedPattern struct {
	Name        string
	Count       int
	Examples    []string // File paths where found
	Description string
	Usage       string // How to use this pattern
}

// Detect performs deep code analysis and returns detected patterns
func (d *CodePatternDetector) Detect() *types.CodePatterns {
	patterns := &types.CodePatterns{
		StateManagement: d.detectStateManagement(),
		DataFetching:    d.detectDataFetching(),
		Routing:         d.detectRouting(),
		Forms:           d.detectForms(),
		Testing:         d.detectTesting(),
		Styling:         d.detectStyling(),
		Authentication:  d.detectAuthentication(),
		APIPatterns:     d.detectAPIPatterns(),
		DatabaseORM:     d.detectDatabasePatterns(),
		Utilities:       d.detectUtilityPatterns(),
	}

	return patterns
}

// KeywordMatch holds information about a keyword match
type keywordMatch struct {
	keyword string
	file    string
	line    int
}

// scanForKeywords scans files for specific keywords
func (d *CodePatternDetector) scanForKeywords(keywords []string, extensions []string) map[string][]string {
	results := make(map[string][]string)

	for _, f := range d.files {
		if f.IsDir {
			continue
		}

		// Check file extension
		hasExt := false
		for _, ext := range extensions {
			if strings.HasSuffix(f.Name, ext) {
				hasExt = true
				break
			}
		}
		if !hasExt {
			continue
		}

		// Read file content
		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		contentStr := string(content)
		for _, kw := range keywords {
			if strings.Contains(contentStr, kw) {
				results[kw] = append(results[kw], f.Path)
			}
		}
	}

	return results
}

// scanForRegex scans files for regex patterns
func (d *CodePatternDetector) scanForRegex(pattern string, extensions []string) []string {
	var matches []string
	re := regexp.MustCompile(pattern)

	for _, f := range d.files {
		if f.IsDir {
			continue
		}

		hasExt := false
		for _, ext := range extensions {
			if strings.HasSuffix(f.Name, ext) {
				hasExt = true
				break
			}
		}
		if !hasExt {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		if re.Match(content) {
			matches = append(matches, f.Path)
		}
	}

	return matches
}

// detectStateManagement detects state management patterns
func (d *CodePatternDetector) detectStateManagement() []types.PatternInfo {
	var patterns []types.PatternInfo
	jsExts := []string{".js", ".jsx", ".ts", ".tsx"}
	vueExts := []string{".vue", ".js", ".ts"}
	pyExts := []string{".py"}

	// React state management
	reactKeywords := map[string]string{
		"useState":       "React useState hook for local component state",
		"useReducer":     "React useReducer for complex state logic",
		"useContext":     "React Context API for prop drilling avoidance",
		"createContext":  "React Context creation",
		"zustand":        "Zustand - lightweight state management",
		"create(":        "Zustand store creation",
		"useStore":       "Zustand/generic store hook",
		"jotai":          "Jotai - atomic state management",
		"atom(":          "Jotai/Recoil atom definition",
		"useAtom":        "Jotai atom hook",
		"recoil":         "Recoil state management",
		"useRecoilState": "Recoil state hook",
		"@reduxjs/toolkit": "Redux Toolkit",
		"createSlice":    "Redux Toolkit slice",
		"useSelector":    "Redux selector hook",
		"useDispatch":    "Redux dispatch hook",
		"configureStore": "Redux store configuration",
	}

	jsResults := d.scanForKeywords(keys(reactKeywords), jsExts)
	for kw, files := range jsResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "state",
				Description: reactKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Vue state management
	vueKeywords := map[string]string{
		"defineStore":  "Pinia store definition",
		"storeToRefs":  "Pinia reactive store refs",
		"createPinia":  "Pinia initialization",
		"ref(":         "Vue ref for reactive primitives",
		"reactive(":    "Vue reactive for objects",
		"computed(":    "Vue computed properties",
		"watch(":       "Vue watch for side effects",
		"watchEffect(": "Vue watchEffect for automatic tracking",
	}

	vueResults := d.scanForKeywords(keys(vueKeywords), vueExts)
	for kw, files := range vueResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "state",
				Description: vueKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Python state patterns
	pyKeywords := map[string]string{
		"session[":     "Flask session management",
		"g.":           "Flask application context globals",
		"current_app":  "Flask current application context",
		"request.":     "Flask/FastAPI request object",
	}

	pyResults := d.scanForKeywords(keys(pyKeywords), pyExts)
	for kw, files := range pyResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "state",
				Description: pyKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
}

// detectDataFetching detects data fetching patterns
func (d *CodePatternDetector) detectDataFetching() []types.PatternInfo {
	var patterns []types.PatternInfo
	jsExts := []string{".js", ".jsx", ".ts", ".tsx"}
	pyExts := []string{".py"}
	goExts := []string{".go"}

	// React/JS data fetching
	jsKeywords := map[string]string{
		"useQuery":        "TanStack Query (React Query) for server state",
		"useMutation":     "TanStack Query mutation hook",
		"useInfiniteQuery": "TanStack Query infinite scrolling",
		"QueryClient":     "TanStack Query client setup",
		"useSWR":          "SWR data fetching hook",
		"axios":           "Axios HTTP client",
		"axios.get":       "Axios GET request",
		"axios.post":      "Axios POST request",
		"fetch(":          "Native Fetch API",
		"$fetch":          "Nuxt/ofetch utility",
		"useFetch":        "Nuxt/custom fetch hook",
		"getServerSideProps": "Next.js server-side data fetching",
		"getStaticProps":  "Next.js static data fetching",
		"useLoaderData":   "Remix loader data hook",
		"trpc":            "tRPC type-safe API calls",
	}

	jsResults := d.scanForKeywords(keys(jsKeywords), jsExts)
	for kw, files := range jsResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "data-fetching",
				Description: jsKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Python data fetching
	pyKeywords := map[string]string{
		"requests.":   "Python requests library",
		"httpx.":      "HTTPX async HTTP client",
		"aiohttp.":    "aiohttp async HTTP client",
		"urllib":      "Python urllib",
	}

	pyResults := d.scanForKeywords(keys(pyKeywords), pyExts)
	for kw, files := range pyResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "data-fetching",
				Description: pyKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Go HTTP clients
	goKeywords := map[string]string{
		"http.Get":      "Go standard HTTP GET",
		"http.Post":     "Go standard HTTP POST",
		"http.Client":   "Go HTTP client",
		"resty.":        "Resty HTTP client",
	}

	goResults := d.scanForKeywords(keys(goKeywords), goExts)
	for kw, files := range goResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "data-fetching",
				Description: goKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
}

// detectRouting detects routing patterns
func (d *CodePatternDetector) detectRouting() []types.PatternInfo {
	var patterns []types.PatternInfo
	jsExts := []string{".js", ".jsx", ".ts", ".tsx"}
	pyExts := []string{".py"}
	goExts := []string{".go"}

	// React routing
	jsKeywords := map[string]string{
		"useRouter":      "Next.js/custom router hook",
		"useNavigate":    "React Router navigation hook",
		"useParams":      "React Router URL params",
		"useSearchParams": "React Router/Next.js search params",
		"<Link":          "Router Link component",
		"<Route":         "React Router Route component",
		"createBrowserRouter": "React Router v6 browser router",
		"next/navigation": "Next.js App Router navigation",
		"next/link":      "Next.js Link component",
		"usePathname":    "Next.js pathname hook",
	}

	jsResults := d.scanForKeywords(keys(jsKeywords), jsExts)
	for kw, files := range jsResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "routing",
				Description: jsKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Python routing
	pyKeywords := map[string]string{
		"@app.route":      "Flask route decorator",
		"@router.":        "FastAPI router decorator",
		"@app.get":        "FastAPI GET endpoint",
		"@app.post":       "FastAPI POST endpoint",
		"Blueprint":       "Flask Blueprint for modular routing",
		"APIRouter":       "FastAPI APIRouter",
		"path(":           "Django URL path",
		"include(":        "Django URL includes",
	}

	pyResults := d.scanForKeywords(keys(pyKeywords), pyExts)
	for kw, files := range pyResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "routing",
				Description: pyKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Go routing - exclude test files to avoid false positives
	goKeywords := map[string]string{
		"gin.Context":     "Gin framework context",
		"gin.Default":     "Gin default router",
		"echo.Context":    "Echo framework context",
		"echo.New":        "Echo router initialization",
		"fiber.Ctx":       "Fiber framework context",
		"fiber.New":       "Fiber app initialization",
		"chi.Router":      "Chi router",
		"chi.NewRouter":   "Chi router initialization",
		"mux.NewRouter":   "Gorilla Mux router",
		"http.HandleFunc": "Go standard HTTP handler",
		"http.Handle":     "Go standard HTTP handler",
	}

	goResults := d.scanForKeywords(keys(goKeywords), goExts)
	for kw, files := range goResults {
		// Filter out test files for routing patterns
		nonTestFiles := filterOutTestFiles(files)
		if len(nonTestFiles) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "routing",
				Description: goKeywords[kw],
				FileCount:   len(nonTestFiles),
				Examples:    limitSlice(nonTestFiles, 3),
			})
		}
	}

	return patterns
}

// filterOutTestFiles removes test files from the list
func filterOutTestFiles(files []string) []string {
	var result []string
	for _, f := range files {
		if !strings.HasSuffix(f, "_test.go") &&
			!strings.Contains(f, ".test.") &&
			!strings.Contains(f, ".spec.") &&
			!strings.Contains(f, "__tests__") {
			result = append(result, f)
		}
	}
	return result
}

// detectForms detects form handling patterns
func (d *CodePatternDetector) detectForms() []types.PatternInfo {
	var patterns []types.PatternInfo
	jsExts := []string{".js", ".jsx", ".ts", ".tsx"}

	jsKeywords := map[string]string{
		"useForm":         "React Hook Form / TanStack Form",
		"useFormContext":  "React Hook Form context",
		"zodResolver":     "Zod schema validation with forms",
		"yupResolver":     "Yup schema validation with forms",
		"Formik":          "Formik form library",
		"useFormik":       "Formik hook",
		"<form":           "HTML form element",
		"onSubmit":        "Form submission handler",
		"handleSubmit":    "Form submit handler pattern",
		"register(":       "React Hook Form field registration",
		"Controller":      "React Hook Form Controller",
		"z.object":        "Zod object schema",
		"z.string":        "Zod string validation",
	}

	jsResults := d.scanForKeywords(keys(jsKeywords), jsExts)
	for kw, files := range jsResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "forms",
				Description: jsKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
}

// detectTesting detects testing patterns
func (d *CodePatternDetector) detectTesting() []types.PatternInfo {
	var patterns []types.PatternInfo
	jsExts := []string{".js", ".jsx", ".ts", ".tsx", ".test.js", ".test.ts", ".spec.js", ".spec.ts"}
	pyExts := []string{".py"}
	goTestExts := []string{"_test.go"} // Only scan test files for Go testing patterns
	goExts := []string{".go"}

	// JS testing
	jsKeywords := map[string]string{
		"describe(":        "Test suite definition (Jest/Vitest/Mocha)",
		"it(":              "Test case (Jest/Vitest/Mocha)",
		"test(":            "Test case (Jest/Vitest)",
		"expect(":          "Assertion (Jest/Vitest/Chai)",
		"vi.mock":          "Vitest mocking",
		"jest.mock":        "Jest mocking",
		"@testing-library": "Testing Library",
		"render(":          "Testing Library render",
		"screen.":          "Testing Library screen queries",
		"userEvent":        "Testing Library user events",
		"fireEvent":        "Testing Library fire events",
		"cy.":              "Cypress commands",
		"playwright":       "Playwright testing",
	}

	jsResults := d.scanForKeywords(keys(jsKeywords), jsExts)
	for kw, files := range jsResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "testing",
				Description: jsKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Python testing
	pyKeywords := map[string]string{
		"pytest":          "Pytest framework",
		"def test_":       "Pytest test function",
		"unittest":        "Python unittest",
		"TestCase":        "unittest TestCase class",
		"@pytest.fixture": "Pytest fixture",
		"mock.":           "Python mocking",
		"@patch":          "unittest mock patch",
	}

	pyResults := d.scanForKeywords(keys(pyKeywords), pyExts)
	for kw, files := range pyResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "testing",
				Description: pyKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Go testing - scan only _test.go files, consolidate similar patterns
	goTestKeywords := []string{"func Test", "t.Run(", "t.Error", "t.Fatal"}
	goTestResults := d.scanForKeywords(goTestKeywords, goTestExts)

	// Consolidate t.Error and t.Errorf, t.Fatal and t.Fatalf
	consolidatedResults := make(map[string][]string)
	for kw, files := range goTestResults {
		// Normalize key names
		normalizedKey := kw
		if strings.HasPrefix(kw, "t.Error") {
			normalizedKey = "t.Error"
		} else if strings.HasPrefix(kw, "t.Fatal") {
			normalizedKey = "t.Fatal"
		}
		// Merge files, dedupe
		existing := consolidatedResults[normalizedKey]
		for _, f := range files {
			found := false
			for _, e := range existing {
				if e == f {
					found = true
					break
				}
			}
			if !found {
				existing = append(existing, f)
			}
		}
		consolidatedResults[normalizedKey] = existing
	}

	goTestDescriptions := map[string]string{
		"func Test": "Go test function",
		"t.Run(":    "Go subtest",
		"t.Error":   "Go test assertions",
		"t.Fatal":   "Go test fatal assertions",
	}

	for kw, files := range consolidatedResults {
		if len(files) > 0 {
			displayName := strings.TrimSuffix(kw, "(")
			desc := goTestDescriptions[kw]
			if desc == "" {
				desc = "Go testing"
			}
			patterns = append(patterns, types.PatternInfo{
				Name:        displayName,
				Category:    "testing",
				Description: desc,
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Go testing helpers - can be in any .go file
	goHelperKeywords := map[string]string{
		"require.":  "Testify require assertions",
		"assert.":   "Testify assert",
		"gomock":    "GoMock mocking",
		"httptest.": "Go HTTP testing",
	}

	goHelperResults := d.scanForKeywords(keys(goHelperKeywords), goExts)
	for kw, files := range goHelperResults {
		if len(files) > 0 {
			displayName := strings.TrimSuffix(kw, ".")
			patterns = append(patterns, types.PatternInfo{
				Name:        displayName,
				Category:    "testing",
				Description: goHelperKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
}

// detectStyling detects styling patterns
func (d *CodePatternDetector) detectStyling() []types.PatternInfo {
	var patterns []types.PatternInfo
	jsExts := []string{".js", ".jsx", ".ts", ".tsx", ".css", ".scss"}

	jsKeywords := map[string]string{
		"className=":     "CSS class usage",
		"tailwind":       "Tailwind CSS",
		"@apply":         "Tailwind @apply directive",
		"styled.":        "styled-components",
		"css`":           "Emotion/styled-components CSS",
		"makeStyles":     "Material-UI makeStyles",
		"useStyles":      "Material-UI styles hook",
		"sx={":           "MUI sx prop",
		"clsx":           "clsx class utility",
		"cn(":            "shadcn/ui cn utility",
		"cva(":           "Class Variance Authority",
		"module.css":     "CSS Modules",
	}

	jsResults := d.scanForKeywords(keys(jsKeywords), jsExts)
	for kw, files := range jsResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "styling",
				Description: jsKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
}

// detectAuthentication detects auth patterns
func (d *CodePatternDetector) detectAuthentication() []types.PatternInfo {
	var patterns []types.PatternInfo
	jsExts := []string{".js", ".jsx", ".ts", ".tsx"}
	pyExts := []string{".py"}
	goExts := []string{".go"}

	// JS auth
	jsKeywords := map[string]string{
		"useAuth":        "Custom auth hook",
		"useSession":     "NextAuth/custom session hook",
		"signIn":         "Auth sign in function",
		"signOut":        "Auth sign out function",
		"getSession":     "NextAuth getSession",
		"NextAuth":       "NextAuth.js",
		"Auth0":          "Auth0 integration",
		"clerk":          "Clerk authentication",
		"supabase.auth":  "Supabase authentication",
		"firebase.auth":  "Firebase authentication",
		"jwt":            "JWT handling",
		"Bearer":         "Bearer token auth",
	}

	jsResults := d.scanForKeywords(keys(jsKeywords), jsExts)
	for kw, files := range jsResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "auth",
				Description: jsKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Python auth
	pyKeywords := map[string]string{
		"login_required":  "Flask login decorator",
		"current_user":    "Flask-Login current user",
		"@jwt_required":   "JWT required decorator",
		"OAuth":           "OAuth integration",
		"HTTPBearer":      "FastAPI HTTP Bearer auth",
		"Depends(":        "FastAPI dependency injection",
	}

	pyResults := d.scanForKeywords(keys(pyKeywords), pyExts)
	for kw, files := range pyResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "auth",
				Description: pyKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Go auth
	goKeywords := map[string]string{
		"jwt.":          "JWT handling",
		"middleware":    "Auth middleware",
		"Authorization": "Authorization header",
		"Bearer":        "Bearer token",
	}

	goResults := d.scanForKeywords(keys(goKeywords), goExts)
	for kw, files := range goResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "auth",
				Description: goKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
}

// detectAPIPatterns detects API design patterns
func (d *CodePatternDetector) detectAPIPatterns() []types.PatternInfo {
	var patterns []types.PatternInfo
	allExts := []string{".js", ".jsx", ".ts", ".tsx", ".py", ".go"}

	keywords := map[string]string{
		"REST":           "RESTful API design",
		"GraphQL":        "GraphQL API",
		"gql`":           "GraphQL query",
		"useQuery":       "GraphQL/React Query",
		"useMutation":    "GraphQL/React Query mutation",
		"tRPC":           "tRPC type-safe API",
		"OpenAPI":        "OpenAPI/Swagger spec",
		"swagger":        "Swagger documentation",
		"grpc":           "gRPC protocol",
		"protobuf":       "Protocol Buffers",
		"websocket":      "WebSocket communication",
		"socket.io":      "Socket.IO real-time",
	}

	results := d.scanForKeywords(keys(keywords), allExts)
	for kw, files := range results {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "api",
				Description: keywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
}

// detectDatabasePatterns detects database/ORM patterns
func (d *CodePatternDetector) detectDatabasePatterns() []types.PatternInfo {
	var patterns []types.PatternInfo
	jsExts := []string{".js", ".ts", ".tsx"}
	pyExts := []string{".py"}
	goExts := []string{".go"}

	// JS/TS ORMs
	jsKeywords := map[string]string{
		"prisma":         "Prisma ORM",
		"PrismaClient":   "Prisma client",
		"drizzle":        "Drizzle ORM",
		"typeorm":        "TypeORM",
		"@Entity":        "TypeORM entity decorator",
		"sequelize":      "Sequelize ORM",
		"mongoose":       "Mongoose ODM (MongoDB)",
		"knex":           "Knex.js query builder",
		"kysely":         "Kysely type-safe SQL",
		"supabase":       "Supabase client",
	}

	jsResults := d.scanForKeywords(keys(jsKeywords), jsExts)
	for kw, files := range jsResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "database",
				Description: jsKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Python ORMs
	pyKeywords := map[string]string{
		"SQLAlchemy":     "SQLAlchemy ORM",
		"Base.metadata":  "SQLAlchemy models",
		"django.db":      "Django ORM",
		"models.Model":   "Django model",
		"tortoise":       "Tortoise ORM",
		"peewee":         "Peewee ORM",
		"mongoengine":    "MongoEngine ODM",
		"motor":          "Motor async MongoDB",
	}

	pyResults := d.scanForKeywords(keys(pyKeywords), pyExts)
	for kw, files := range pyResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "database",
				Description: pyKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Go ORMs - use more specific patterns to avoid false positives
	goKeywords := map[string]string{
		"gorm.Open(":     "GORM ORM",
		"gorm.Model":     "GORM model embedding",
		"sqlx.Connect":   "sqlx database library",
		"sqlx.Open":      "sqlx database library",
		"sql.Open(":      "Go standard SQL",
		"pgx.Connect":    "pgx PostgreSQL driver",
		"mongo.Connect":  "MongoDB Go driver",
		"bun.NewDB":      "Bun ORM",
	}

	goResults := d.scanForKeywords(keys(goKeywords), goExts)
	for kw, files := range goResults {
		if len(files) > 0 {
			// Clean display name
			displayName := strings.Split(kw, "(")[0]
			displayName = strings.Split(displayName, ".")[0] + "." + strings.Split(displayName, ".")[1]
			patterns = append(patterns, types.PatternInfo{
				Name:        displayName,
				Category:    "database",
				Description: goKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Ent ORM - needs regex to avoid false positives (content, environment, etc.)
	entFiles := d.scanForRegex(`ent\.(Client|Schema|Field|Edge|Mixin)`, goExts)
	if len(entFiles) > 0 {
		patterns = append(patterns, types.PatternInfo{
			Name:        "ent",
			Category:    "database",
			Description: "Ent ORM (entgo.io)",
			FileCount:   len(entFiles),
			Examples:    limitSlice(entFiles, 3),
		})
	}

	return patterns
}

// detectUtilityPatterns detects utility/helper patterns
func (d *CodePatternDetector) detectUtilityPatterns() []types.PatternInfo {
	var patterns []types.PatternInfo
	jsExts := []string{".js", ".jsx", ".ts", ".tsx"}

	jsKeywords := map[string]string{
		"lodash":         "Lodash utility library",
		"dayjs":          "Day.js date library",
		"moment":         "Moment.js date library",
		"date-fns":       "date-fns date utilities",
		"uuid":           "UUID generation",
		"nanoid":         "Nano ID generation",
		"zod":            "Zod schema validation",
		"yup":            "Yup schema validation",
		"ajv":            "AJV JSON schema validation",
		"immer":          "Immer immutable updates",
		"ramda":          "Ramda FP utilities",
	}

	jsResults := d.scanForKeywords(keys(jsKeywords), jsExts)
	for kw, files := range jsResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "utilities",
				Description: jsKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
}

// countFileLines counts lines in a file (for size estimation)
func countFileLines(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count
}

// Helper functions
func keys(m map[string]string) []string {
	k := make([]string, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	return k
}

func limitSlice(s []string, max int) []string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
