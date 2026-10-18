package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Priyans-hu/argus/pkg/types"
)

// CodePatternDetector performs deep analysis of code for framework-specific patterns
type CodePatternDetector struct {
	rootPath string
	files    []types.FileInfo
	deps     DepSet
}

// NewCodePatternDetector creates a code pattern detector; library patterns
// need their package among deps.
func NewCodePatternDetector(rootPath string, files []types.FileInfo, deps []types.Dependency) *CodePatternDetector {
	return &CodePatternDetector{rootPath: rootPath, files: files, deps: NewDepSet(deps)}
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
		GoPatterns:      d.detectGoPatterns(),
		RustPatterns:    d.detectRustPatterns(),
		PythonPatterns:  d.detectPythonPatterns(),
	}

	return patterns
}

// keywordRow is a keyword that marks a pattern. A row naming packages is
// reported only when one of them is a declared dependency; a row without
// packages is intrinsic to the language or platform.
type keywordRow struct {
	keyword     string
	description string
	packages    []string
}

// keywordTable holds the rows of one category, scanned in the files with one
// of exts that include accepts.
type keywordTable struct {
	category string
	exts     []string
	include  func(rel string) bool
	rows     []keywordRow
}

// report returns, in row order, a pattern for every evidenced row whose
// keyword is found.
func (d *CodePatternDetector) report(t keywordTable) []types.PatternInfo {
	var keywords []string
	for _, r := range t.rows {
		if len(r.packages) == 0 || d.deps.declares(r.packages) {
			keywords = append(keywords, r.keyword)
		}
	}
	found := d.scanKeywords(keywords, t.exts, t.include)
	var patterns []types.PatternInfo
	for _, r := range t.rows {
		if files := found[r.keyword]; len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name: r.keyword, Category: t.category, Description: r.description,
				FileCount: len(files), Examples: limitSlice(files, 3),
			})
		}
	}
	return patterns
}

var (
	jsExts = []string{".js", ".jsx", ".ts", ".tsx"}

	reactPkgs      = []string{"react", "preact"}
	reduxPkgs      = []string{"@reduxjs/toolkit", "redux", "react-redux"}
	reactQueryPkgs = []string{"@tanstack/react-query", "react-query"}
	routerPkgs     = []string{"react-router", "react-router-dom", "@remix-run/react", "next"}
	nextPkgs       = []string{"next"}
	nextAuthPkgs   = []string{"next-auth"}
	authPkgs       = []string{"next-auth", "@auth/", "firebase", "@supabase/", "@clerk/", "aws-amplify"}
	hookFormPkgs   = []string{"react-hook-form"}
	testingLibPkgs = []string{"@testing-library/"}
	tailwindPkgs   = []string{"tailwindcss", "@tailwindcss/"}
	muiStylesPkgs  = []string{"@material-ui/core", "@mui/styles"}
	axiosPkgs      = []string{"axios"}
	prismaPkgs     = []string{"@prisma/client", "prisma"}
	graphqlPkgs    = []string{"graphql", "@apollo/client", "@apollo/server", "apollo-server", "graphql-request", "urql", "graphql-yoga",
		"graphene", "strawberry-graphql", "ariadne", "github.com/99designs/gqlgen", "github.com/graphql-go/graphql"}
)

var reactStateTable = keywordTable{"state", jsExts, nonTestFile, []keywordRow{
	{"useState", "React useState hook for local component state", reactPkgs},
	{"useReducer", "React useReducer for complex state logic", reactPkgs},
	{"useContext", "React Context API for prop drilling avoidance", reactPkgs},
	{"createContext", "React Context creation", reactPkgs},
	{"zustand", "Zustand - lightweight state management", []string{"zustand"}},
	{"create(", "Zustand store creation", []string{"zustand"}},
	{"useStore", "Zustand/generic store hook", []string{"zustand", "@nanostores/react", "@tanstack/react-store"}},
	{"jotai", "Jotai - atomic state management", []string{"jotai"}},
	{"atom(", "Jotai/Recoil atom definition", []string{"jotai", "recoil"}},
	{"useAtom", "Jotai atom hook", []string{"jotai"}},
	{"recoil", "Recoil state management", []string{"recoil"}},
	{"useRecoilState", "Recoil state hook", []string{"recoil"}},
	{"@reduxjs/toolkit", "Redux Toolkit", reduxPkgs},
	{"createSlice", "Redux Toolkit slice", reduxPkgs},
	{"useSelector", "Redux selector hook", reduxPkgs},
	{"useDispatch", "Redux dispatch hook", reduxPkgs},
	{"configureStore", "Redux store configuration", reduxPkgs},
}}

var vueStateTable = keywordTable{"state", []string{".vue", ".js", ".ts"}, nonTestFile, []keywordRow{
	{"defineStore", "Pinia store definition", []string{"pinia"}},
	{"storeToRefs", "Pinia reactive store refs", []string{"pinia"}},
	{"createPinia", "Pinia initialization", []string{"pinia"}},
	{"ref(", "Vue ref for reactive primitives", []string{"vue", "nuxt"}},
	{"reactive(", "Vue reactive for objects", []string{"vue", "nuxt"}},
	{"computed(", "Vue computed properties", []string{"vue", "nuxt"}},
	{"watch(", "Vue watch for side effects", []string{"vue", "nuxt"}},
	{"watchEffect(", "Vue watchEffect for automatic tracking", []string{"vue", "nuxt"}},
}}

var jsDataFetchingTable = keywordTable{"data-fetching", jsExts, nonTestFile, []keywordRow{
	{"useQuery", "TanStack Query (React Query) for server state", reactQueryPkgs},
	{"useMutation", "TanStack Query mutation hook", reactQueryPkgs},
	{"useInfiniteQuery", "TanStack Query infinite scrolling", reactQueryPkgs},
	{"QueryClient", "TanStack Query client setup", reactQueryPkgs},
	{"useSWR", "SWR data fetching hook", []string{"swr"}},
	{"axios", "Axios HTTP client", axiosPkgs},
	{"axios.get", "Axios GET request", axiosPkgs},
	{"axios.post", "Axios POST request", axiosPkgs},
	{"fetch(", "Native Fetch API", nil},
	{"$fetch", "Nuxt/ofetch utility", []string{"nuxt", "ofetch"}},
	{"useFetch", "Nuxt/custom fetch hook", []string{"nuxt", "@vueuse/core"}},
	{"getServerSideProps", "Next.js server-side data fetching", nextPkgs},
	{"getStaticProps", "Next.js static data fetching", nextPkgs},
	{"useLoaderData", "Remix loader data hook", []string{"@remix-run/react", "react-router", "react-router-dom"}},
	{"trpc", "tRPC type-safe API calls", []string{"@trpc/"}},
}}

var jsRoutingTable = keywordTable{"routing", jsExts, nonTestFile, []keywordRow{
	{"useRouter", "Next.js/custom router hook", nextPkgs},
	{"useNavigate", "React Router navigation hook", []string{"react-router", "react-router-dom", "@remix-run/react"}},
	{"useParams", "React Router URL params", routerPkgs},
	{"useSearchParams", "React Router/Next.js search params", routerPkgs},
	{"<Link", "Router Link component", routerPkgs},
	{"<Route", "React Router Route component", []string{"react-router", "react-router-dom"}},
	{"createBrowserRouter", "React Router v6 browser router", []string{"react-router", "react-router-dom"}},
	{"next/navigation", "Next.js App Router navigation", nextPkgs},
	{"next/link", "Next.js Link component", nextPkgs},
	{"usePathname", "Next.js pathname hook", nextPkgs},
}}

var jsFormsTable = keywordTable{"forms", jsExts, nonTestFile, []keywordRow{
	{"useForm", "React Hook Form / TanStack Form", []string{"react-hook-form", "@tanstack/react-form"}},
	{"useFormContext", "React Hook Form context", hookFormPkgs},
	{"zodResolver", "Zod schema validation with forms", []string{"@hookform/resolvers"}},
	{"yupResolver", "Yup schema validation with forms", []string{"@hookform/resolvers"}},
	{"Formik", "Formik form library", []string{"formik"}},
	{"useFormik", "Formik hook", []string{"formik"}},
	{"<form", "HTML form element", nil},
	{"register(", "React Hook Form field registration", hookFormPkgs},
	{"Controller", "React Hook Form Controller", hookFormPkgs},
	{"z.object", "Zod object schema", []string{"zod"}},
	{"z.string", "Zod string validation", []string{"zod"}},
}}

// jsTestingTable is read from test files only.
var jsTestingTable = keywordTable{"testing", jsExts, IsTestFile, []keywordRow{
	{"describe(", "Test suite definition (Jest/Vitest/Mocha)", nil},
	{"it(", "Test case (Jest/Vitest/Mocha)", nil},
	{"test(", "Test case (Jest/Vitest)", nil},
	{"expect(", "Assertion (Jest/Vitest/Chai)", nil},
	{"vi.mock", "Vitest mocking", []string{"vitest"}},
	{"jest.mock", "Jest mocking", []string{"jest", "@jest/globals"}},
	{"@testing-library", "Testing Library", testingLibPkgs},
	{"render(", "Testing Library render", testingLibPkgs},
	{"screen.", "Testing Library screen queries", testingLibPkgs},
	{"userEvent", "Testing Library user events", testingLibPkgs},
	{"fireEvent", "Testing Library fire events", testingLibPkgs},
	{"cy.", "Cypress commands", []string{"cypress"}},
	{"playwright", "Playwright testing", []string{"playwright", "@playwright/test"}},
}}

var jsStylingTable = keywordTable{"styling", []string{".js", ".jsx", ".ts", ".tsx", ".css", ".scss"}, nonTestFile, []keywordRow{
	{"className=", "CSS class usage", nil},
	{"tailwind", "Tailwind CSS", tailwindPkgs},
	{"@apply", "Tailwind @apply directive", tailwindPkgs},
	{"styled.", "styled-components", []string{"styled-components", "@emotion/styled"}},
	{"css`", "Emotion/styled-components CSS", []string{"@emotion/react", "@emotion/css", "styled-components"}},
	{"makeStyles", "Material-UI makeStyles", muiStylesPkgs},
	{"useStyles", "Material-UI styles hook", muiStylesPkgs},
	{"sx={", "MUI sx prop", []string{"@mui/material", "@mui/system"}},
	{"clsx", "clsx class utility", []string{"clsx"}},
	{"cn(", "shadcn/ui cn utility", []string{"tailwind-merge", "clsx"}},
	{"cva(", "Class Variance Authority", []string{"class-variance-authority"}},
	{"module.css", "CSS Modules", nil},
}}

var jsAuthTable = keywordTable{"auth", jsExts, nonTestFile, []keywordRow{
	{"useAuth", "Custom auth hook", reactPkgs},
	{"useSession", "NextAuth/custom session hook", nextAuthPkgs},
	{"signIn", "Auth sign in function", authPkgs},
	{"signOut", "Auth sign out function", authPkgs},
	{"getSession", "NextAuth getSession", nextAuthPkgs},
	{"NextAuth", "NextAuth.js", nextAuthPkgs},
	{"Auth0", "Auth0 integration", []string{"auth0", "@auth0/"}},
	{"clerk", "Clerk authentication", []string{"@clerk/"}},
	{"supabase.auth", "Supabase authentication", []string{"@supabase/"}},
	{"firebase.auth", "Firebase authentication", []string{"firebase", "firebase-admin"}},
	{"jwt", "JWT handling", []string{"jsonwebtoken", "jose", "jwt-decode", "express-jwt", "passport-jwt", "@nestjs/jwt"}},
}}

// apiTable is read in JS, Python and Go files; its packages come from all three.
var apiTable = keywordTable{"api", []string{".js", ".jsx", ".ts", ".tsx", ".py", ".go"}, nonTestFile, []keywordRow{
	{"GraphQL", "GraphQL API", graphqlPkgs},
	{"gql`", "GraphQL query", []string{"graphql-tag", "@apollo/client", "graphql-request", "urql"}},
	{"useQuery", "GraphQL/React Query", append([]string{"@apollo/client", "urql"}, reactQueryPkgs...)},
	{"useMutation", "GraphQL/React Query mutation", append([]string{"@apollo/client", "urql"}, reactQueryPkgs...)},
	{"tRPC", "tRPC type-safe API", []string{"@trpc/"}},
	{"OpenAPI", "OpenAPI/Swagger spec", []string{"openapi-typescript", "openapi-fetch", "swagger-ui-express", "@nestjs/swagger",
		"fastapi", "github.com/getkin/kin-openapi", "github.com/swaggo/swag"}},
	{"grpc", "gRPC protocol", []string{"@grpc/grpc-js", "grpcio", "google.golang.org/grpc"}},
	{"socket.io", "Socket.IO real-time", []string{"socket.io", "socket.io-client"}},
}}

var jsDatabaseTable = keywordTable{"database", []string{".js", ".ts", ".tsx"}, nonTestFile, []keywordRow{
	{"prisma", "Prisma ORM", prismaPkgs},
	{"PrismaClient", "Prisma client", prismaPkgs},
	{"drizzle", "Drizzle ORM", []string{"drizzle-orm"}},
	{"typeorm", "TypeORM", []string{"typeorm"}},
	{"@Entity", "TypeORM entity decorator", []string{"typeorm", "@mikro-orm/core"}},
	{"sequelize", "Sequelize ORM", []string{"sequelize", "sequelize-typescript"}},
	{"mongoose", "Mongoose ODM (MongoDB)", []string{"mongoose"}},
	{"knex", "Knex.js query builder", []string{"knex"}},
	{"kysely", "Kysely type-safe SQL", []string{"kysely"}},
	{"supabase", "Supabase client", []string{"@supabase/"}},
}}

var jsUtilitiesTable = keywordTable{"utilities", jsExts, nonTestFile, []keywordRow{
	{"lodash", "Lodash utility library", []string{"lodash", "lodash-es"}},
	{"dayjs", "Day.js date library", []string{"dayjs"}},
	{"moment", "Moment.js date library", []string{"moment"}},
	{"date-fns", "date-fns date utilities", []string{"date-fns"}},
	{"uuid", "UUID generation", []string{"uuid"}},
	{"nanoid", "Nano ID generation", []string{"nanoid"}},
	{"zod", "Zod schema validation", []string{"zod"}},
	{"yup", "Yup schema validation", []string{"yup"}},
	{"ajv", "AJV JSON schema validation", []string{"ajv"}},
	{"immer", "Immer immutable updates", []string{"immer"}},
	{"ramda", "Ramda FP utilities", []string{"ramda"}},
}}

// bareCalls are keywords that also must not follow '.', so that method calls
// such as Object.create( or foo.it( do not count.
var bareCalls = map[string]bool{
	"it(": true, "test(": true, "describe(": true, "expect(": true, "render(": true, "ref(": true, "watch(": true,
	"atom(": true, "create(": true, "path(": true, "include(": true, "register(": true, "cn(": true, "cva(": true,
}

func isIdentRune(r rune) bool {
	return r == '_' || r == '$' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// containsKeyword reports whether content holds kw where it is not the tail
// of a longer identifier: a keyword starting with a letter, digit, '_' or '$'
// must not follow one of those, and a bare call must not follow '.' either.
// The end of the keyword is not checked.
func containsKeyword(content, kw string) bool {
	first, _ := utf8.DecodeRuneInString(kw)
	for from := 0; ; {
		i := strings.Index(content[from:], kw)
		if i < 0 {
			return false
		}
		i += from
		prev, _ := utf8.DecodeLastRuneInString(content[:i])
		if !isIdentRune(first) || !isIdentRune(prev) && (prev != '.' || !bareCalls[kw]) {
			return true
		}
		from = i + 1
	}
}

func nonTestFile(rel string) bool { return !IsTestFile(rel) }

func anyFile(string) bool { return true }

// eachFile calls fn with the content of every file that has one of
// extensions and that include accepts. Test files feed only testing
// patterns, so most scans pass nonTestFile.
func (d *CodePatternDetector) eachFile(extensions []string, include func(rel string) bool, fn func(rel, content string)) {
	for _, f := range d.files {
		hasExt := slices.ContainsFunc(extensions, func(ext string) bool { return strings.HasSuffix(f.Name, ext) })
		if f.IsDir || !hasExt || !include(f.Path) {
			continue
		}
		if content, err := os.ReadFile(filepath.Join(d.rootPath, f.Path)); err == nil {
			fn(f.Path, string(content))
		}
	}
}

// scanKeywords returns, per keyword, the files include accepts that contain it.
func (d *CodePatternDetector) scanKeywords(keywords, extensions []string, include func(rel string) bool) map[string][]string {
	results := make(map[string][]string)
	d.eachFile(extensions, include, func(rel, content string) {
		for _, kw := range keywords {
			if containsKeyword(content, kw) {
				results[kw] = append(results[kw], rel)
			}
		}
	})
	return results
}

// scanForKeywords scans non-test files for keywords.
func (d *CodePatternDetector) scanForKeywords(keywords, extensions []string) map[string][]string {
	return d.scanKeywords(keywords, extensions, nonTestFile)
}

// scanTestKeywords scans test files only.
func (d *CodePatternDetector) scanTestKeywords(keywords, extensions []string) map[string][]string {
	return d.scanKeywords(keywords, extensions, IsTestFile)
}

// scanForRegex returns the non-test files that match pattern.
func (d *CodePatternDetector) scanForRegex(pattern string, extensions []string) []string {
	var matches []string
	re := regexp.MustCompile(pattern)
	d.eachFile(extensions, nonTestFile, func(rel, content string) {
		if re.MatchString(content) {
			matches = append(matches, rel)
		}
	})
	return matches
}

// PatternWithImport defines a pattern that requires a specific import to be valid
type PatternWithImport struct {
	Keyword     string   // The keyword to search for
	Description string   // Description of the pattern
	Imports     []string // Required imports (any of these must be present)
}

// scanForKeywordsWithImports scans non-test files for keywords, counting a
// file only when it also contains one of the pattern's imports, so that
// framework-specific keywords match only where the framework is imported.
func (d *CodePatternDetector) scanForKeywordsWithImports(patterns []PatternWithImport, extensions []string) map[string][]string {
	results := make(map[string][]string)
	d.eachFile(extensions, nonTestFile, func(rel, content string) {
		for _, p := range patterns {
			hasImport := len(p.Imports) == 0 || slices.ContainsFunc(p.Imports, func(imp string) bool { return strings.Contains(content, imp) })
			if hasImport && containsKeyword(content, p.Keyword) {
				results[p.Keyword] = append(results[p.Keyword], rel)
			}
		}
	})
	return results
}

// detectStateManagement detects state management patterns
func (d *CodePatternDetector) detectStateManagement() []types.PatternInfo {
	patterns := slices.Concat(d.report(reactStateTable), d.report(vueStateTable))
	pyExts := []string{".py"}

	// Python state patterns - use import-aware scanning to reduce false positives
	pyPatterns := []PatternWithImport{
		{Keyword: "session[", Description: "Flask session management", Imports: []string{"from flask", "import flask"}},
		{Keyword: "g.", Description: "Flask application context globals", Imports: []string{"from flask import g", "from flask import", "import flask"}},
		{Keyword: "current_app", Description: "Flask current application context", Imports: []string{"from flask", "import flask"}},
		{Keyword: "request.", Description: "Flask/FastAPI request object", Imports: []string{"from flask", "import flask", "from fastapi", "import fastapi", "from starlette", "import starlette"}},
	}

	pyResults := d.scanForKeywordsWithImports(pyPatterns, pyExts)
	pyDescriptions := make(map[string]string)
	for _, p := range pyPatterns {
		pyDescriptions[p.Keyword] = p.Description
	}

	for kw, files := range pyResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "state",
				Description: pyDescriptions[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
}

// detectDataFetching detects data fetching patterns
func (d *CodePatternDetector) detectDataFetching() []types.PatternInfo {
	patterns := d.report(jsDataFetchingTable)
	pyExts := []string{".py"}
	goExts := []string{".go"}

	// Python data fetching
	pyKeywords := map[string]string{
		"requests.": "Python requests library",
		"httpx.":    "HTTPX async HTTP client",
		"aiohttp.":  "aiohttp async HTTP client",
		"urllib":    "Python urllib",
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
		"http.Get":    "Go standard HTTP GET",
		"http.Post":   "Go standard HTTP POST",
		"http.Client": "Go HTTP client",
		"resty.":      "Resty HTTP client",
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
	patterns := d.report(jsRoutingTable)
	pyExts := []string{".py"}
	goExts := []string{".go"}

	// Python routing - use import-aware scanning to reduce false positives
	pyRoutingPatterns := []PatternWithImport{
		{Keyword: "@app.route", Description: "Flask route decorator", Imports: []string{"from flask", "import flask"}},
		{Keyword: "@router.", Description: "FastAPI router decorator", Imports: []string{"from fastapi", "import fastapi"}},
		{Keyword: "@app.get", Description: "FastAPI GET endpoint", Imports: []string{"from fastapi", "import fastapi"}},
		{Keyword: "@app.post", Description: "FastAPI POST endpoint", Imports: []string{"from fastapi", "import fastapi"}},
		{Keyword: "Blueprint", Description: "Flask Blueprint for modular routing", Imports: []string{"from flask", "import flask"}},
		{Keyword: "APIRouter", Description: "FastAPI APIRouter", Imports: []string{"from fastapi", "import fastapi"}},
		{Keyword: "path(", Description: "Django URL path", Imports: []string{"from django.urls", "django.urls.path"}},
		{Keyword: "include(", Description: "Django URL includes", Imports: []string{"from django.urls", "django.urls.include"}},
	}

	pyRoutingResults := d.scanForKeywordsWithImports(pyRoutingPatterns, pyExts)
	pyRoutingDescriptions := make(map[string]string)
	for _, p := range pyRoutingPatterns {
		pyRoutingDescriptions[p.Keyword] = p.Description
	}

	for kw, files := range pyRoutingResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "routing",
				Description: pyRoutingDescriptions[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Go routing
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
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "routing",
				Description: goKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
}

// detectForms detects form handling patterns
func (d *CodePatternDetector) detectForms() []types.PatternInfo {
	return d.report(jsFormsTable)
}

// detectTesting detects testing patterns
func (d *CodePatternDetector) detectTesting() []types.PatternInfo {
	patterns := d.report(jsTestingTable)
	pyExts := []string{".py"}
	goTestExts := []string{"_test.go"}
	goExts := []string{".go"}

	// Python testing, in test files only
	pyKeywords := map[string]string{
		"pytest":          "Pytest framework",
		"def test_":       "Pytest test function",
		"unittest":        "Python unittest",
		"TestCase":        "unittest TestCase class",
		"@pytest.fixture": "Pytest fixture",
		"mock.":           "Python mocking",
		"@patch":          "unittest mock patch",
	}

	pyResults := d.scanTestKeywords(keys(pyKeywords), pyExts)
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
	goTestResults := d.scanTestKeywords(goTestKeywords, goTestExts)

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

	goHelperResults := d.scanKeywords(keys(goHelperKeywords), goExts, anyFile)
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
	return d.report(jsStylingTable)
}

// detectAuthentication detects auth patterns
func (d *CodePatternDetector) detectAuthentication() []types.PatternInfo {
	patterns := d.report(jsAuthTable)
	pyExts := []string{".py"}
	goExts := []string{".go"}

	// Python auth
	pyKeywords := map[string]string{
		"login_required": "Flask login decorator",
		"current_user":   "Flask-Login current user",
		"@jwt_required":  "JWT required decorator",
		"OAuth":          "OAuth integration",
		"HTTPBearer":     "FastAPI HTTP Bearer auth",
		"Depends(":       "FastAPI dependency injection",
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
		"jwt.": "JWT handling",
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
	return d.report(apiTable)
}

// detectDatabasePatterns detects database/ORM patterns
func (d *CodePatternDetector) detectDatabasePatterns() []types.PatternInfo {
	patterns := d.report(jsDatabaseTable)
	pyExts := []string{".py"}
	goExts := []string{".go"}

	// Python ORMs
	pyKeywords := map[string]string{
		"SQLAlchemy":    "SQLAlchemy ORM",
		"Base.metadata": "SQLAlchemy models",
		"django.db":     "Django ORM",
		"models.Model":  "Django model",
		"tortoise":      "Tortoise ORM",
		"peewee":        "Peewee ORM",
		"mongoengine":   "MongoEngine ODM",
		"motor":         "Motor async MongoDB",
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
		"gorm.Open(":    "GORM ORM",
		"gorm.Model":    "GORM model embedding",
		"sqlx.Connect":  "sqlx database library",
		"sqlx.Open":     "sqlx database library",
		"sql.Open(":     "Go standard SQL",
		"pgx.Connect":   "pgx PostgreSQL driver",
		"mongo.Connect": "MongoDB Go driver",
		"bun.NewDB":     "Bun ORM",
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
	return d.report(jsUtilitiesTable)
}

// detectGoPatterns detects Go-specific patterns
func (d *CodePatternDetector) detectGoPatterns() []types.PatternInfo {
	var patterns []types.PatternInfo
	goExts := []string{".go"}

	// CLI frameworks
	cliKeywords := map[string]string{
		"cobra.Command": "Cobra CLI framework",
		"urfave/cli":    "urfave/cli framework",
		"kingpin.":      "Kingpin CLI framework",
		"spf13/pflag":   "spf13 pflag for flags",
	}

	cliResults := d.scanForKeywords(keys(cliKeywords), goExts)
	for kw, files := range cliResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "go-cli",
				Description: cliKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Configuration management
	configKeywords := map[string]string{
		"viper.":      "Viper configuration",
		"spf13/viper": "Viper config library",
		"godotenv":    "GoDotEnv environment variables",
		"envconfig":   "Kelseyhightower envconfig",
	}

	configResults := d.scanForKeywords(keys(configKeywords), goExts)
	for kw, files := range configResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "go-config",
				Description: configKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Concurrency patterns
	concurrencyKeywords := map[string]string{
		"go func":             "Goroutines",
		"make(chan":           "Channels",
		"sync.Mutex":          "Mutex synchronization",
		"sync.RWMutex":        "Read-write mutex",
		"sync.WaitGroup":      "WaitGroup for goroutine coordination",
		"sync.Once":           "sync.Once for one-time initialization",
		"select {":            "Select statement for channel operations",
		"context.Context":     "Context for cancellation and deadlines",
		"context.WithCancel":  "Cancelable context",
		"context.WithTimeout": "Context with timeout",
	}

	concurrencyResults := d.scanForKeywords(keys(concurrencyKeywords), goExts)
	for kw, files := range concurrencyResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "go-concurrency",
				Description: concurrencyKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Logging libraries
	loggingKeywords := map[string]string{
		"logrus.":  "Logrus structured logging",
		"zap.":     "Zap high-performance logging",
		"zerolog.": "Zerolog logging",
		"slog.":    "Go 1.21+ structured logging",
	}

	loggingResults := d.scanForKeywords(keys(loggingKeywords), goExts)
	for kw, files := range loggingResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "go-logging",
				Description: loggingKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Error handling patterns
	errorKeywords := map[string]string{
		"errors.Is":   "Go 1.13+ error comparison",
		"errors.As":   "Go 1.13+ error type assertion",
		"errors.Wrap": "pkg/errors error wrapping",
		"fmt.Errorf":  "Error formatting",
	}

	errorResults := d.scanForKeywords(keys(errorKeywords), goExts)
	for kw, files := range errorResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "go-errors",
				Description: errorKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Dependency injection
	diKeywords := map[string]string{
		"wire.": "Google Wire dependency injection",
		"dig.":  "Uber Dig dependency injection",
		"fx.":   "Uber Fx application framework",
	}

	diResults := d.scanForKeywords(keys(diKeywords), goExts)
	for kw, files := range diResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "go-di",
				Description: diKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Serialization and data formats
	serializationKeywords := map[string]string{
		"json.Marshal":   "JSON serialization",
		"json.Unmarshal": "JSON deserialization",
		"yaml.Marshal":   "YAML serialization",
		"yaml.Unmarshal": "YAML deserialization",
		"encoding/xml":   "XML encoding",
	}

	serializationResults := d.scanForKeywords(keys(serializationKeywords), goExts)
	for kw, files := range serializationResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "go-serialization",
				Description: serializationKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
}

// detectRustPatterns detects Rust-specific patterns
func (d *CodePatternDetector) detectRustPatterns() []types.PatternInfo {
	var patterns []types.PatternInfo
	rsExts := []string{".rs"}

	// Async runtime
	asyncKeywords := map[string]string{
		"tokio::":        "Tokio async runtime",
		"#[tokio::main]": "Tokio main entry point",
		"#[tokio::test]": "Tokio async test",
		"async_std::":    "async-std runtime",
		"async fn":       "Async function definition",
		".await":         "Await syntax",
		"futures::":      "Futures crate utilities",
		"Pin<":           "Pinned futures",
		"poll(":          "Future polling",
	}

	asyncResults := d.scanForKeywords(keys(asyncKeywords), rsExts)
	for kw, files := range asyncResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "rust-async",
				Description: asyncKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Web frameworks
	webKeywords := map[string]string{
		"actix_web::":   "Actix-web framework",
		"#[get(":        "Actix/Rocket route macro",
		"#[post(":       "Actix/Rocket POST route",
		"axum::":        "Axum web framework",
		"Router::new()": "Axum router",
		"rocket::":      "Rocket web framework",
		"#[launch]":     "Rocket launch macro",
		"warp::":        "Warp web framework",
		"hyper::":       "Hyper HTTP library",
	}

	webResults := d.scanForKeywords(keys(webKeywords), rsExts)
	for kw, files := range webResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "rust-web",
				Description: webKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Serialization
	serdeKeywords := map[string]string{
		"serde::":              "Serde serialization",
		"#[derive(Serialize":   "Serde Serialize derive",
		"#[derive(Deserialize": "Serde Deserialize derive",
		"serde_json::":         "Serde JSON",
		"serde_yaml::":         "Serde YAML",
		"toml::":               "TOML parsing",
	}

	serdeResults := d.scanForKeywords(keys(serdeKeywords), rsExts)
	for kw, files := range serdeResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "rust-serde",
				Description: serdeKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Error handling
	errorKeywords := map[string]string{
		"anyhow::":    "Anyhow error handling",
		"thiserror::": "thiserror derive macros",
		"#[error(":    "thiserror error attribute",
		"Result<":     "Result type usage",
		"?;":          "Question mark operator",
		".unwrap()":   "Unwrap (panic on error)",
		".expect(":    "Expect (panic with message)",
		".unwrap_or(": "Unwrap with default",
		".map_err(":   "Error mapping",
	}

	errorResults := d.scanForKeywords(keys(errorKeywords), rsExts)
	for kw, files := range errorResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "rust-error",
				Description: errorKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// CLI
	cliKeywords := map[string]string{
		"clap::":          "Clap CLI framework",
		"#[derive(Parser": "Clap derive Parser",
		"#[command(":      "Clap command attribute",
		"#[arg(":          "Clap argument attribute",
		"structopt::":     "StructOpt (deprecated, use clap)",
		"argh::":          "Argh CLI parser",
	}

	cliResults := d.scanForKeywords(keys(cliKeywords), rsExts)
	for kw, files := range cliResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "rust-cli",
				Description: cliKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Database
	dbKeywords := map[string]string{
		"sqlx::":           "SQLx async database",
		"diesel::":         "Diesel ORM",
		"sea_orm::":        "SeaORM async ORM",
		"rusqlite::":       "Rusqlite SQLite",
		"tokio_postgres::": "Tokio PostgreSQL",
		"mongodb::":        "MongoDB driver",
		"redis::":          "Redis client",
	}

	dbResults := d.scanForKeywords(keys(dbKeywords), rsExts)
	for kw, files := range dbResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "rust-database",
				Description: dbKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Testing, in any file: unit tests live in #[cfg(test)] modules beside the code
	testKeywords := map[string]string{
		"#[test]":      "Rust unit test",
		"#[cfg(test)]": "Test module configuration",
		"assert!":      "Assert macro",
		"assert_eq!":   "Assert equality macro",
		"assert_ne!":   "Assert not equal macro",
		"mockall::":    "Mockall mocking",
		"#[mock]":      "Mockall mock attribute",
		"proptest::":   "Property-based testing",
	}

	testResults := d.scanKeywords(keys(testKeywords), rsExts, anyFile)
	for kw, files := range testResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "rust-testing",
				Description: testKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Concurrency
	concKeywords := map[string]string{
		"std::thread":   "Standard library threads",
		"Arc<":          "Atomic reference counting",
		"Mutex<":        "Mutex synchronization",
		"RwLock<":       "Read-write lock",
		"mpsc::":        "Multi-producer single-consumer channels",
		"crossbeam::":   "Crossbeam concurrency utilities",
		"rayon::":       "Rayon data parallelism",
		"parking_lot::": "Parking lot synchronization",
	}

	concResults := d.scanForKeywords(keys(concKeywords), rsExts)
	for kw, files := range concResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "rust-concurrency",
				Description: concKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Logging and tracing
	logKeywords := map[string]string{
		"log::":         "Log crate",
		"tracing::":     "Tracing framework",
		"#[instrument]": "Tracing instrument macro",
		"info!":         "Info log macro",
		"debug!":        "Debug log macro",
		"error!":        "Error log macro",
		"warn!":         "Warning log macro",
		"env_logger::":  "Environment logger",
	}

	logResults := d.scanForKeywords(keys(logKeywords), rsExts)
	for kw, files := range logResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "rust-logging",
				Description: logKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
}

// detectPythonPatterns detects Python-specific patterns beyond Flask/FastAPI
func (d *CodePatternDetector) detectPythonPatterns() []types.PatternInfo {
	var patterns []types.PatternInfo
	pyExts := []string{".py"}

	// Type hints
	typeKeywords := map[string]string{
		"from typing import": "Python typing imports",
		"Optional[":          "Optional type hint",
		"List[":              "List type hint",
		"Dict[":              "Dict type hint",
		"Union[":             "Union type hint",
		"Callable[":          "Callable type hint",
		"TypeVar":            "Generic type variable",
		"@dataclass":         "Dataclass decorator",
		"@dataclasses":       "Dataclass module",
		"pydantic":           "Pydantic validation",
		"BaseModel":          "Pydantic BaseModel",
	}

	typeResults := d.scanForKeywords(keys(typeKeywords), pyExts)
	for kw, files := range typeResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "python-typing",
				Description: typeKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Async patterns
	asyncKeywords := map[string]string{
		"async def":   "Async function definition",
		"await ":      "Await expression",
		"asyncio.":    "Asyncio module",
		"aiohttp.":    "Aiohttp async HTTP",
		"httpx.Async": "HTTPX async client",
		"anyio.":      "AnyIO async library",
		"trio.":       "Trio async library",
	}

	asyncResults := d.scanForKeywords(keys(asyncKeywords), pyExts)
	for kw, files := range asyncResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "python-async",
				Description: asyncKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Data science / ML
	dsKeywords := map[string]string{
		"import pandas":       "Pandas data analysis",
		"import numpy":        "NumPy numerical computing",
		"import matplotlib":   "Matplotlib plotting",
		"import seaborn":      "Seaborn visualization",
		"import sklearn":      "Scikit-learn ML",
		"import torch":        "PyTorch deep learning",
		"import tensorflow":   "TensorFlow deep learning",
		"import keras":        "Keras deep learning",
		"import transformers": "Hugging Face Transformers",
		"import langchain":    "LangChain LLM framework",
		"import openai":       "OpenAI API",
		"import anthropic":    "Anthropic API",
	}

	dsResults := d.scanForKeywords(keys(dsKeywords), pyExts)
	for kw, files := range dsResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "python-data-science",
				Description: dsKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// CLI patterns
	cliKeywords := map[string]string{
		"argparse.": "Argparse CLI",
		"click.":    "Click CLI framework",
		"@click.":   "Click decorator",
		"typer.":    "Typer CLI framework",
		"fire.":     "Python Fire CLI",
		"rich.":     "Rich terminal formatting",
	}

	cliResults := d.scanForKeywords(keys(cliKeywords), pyExts)
	for kw, files := range cliResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "python-cli",
				Description: cliKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Logging and debugging
	logKeywords := map[string]string{
		"import logging":    "Python logging module",
		"logging.getLogger": "Logger instance",
		"structlog.":        "Structlog structured logging",
		"loguru.":           "Loguru logging",
	}

	logResults := d.scanForKeywords(keys(logKeywords), pyExts)
	for kw, files := range logResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "python-logging",
				Description: logKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	// Task queues and background jobs
	taskKeywords := map[string]string{
		"from celery":     "Celery task queue",
		"@celery.task":    "Celery task decorator",
		"import rq":       "RQ (Redis Queue)",
		"import dramatiq": "Dramatiq task processing",
		"import huey":     "Huey task queue",
	}

	taskResults := d.scanForKeywords(keys(taskKeywords), pyExts)
	for kw, files := range taskResults {
		if len(files) > 0 {
			patterns = append(patterns, types.PatternInfo{
				Name:        kw,
				Category:    "python-tasks",
				Description: taskKeywords[kw],
				FileCount:   len(files),
				Examples:    limitSlice(files, 3),
			})
		}
	}

	return patterns
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
