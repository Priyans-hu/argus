package detector

import (
	"os"
	"path/filepath"
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

// Python rows name PEP 503 package names; rows that are import statements,
// standard library or syntax are intrinsic.
var pyExts = []string{".py"}

var pyDataFetchingTable = keywordTable{"data-fetching", pyExts, nonTestFile, []keywordRow{
	{"requests.", "Python requests library", []string{"requests"}},
	{"httpx.", "HTTPX async HTTP client", []string{"httpx"}},
	{"aiohttp.", "aiohttp async HTTP client", []string{"aiohttp"}},
	{"urllib", "Python urllib", nil},
}}

// pyTestingTable is read from test files only.
var pyTestingTable = keywordTable{"testing", pyExts, IsTestFile, []keywordRow{
	{"pytest", "Pytest framework", []string{"pytest"}},
	{"def test_", "Pytest test function", nil},
	{"unittest", "Python unittest", nil},
	{"TestCase", "unittest TestCase class", nil},
	{"@pytest.fixture", "Pytest fixture", []string{"pytest"}},
	{"mock.", "Python mocking", []string{"pytest-mock", "mock"}},
	{"@patch", "unittest mock patch", nil},
}}

var pyAuthTable = keywordTable{"auth", pyExts, nonTestFile, []keywordRow{
	{"login_required", "Flask login decorator", []string{"flask-login", "django"}},
	{"current_user", "Flask-Login current user", []string{"flask-login"}},
	{"@jwt_required", "JWT required decorator", []string{"flask-jwt-extended"}},
	{"OAuth", "OAuth integration", []string{"authlib", "oauthlib"}},
	{"HTTPBearer", "FastAPI HTTP Bearer auth", []string{"fastapi"}},
	{"Depends(", "FastAPI dependency injection", []string{"fastapi"}},
}}

var pyDatabaseTable = keywordTable{"database", pyExts, nonTestFile, []keywordRow{
	{"SQLAlchemy", "SQLAlchemy ORM", []string{"sqlalchemy", "flask-sqlalchemy"}},
	{"Base.metadata", "SQLAlchemy models", []string{"sqlalchemy", "flask-sqlalchemy"}},
	{"django.db", "Django ORM", []string{"django"}},
	{"models.Model", "Django model", []string{"django"}},
	{"tortoise", "Tortoise ORM", []string{"tortoise-orm"}},
	{"peewee", "Peewee ORM", []string{"peewee"}},
	{"mongoengine", "MongoEngine ODM", []string{"mongoengine"}},
	{"motor", "Motor async MongoDB", []string{"motor"}},
}}

var pythonTables = []keywordTable{
	{"python-typing", pyExts, nonTestFile, []keywordRow{
		{"from typing import", "Python typing imports", nil},
		{"Optional[", "Optional type hint", nil},
		{"List[", "List type hint", nil},
		{"Dict[", "Dict type hint", nil},
		{"Union[", "Union type hint", nil},
		{"Callable[", "Callable type hint", nil},
		{"TypeVar", "Generic type variable", nil},
		{"@dataclass", "Dataclass decorator", nil},
		{"@dataclasses", "Dataclass module", nil},
		{"pydantic", "Pydantic validation", []string{"pydantic"}},
		{"BaseModel", "Pydantic BaseModel", []string{"pydantic"}},
	}},
	{"python-async", pyExts, nonTestFile, []keywordRow{
		{"async def", "Async function definition", nil},
		{"await ", "Await expression", nil},
		{"asyncio.", "Asyncio module", nil},
		{"aiohttp.", "Aiohttp async HTTP", []string{"aiohttp"}},
		{"httpx.Async", "HTTPX async client", []string{"httpx"}},
		{"anyio.", "AnyIO async library", []string{"anyio"}},
		{"trio.", "Trio async library", []string{"trio"}},
	}},
	{"python-data-science", pyExts, nonTestFile, []keywordRow{
		{"import pandas", "Pandas data analysis", nil},
		{"import numpy", "NumPy numerical computing", nil},
		{"import matplotlib", "Matplotlib plotting", nil},
		{"import seaborn", "Seaborn visualization", nil},
		{"import sklearn", "Scikit-learn ML", nil},
		{"import torch", "PyTorch deep learning", nil},
		{"import tensorflow", "TensorFlow deep learning", nil},
		{"import keras", "Keras deep learning", nil},
		{"import transformers", "Hugging Face Transformers", nil},
		{"import langchain", "LangChain LLM framework", nil},
		{"import openai", "OpenAI API", nil},
		{"import anthropic", "Anthropic API", nil},
	}},
	{"python-cli", pyExts, nonTestFile, []keywordRow{
		{"argparse.", "Argparse CLI", nil},
		{"click.", "Click CLI framework", []string{"click"}},
		{"@click.", "Click decorator", []string{"click"}},
		{"typer.", "Typer CLI framework", []string{"typer"}},
		{"fire.", "Python Fire CLI", []string{"fire"}},
		{"rich.", "Rich terminal formatting", []string{"rich"}},
	}},
	{"python-logging", pyExts, nonTestFile, []keywordRow{
		{"import logging", "Python logging module", nil},
		{"logging.getLogger", "Logger instance", nil},
		{"structlog.", "Structlog structured logging", []string{"structlog"}},
		{"loguru.", "Loguru logging", []string{"loguru"}},
	}},
	{"python-tasks", pyExts, nonTestFile, []keywordRow{
		{"from celery", "Celery task queue", nil},
		{"@celery.task", "Celery task decorator", []string{"celery"}},
		{"import rq", "RQ (Redis Queue)", nil},
		{"import dramatiq", "Dramatiq task processing", nil},
		{"import huey", "Huey task queue", nil},
	}},
}

// Rust rows name crates (lower-case, "-" for "_"); syntax and standard
// library rows are intrinsic.
var rsExts = []string{".rs"}

var (
	actixRocketCrates = []string{"actix-web", "rocket"}
	logCrates         = []string{"log", "tracing"}
)

var rustTables = []keywordTable{
	{"rust-async", rsExts, nonTestFile, []keywordRow{
		{"tokio::", "Tokio async runtime", []string{"tokio"}},
		{"#[tokio::main]", "Tokio main entry point", []string{"tokio"}},
		{"#[tokio::test]", "Tokio async test", []string{"tokio"}},
		{"async_std::", "async-std runtime", []string{"async-std"}},
		{"async fn", "Async function definition", nil},
		{".await", "Await syntax", nil},
		{"futures::", "Futures crate utilities", []string{"futures"}},
		{"Pin<", "Pinned futures", nil},
		{"poll(", "Future polling", nil},
	}},
	{"rust-web", rsExts, nonTestFile, []keywordRow{
		{"actix_web::", "Actix-web framework", []string{"actix-web"}},
		{"#[get(", "Actix/Rocket route macro", actixRocketCrates},
		{"#[post(", "Actix/Rocket POST route", actixRocketCrates},
		{"axum::", "Axum web framework", []string{"axum"}},
		{"Router::new()", "Axum router", []string{"axum"}},
		{"rocket::", "Rocket web framework", []string{"rocket"}},
		{"#[launch]", "Rocket launch macro", []string{"rocket"}},
		{"warp::", "Warp web framework", []string{"warp"}},
		{"hyper::", "Hyper HTTP library", []string{"hyper"}},
	}},
	{"rust-serde", rsExts, nonTestFile, []keywordRow{
		{"serde::", "Serde serialization", []string{"serde"}},
		{"#[derive(Serialize", "Serde Serialize derive", []string{"serde"}},
		{"#[derive(Deserialize", "Serde Deserialize derive", []string{"serde"}},
		{"serde_json::", "Serde JSON", []string{"serde-json"}},
		{"serde_yaml::", "Serde YAML", []string{"serde-yaml"}},
		{"toml::", "TOML parsing", []string{"toml"}},
	}},
	{"rust-error", rsExts, nonTestFile, []keywordRow{
		{"anyhow::", "Anyhow error handling", []string{"anyhow"}},
		{"thiserror::", "thiserror derive macros", []string{"thiserror"}},
		{"#[error(", "thiserror error attribute", []string{"thiserror"}},
		{"Result<", "Result type usage", nil},
		{"?;", "Question mark operator", nil},
		{".unwrap()", "Unwrap (panic on error)", nil},
		{".expect(", "Expect (panic with message)", nil},
		{".unwrap_or(", "Unwrap with default", nil},
		{".map_err(", "Error mapping", nil},
	}},
	{"rust-cli", rsExts, nonTestFile, []keywordRow{
		{"clap::", "Clap CLI framework", []string{"clap"}},
		{"#[derive(Parser", "Clap derive Parser", []string{"clap"}},
		{"#[command(", "Clap command attribute", []string{"clap"}},
		{"#[arg(", "Clap argument attribute", []string{"clap"}},
		{"structopt::", "StructOpt (deprecated, use clap)", []string{"structopt"}},
		{"argh::", "Argh CLI parser", []string{"argh"}},
	}},
	{"rust-database", rsExts, nonTestFile, []keywordRow{
		{"sqlx::", "SQLx async database", []string{"sqlx"}},
		{"diesel::", "Diesel ORM", []string{"diesel"}},
		{"sea_orm::", "SeaORM async ORM", []string{"sea-orm"}},
		{"rusqlite::", "Rusqlite SQLite", []string{"rusqlite"}},
		{"tokio_postgres::", "Tokio PostgreSQL", []string{"tokio-postgres"}},
		{"mongodb::", "MongoDB driver", []string{"mongodb"}},
		{"redis::", "Redis client", []string{"redis"}},
	}},
	// read from any file: unit tests live in #[cfg(test)] modules beside the code
	{"rust-testing", rsExts, anyFile, []keywordRow{
		{"#[test]", "Rust unit test", nil},
		{"#[cfg(test)]", "Test module configuration", nil},
		{"assert!", "Assert macro", nil},
		{"assert_eq!", "Assert equality macro", nil},
		{"assert_ne!", "Assert not equal macro", nil},
		{"mockall::", "Mockall mocking", []string{"mockall"}},
		{"#[mock]", "Mockall mock attribute", []string{"mockall"}},
		{"proptest::", "Property-based testing", []string{"proptest"}},
	}},
	{"rust-concurrency", rsExts, nonTestFile, []keywordRow{
		{"std::thread", "Standard library threads", nil},
		{"Arc<", "Atomic reference counting", nil},
		{"Mutex<", "Mutex synchronization", nil},
		{"RwLock<", "Read-write lock", nil},
		{"mpsc::", "Multi-producer single-consumer channels", nil},
		{"crossbeam::", "Crossbeam concurrency utilities", []string{"crossbeam"}},
		{"rayon::", "Rayon data parallelism", []string{"rayon"}},
		{"parking_lot::", "Parking lot synchronization", []string{"parking-lot"}},
	}},
	{"rust-logging", rsExts, nonTestFile, []keywordRow{
		{"log::", "Log crate", []string{"log"}},
		{"tracing::", "Tracing framework", []string{"tracing"}},
		{"#[instrument]", "Tracing instrument macro", []string{"tracing"}},
		{"info!", "Info log macro", logCrates},
		{"debug!", "Debug log macro", logCrates},
		{"error!", "Error log macro", logCrates},
		{"warn!", "Warning log macro", logCrates},
		{"env_logger::", "Environment logger", []string{"env-logger"}},
	}},
}

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
	patterns := slices.Concat(d.report(jsDataFetchingTable), d.report(pyDataFetchingTable))
	goExts := []string{".go"}

	// Go HTTP clients
	goKeywords := map[string]string{
		"http.Get":    "Go standard HTTP GET",
		"http.Post":   "Go standard HTTP POST",
		"http.Client": "Go HTTP client",
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
	patterns := slices.Concat(d.report(jsTestingTable), d.report(pyTestingTable))
	goTestExts := []string{"_test.go"}
	goExts := []string{".go"}

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
	patterns := slices.Concat(d.report(jsAuthTable), d.report(pyAuthTable))

	return patterns
}

// detectAPIPatterns detects API design patterns
func (d *CodePatternDetector) detectAPIPatterns() []types.PatternInfo {
	return d.report(apiTable)
}

// detectDatabasePatterns detects database/ORM patterns
func (d *CodePatternDetector) detectDatabasePatterns() []types.PatternInfo {
	patterns := slices.Concat(d.report(jsDatabaseTable), d.report(pyDatabaseTable))
	goExts := []string{".go"}

	// Go standard SQL; Go database libraries come from the Go AST detector
	if files := d.scanForKeywords([]string{"sql.Open("}, goExts)["sql.Open("]; len(files) > 0 {
		patterns = append(patterns, types.PatternInfo{
			Name:        "sql.Open",
			Category:    "database",
			Description: "Go standard SQL",
			FileCount:   len(files),
			Examples:    limitSlice(files, 3),
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
	for _, t := range rustTables {
		patterns = append(patterns, d.report(t)...)
	}
	return patterns
}

// detectPythonPatterns detects Python-specific patterns beyond Flask/FastAPI
func (d *CodePatternDetector) detectPythonPatterns() []types.PatternInfo {
	var patterns []types.PatternInfo
	for _, t := range pythonTables {
		patterns = append(patterns, d.report(t)...)
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
