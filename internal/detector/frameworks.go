package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// FrameworkDetector detects framework-specific patterns and conventions
type FrameworkDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewFrameworkDetector creates a new framework detector
func NewFrameworkDetector(rootPath string, files []types.FileInfo) *FrameworkDetector {
	return &FrameworkDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// Detect analyzes the codebase for framework-specific patterns
func (d *FrameworkDetector) Detect() ([]types.Convention, error) {
	var conventions []types.Convention

	// Detect React patterns
	conventions = append(conventions, d.detectReactPatterns()...)

	// Detect Vue patterns
	conventions = append(conventions, d.detectVuePatterns()...)

	// Detect Angular patterns
	conventions = append(conventions, d.detectAngularPatterns()...)

	// Detect Next.js patterns
	conventions = append(conventions, d.detectNextJSPatterns()...)

	// Detect Spring Boot patterns
	conventions = append(conventions, d.detectSpringBootPatterns()...)

	// Detect Express/Fastify patterns
	conventions = append(conventions, d.detectNodeBackendPatterns()...)

	// Detect Django/FastAPI patterns
	conventions = append(conventions, d.detectPythonBackendPatterns()...)

	// Detect Go web framework patterns
	conventions = append(conventions, d.detectGoWebPatterns()...)

	return conventions, nil
}

// detectReactPatterns detects React-specific patterns
func (d *FrameworkDetector) detectReactPatterns() []types.Convention {
	var conventions []types.Convention

	// Check if React is used
	if !d.hasFramework("react") {
		return conventions
	}

	// Pattern counters
	functionalComponents := 0
	classComponents := 0
	hooksUsage := make(map[string]int)
	stateManagement := ""

	// Regex patterns
	functionalRegex := regexp.MustCompile(`(?:export\s+)?(?:const|function)\s+\w+\s*[=:]?\s*(?:\([^)]*\)|[^=])*\s*(?:=>|{)\s*(?:[^}]*)?(?:return\s+)?[(<]`)
	classRegex := regexp.MustCompile(`class\s+\w+\s+extends\s+(?:React\.)?(?:Component|PureComponent)`)
	useStateRegex := regexp.MustCompile(`useState\s*[<(]`)
	useEffectRegex := regexp.MustCompile(`useEffect\s*\(`)
	useContextRegex := regexp.MustCompile(`useContext\s*\(`)
	useReducerRegex := regexp.MustCompile(`useReducer\s*\(`)
	useQueryRegex := regexp.MustCompile(`use(?:Query|Mutation|QueryClient)\s*\(`)
	useSWRRegex := regexp.MustCompile(`useSWR\s*\(`)
	reduxRegex := regexp.MustCompile(`useSelector|useDispatch|connect\s*\(`)
	zustandRegex := regexp.MustCompile(`create\s*\(\s*\(\s*set\s*(?:,\s*get)?\s*\)`)
	jotaiRegex := regexp.MustCompile(`atom\s*\(|useAtom\s*\(`)
	recoilRegex := regexp.MustCompile(`atom\s*\(|useRecoilState\s*\(`)

	sampledFiles := 0
	maxSamples := 40

	for _, f := range d.files {
		if f.IsDir || sampledFiles >= maxSamples {
			continue
		}

		if !isReactFile(f.Extension, f.Name) {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)

		// Check component types
		if functionalRegex.MatchString(contentStr) && strings.Contains(contentStr, "return") {
			functionalComponents++
		}
		if classRegex.MatchString(contentStr) {
			classComponents++
		}

		// Check hooks usage
		if useStateRegex.MatchString(contentStr) {
			hooksUsage["useState"]++
		}
		if useEffectRegex.MatchString(contentStr) {
			hooksUsage["useEffect"]++
		}
		if useContextRegex.MatchString(contentStr) {
			hooksUsage["useContext"]++
		}
		if useReducerRegex.MatchString(contentStr) {
			hooksUsage["useReducer"]++
		}
		if useQueryRegex.MatchString(contentStr) {
			hooksUsage["react-query"]++
		}
		if useSWRRegex.MatchString(contentStr) {
			hooksUsage["swr"]++
		}

		// Check state management
		if reduxRegex.MatchString(contentStr) {
			stateManagement = "Redux"
		}
		if zustandRegex.MatchString(contentStr) {
			stateManagement = "Zustand"
		}
		if jotaiRegex.MatchString(contentStr) {
			stateManagement = "Jotai"
		}
		if recoilRegex.MatchString(contentStr) {
			stateManagement = "Recoil"
		}

		sampledFiles++
	}

	// Report findings
	if functionalComponents > classComponents && functionalComponents >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "react",
			Description: "Functional components with hooks (modern React pattern)",
			Example:     "const Component = () => { return <div>...</div> }",
		})
	} else if classComponents > functionalComponents && classComponents >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "react",
			Description: "Class components (legacy React pattern)",
			Example:     "class Component extends React.Component { render() { ... } }",
		})
	}

	// Report hooks usage
	if hooksUsage["useState"] >= 5 && hooksUsage["useEffect"] >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "react",
			Description: "Standard React hooks for state and effects",
		})
	}

	if hooksUsage["useContext"] >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "react",
			Description: "React Context for shared state",
		})
	}

	// Report data fetching
	if hooksUsage["react-query"] >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "react",
			Description: "TanStack Query (React Query) for server state management",
			Example:     "const { data, isLoading } = useQuery({ queryKey: ['key'], queryFn })",
		})
	} else if hooksUsage["swr"] >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "react",
			Description: "SWR for data fetching with caching",
			Example:     "const { data, error } = useSWR('/api/data', fetcher)",
		})
	}

	// Report state management
	if stateManagement != "" {
		conventions = append(conventions, types.Convention{
			Category:    "react",
			Description: stateManagement + " for global state management",
		})
	}

	return conventions
}

// detectVuePatterns detects Vue-specific patterns
func (d *FrameworkDetector) detectVuePatterns() []types.Convention {
	var conventions []types.Convention

	if !d.hasFramework("vue") {
		return conventions
	}

	compositionAPI := 0
	optionsAPI := 0
	scriptSetup := 0
	pinia := 0
	vuex := 0

	compositionRegex := regexp.MustCompile(`(?:ref|reactive|computed|watch|onMounted)\s*\(`)
	optionsRegex := regexp.MustCompile(`export\s+default\s*\{[^}]*(?:data|methods|computed|watch)\s*[:(]`)
	scriptSetupRegex := regexp.MustCompile(`<script\s+setup`)
	piniaRegex := regexp.MustCompile(`defineStore\s*\(|useStore\s*\(`)
	vuexRegex := regexp.MustCompile(`(?:mapState|mapGetters|mapActions|mapMutations)\s*\(|\$store`)

	sampledFiles := 0
	maxSamples := 30

	for _, f := range d.files {
		if f.IsDir || sampledFiles >= maxSamples {
			continue
		}

		if f.Extension != ".vue" && f.Extension != ".ts" && f.Extension != ".js" {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)

		if scriptSetupRegex.MatchString(contentStr) {
			scriptSetup++
		}
		if compositionRegex.MatchString(contentStr) {
			compositionAPI++
		}
		if optionsRegex.MatchString(contentStr) {
			optionsAPI++
		}
		if piniaRegex.MatchString(contentStr) {
			pinia++
		}
		if vuexRegex.MatchString(contentStr) {
			vuex++
		}

		sampledFiles++
	}

	// Report findings
	if scriptSetup >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "vue",
			Description: "Vue 3 <script setup> syntax (recommended)",
			Example:     "<script setup>\nconst count = ref(0)\n</script>",
		})
	} else if compositionAPI > optionsAPI && compositionAPI >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "vue",
			Description: "Vue Composition API",
			Example:     "setup() { const count = ref(0); return { count } }",
		})
	} else if optionsAPI >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "vue",
			Description: "Vue Options API",
			Example:     "export default { data() { return { count: 0 } } }",
		})
	}

	if pinia >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "vue",
			Description: "Pinia for state management (Vue 3 recommended)",
		})
	} else if vuex >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "vue",
			Description: "Vuex for state management",
		})
	}

	return conventions
}

// detectAngularPatterns detects Angular-specific patterns
func (d *FrameworkDetector) detectAngularPatterns() []types.Convention {
	var conventions []types.Convention

	if !d.hasFramework("angular") {
		return conventions
	}

	standaloneComponents := 0
	moduleComponents := 0
	signals := 0
	rxjs := 0
	ngrx := 0

	standaloneRegex := regexp.MustCompile(`@Component\s*\(\s*\{[^}]*standalone\s*:\s*true`)
	moduleRegex := regexp.MustCompile(`@NgModule\s*\(`)
	componentRegex := regexp.MustCompile(`@Component\s*\(`)
	signalsRegex := regexp.MustCompile(`(?:signal|computed|effect)\s*\(`)
	rxjsRegex := regexp.MustCompile(`(?:Observable|Subject|BehaviorSubject|pipe)\s*[<(]|\.subscribe\s*\(`)
	ngrxRegex := regexp.MustCompile(`@ngrx|createAction|createReducer|createEffect|Store\s*<`)

	sampledFiles := 0
	maxSamples := 30

	for _, f := range d.files {
		if f.IsDir || sampledFiles >= maxSamples {
			continue
		}

		if f.Extension != ".ts" {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)

		if standaloneRegex.MatchString(contentStr) {
			standaloneComponents++
		}
		if moduleRegex.MatchString(contentStr) {
			moduleComponents++
		}
		if componentRegex.MatchString(contentStr) && !standaloneRegex.MatchString(contentStr) {
			moduleComponents++
		}
		if signalsRegex.MatchString(contentStr) {
			signals++
		}
		if rxjsRegex.MatchString(contentStr) {
			rxjs++
		}
		if ngrxRegex.MatchString(contentStr) {
			ngrx++
		}

		sampledFiles++
	}

	// Report findings
	if standaloneComponents >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "angular",
			Description: "Standalone components (Angular 14+ pattern)",
			Example:     "@Component({ standalone: true, imports: [...] })",
		})
	} else if moduleComponents >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "angular",
			Description: "NgModule-based architecture",
		})
	}

	if signals >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "angular",
			Description: "Angular Signals for reactive state (Angular 16+)",
			Example:     "count = signal(0); doubled = computed(() => count() * 2)",
		})
	}

	if rxjs >= 5 {
		conventions = append(conventions, types.Convention{
			Category:    "angular",
			Description: "RxJS for reactive programming",
		})
	}

	if ngrx >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "angular",
			Description: "NgRx for state management",
		})
	}

	return conventions
}

// detectNextJSPatterns detects Next.js-specific patterns
func (d *FrameworkDetector) detectNextJSPatterns() []types.Convention {
	var conventions []types.Convention

	if !d.hasFramework("next") {
		return conventions
	}

	// Check for app router vs pages router
	hasAppRouter := false
	hasPagesRouter := false
	serverComponents := 0
	clientComponents := 0
	serverActions := 0
	apiRoutes := 0

	for _, f := range d.files {
		// Check directory structure
		if strings.HasPrefix(f.Path, "app/") || strings.HasPrefix(f.Path, "src/app/") {
			hasAppRouter = true
		}
		if strings.HasPrefix(f.Path, "pages/") || strings.HasPrefix(f.Path, "src/pages/") {
			hasPagesRouter = true
		}

		// Check for API routes
		if strings.Contains(f.Path, "/api/") && (f.Name == "route.ts" || f.Name == "route.js" ||
			strings.HasSuffix(f.Name, ".ts") || strings.HasSuffix(f.Name, ".js")) {
			apiRoutes++
		}
	}

	// Sample files for patterns
	useClientRegex := regexp.MustCompile(`['"]use client['"]`)
	useServerRegex := regexp.MustCompile(`['"]use server['"]`)
	serverActionRegex := regexp.MustCompile(`async\s+function\s+\w+.*\{[^}]*['"]use server['"]`)

	sampledFiles := 0
	maxSamples := 30

	for _, f := range d.files {
		if f.IsDir || sampledFiles >= maxSamples {
			continue
		}

		if !isReactFile(f.Extension, f.Name) {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)

		if useClientRegex.MatchString(contentStr) {
			clientComponents++
		}
		// Files without 'use client' in app router are server components
		if hasAppRouter && !useClientRegex.MatchString(contentStr) &&
			(strings.Contains(f.Path, "/app/") || strings.Contains(f.Path, "src/app/")) {
			serverComponents++
		}
		if useServerRegex.MatchString(contentStr) || serverActionRegex.MatchString(contentStr) {
			serverActions++
		}

		sampledFiles++
	}

	// Report findings
	if hasAppRouter && !hasPagesRouter {
		conventions = append(conventions, types.Convention{
			Category:    "nextjs",
			Description: "Next.js App Router (recommended for new projects)",
			Example:     "app/page.tsx, app/layout.tsx, app/api/route.ts",
		})
	} else if hasPagesRouter && !hasAppRouter {
		conventions = append(conventions, types.Convention{
			Category:    "nextjs",
			Description: "Next.js Pages Router",
			Example:     "pages/index.tsx, pages/api/hello.ts",
		})
	} else if hasAppRouter && hasPagesRouter {
		conventions = append(conventions, types.Convention{
			Category:    "nextjs",
			Description: "Next.js hybrid routing (App Router + Pages Router)",
		})
	}

	if serverComponents >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "nextjs",
			Description: "React Server Components (default in App Router)",
		})
	}

	if clientComponents >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "nextjs",
			Description: "'use client' directive for client components",
			Example:     "'use client'\\nexport default function Button() { ... }",
		})
	}

	if serverActions >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "nextjs",
			Description: "Server Actions for mutations",
			Example:     "'use server'\\nasync function createItem(formData) { ... }",
		})
	}

	if apiRoutes >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "nextjs",
			Description: "API routes for backend endpoints",
		})
	}

	return conventions
}

// detectSpringBootPatterns detects Spring Boot patterns
func (d *FrameworkDetector) detectSpringBootPatterns() []types.Convention {
	var conventions []types.Convention

	if !d.hasFramework("spring") {
		return conventions
	}

	restControllers := 0
	services := 0
	repositories := 0
	entityClasses := 0
	lombokUsage := 0
	webflux := 0

	restControllerRegex := regexp.MustCompile(`@RestController|@Controller`)
	serviceRegex := regexp.MustCompile(`@Service`)
	repositoryRegex := regexp.MustCompile(`@Repository|extends\s+(?:JpaRepository|CrudRepository|MongoRepository)`)
	entityRegex := regexp.MustCompile(`@Entity|@Document|@Table`)
	lombokRegex := regexp.MustCompile(`@(?:Data|Getter|Setter|Builder|NoArgsConstructor|AllArgsConstructor|Slf4j)`)
	webfluxRegex := regexp.MustCompile(`Mono<|Flux<|@EnableWebFlux`)

	sampledFiles := 0
	maxSamples := 40

	for _, f := range d.files {
		if f.IsDir || sampledFiles >= maxSamples {
			continue
		}

		if f.Extension != ".java" && f.Extension != ".kt" {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)

		if restControllerRegex.MatchString(contentStr) {
			restControllers++
		}
		if serviceRegex.MatchString(contentStr) {
			services++
		}
		if repositoryRegex.MatchString(contentStr) {
			repositories++
		}
		if entityRegex.MatchString(contentStr) {
			entityClasses++
		}
		if lombokRegex.MatchString(contentStr) {
			lombokUsage++
		}
		if webfluxRegex.MatchString(contentStr) {
			webflux++
		}

		sampledFiles++
	}

	// Report findings
	if restControllers >= 2 && services >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "spring",
			Description: "Layered architecture: Controller → Service → Repository",
		})
	}

	if restControllers >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "spring",
			Description: "@RestController with @RequestMapping for REST APIs",
			Example:     "@RestController\\n@RequestMapping(\"/api/users\")",
		})
	}

	if repositories >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "spring",
			Description: "Spring Data JPA repositories for data access",
		})
	}

	if lombokUsage >= 5 {
		conventions = append(conventions, types.Convention{
			Category:    "spring",
			Description: "Lombok for reducing boilerplate (@Data, @Builder, etc.)",
		})
	}

	if webflux >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "spring",
			Description: "Spring WebFlux for reactive programming (Mono/Flux)",
		})
	}

	return conventions
}

// detectNodeBackendPatterns detects Express/Fastify patterns
func (d *FrameworkDetector) detectNodeBackendPatterns() []types.Convention {
	var conventions []types.Convention

	hasExpress := d.hasFramework("express")
	hasFastify := d.hasFramework("fastify")
	hasNest := d.hasFramework("nest")

	if !hasExpress && !hasFastify && !hasNest {
		return conventions
	}

	middlewareCount := 0
	routerCount := 0
	controllerCount := 0

	expressRouterRegex := regexp.MustCompile(`(?:express\.)?Router\(\)|app\.(?:get|post|put|delete|patch)\s*\(`)
	fastifyRouteRegex := regexp.MustCompile(`fastify\.(?:get|post|put|delete|patch)\s*\(|\.route\s*\(`)
	middlewareRegex := regexp.MustCompile(`app\.use\s*\(|\.use\s*\(`)
	nestControllerRegex := regexp.MustCompile(`@Controller\s*\(`)

	sampledFiles := 0
	maxSamples := 30

	for _, f := range d.files {
		if f.IsDir || sampledFiles >= maxSamples {
			continue
		}

		if f.Extension != ".js" && f.Extension != ".ts" {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)

		if expressRouterRegex.MatchString(contentStr) || fastifyRouteRegex.MatchString(contentStr) {
			routerCount++
		}
		if middlewareRegex.MatchString(contentStr) {
			middlewareCount++
		}
		if nestControllerRegex.MatchString(contentStr) {
			controllerCount++
		}

		sampledFiles++
	}

	// Report findings
	if hasNest && controllerCount >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "nestjs",
			Description: "NestJS with decorators (@Controller, @Injectable)",
			Example:     "@Controller('users')\\nexport class UsersController { ... }",
		})
	}

	if hasExpress && routerCount >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "express",
			Description: "Express.js route handlers",
			Example:     "app.get('/api/users', (req, res) => { ... })",
		})
	}

	if hasFastify && routerCount >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "fastify",
			Description: "Fastify route handlers",
			Example:     "fastify.get('/api/users', async (request, reply) => { ... })",
		})
	}

	if middlewareCount >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "node-backend",
			Description: "Middleware pattern for request processing",
		})
	}

	return conventions
}

// detectPythonBackendPatterns detects Django/FastAPI/Flask patterns
func (d *FrameworkDetector) detectPythonBackendPatterns() []types.Convention {
	var conventions []types.Convention

	hasDjango := d.hasFramework("django")
	hasFastAPI := d.hasFramework("fastapi")
	hasFlask := d.hasFramework("flask")

	if !hasDjango && !hasFastAPI && !hasFlask {
		return conventions
	}

	fastapiRoutes := 0
	djangoViews := 0
	flaskRoutes := 0
	pydanticModels := 0

	fastapiRegex := regexp.MustCompile(`@app\.(?:get|post|put|delete|patch)\s*\(|@router\.(?:get|post|put|delete|patch)\s*\(`)
	djangoViewRegex := regexp.MustCompile(`class\s+\w+\s*\(\s*(?:APIView|ViewSet|ModelViewSet|View)\s*\)`)
	flaskRegex := regexp.MustCompile(`@app\.route\s*\(|@blueprint\.route\s*\(`)
	pydanticRegex := regexp.MustCompile(`class\s+\w+\s*\(\s*(?:BaseModel|BaseSettings)\s*\)`)

	sampledFiles := 0
	maxSamples := 30

	for _, f := range d.files {
		if f.IsDir || sampledFiles >= maxSamples {
			continue
		}

		if f.Extension != ".py" {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)

		if fastapiRegex.MatchString(contentStr) {
			fastapiRoutes++
		}
		if djangoViewRegex.MatchString(contentStr) {
			djangoViews++
		}
		if flaskRegex.MatchString(contentStr) {
			flaskRoutes++
		}
		if pydanticRegex.MatchString(contentStr) {
			pydanticModels++
		}

		sampledFiles++
	}

	// Report findings
	if hasFastAPI && fastapiRoutes >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "fastapi",
			Description: "FastAPI route decorators with type hints",
			Example:     "@app.get('/users/{user_id}')\\nasync def get_user(user_id: int): ...",
		})
	}

	if hasDjango && djangoViews >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "django",
			Description: "Django class-based views / DRF ViewSets",
		})
	}

	if hasFlask && flaskRoutes >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "flask",
			Description: "Flask route decorators",
			Example:     "@app.route('/users', methods=['GET'])",
		})
	}

	if pydanticModels >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "python",
			Description: "Pydantic models for data validation",
			Example:     "class User(BaseModel):\\n    name: str\\n    email: EmailStr",
		})
	}

	return conventions
}

// detectGoWebPatterns detects Go web framework patterns
func (d *FrameworkDetector) detectGoWebPatterns() []types.Convention {
	var conventions []types.Convention

	hasGin := d.hasFramework("gin")
	hasEcho := d.hasFramework("echo")
	hasFiber := d.hasFramework("fiber")
	hasChi := d.hasFramework("chi")

	if !hasGin && !hasEcho && !hasFiber && !hasChi {
		return conventions
	}

	ginRoutes := 0
	echoRoutes := 0
	fiberRoutes := 0
	chiRoutes := 0

	ginRegex := regexp.MustCompile(`\.(?:GET|POST|PUT|DELETE|PATCH)\s*\(|gin\.Context`)
	echoRegex := regexp.MustCompile(`e\.(?:GET|POST|PUT|DELETE|PATCH)\s*\(|echo\.Context`)
	fiberRegex := regexp.MustCompile(`app\.(?:Get|Post|Put|Delete|Patch)\s*\(|\*fiber\.Ctx`)
	chiRegex := regexp.MustCompile(`r\.(?:Get|Post|Put|Delete|Patch)\s*\(|chi\.Router`)

	sampledFiles := 0
	maxSamples := 30

	for _, f := range d.files {
		if f.IsDir || sampledFiles >= maxSamples {
			continue
		}

		if f.Extension != ".go" {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil || len(content) > 500000 {
			continue
		}

		contentStr := string(content)

		if ginRegex.MatchString(contentStr) {
			ginRoutes++
		}
		if echoRegex.MatchString(contentStr) {
			echoRoutes++
		}
		if fiberRegex.MatchString(contentStr) {
			fiberRoutes++
		}
		if chiRegex.MatchString(contentStr) {
			chiRoutes++
		}

		sampledFiles++
	}

	// Report findings
	if hasGin && ginRoutes >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "gin",
			Description: "Gin HTTP handlers with gin.Context",
			Example:     "r.GET(\"/users/:id\", func(c *gin.Context) { ... })",
		})
	}

	if hasEcho && echoRoutes >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "echo",
			Description: "Echo HTTP handlers",
			Example:     "e.GET(\"/users/:id\", getUser)",
		})
	}

	if hasFiber && fiberRoutes >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "fiber",
			Description: "Fiber HTTP handlers (Express-like)",
			Example:     "app.Get(\"/users/:id\", func(c *fiber.Ctx) error { ... })",
		})
	}

	if hasChi && chiRoutes >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "chi",
			Description: "Chi router with middleware support",
			Example:     "r.Get(\"/users/{id}\", getUser)",
		})
	}

	return conventions
}

// hasFramework checks if a framework is detected in the project
func (d *FrameworkDetector) hasFramework(name string) bool {
	name = strings.ToLower(name)

	// Check package.json for JS frameworks
	pkgPath := filepath.Join(d.rootPath, "package.json")
	if content, err := os.ReadFile(pkgPath); err == nil {
		contentLower := strings.ToLower(string(content))
		switch name {
		case "react":
			return strings.Contains(contentLower, "\"react\"")
		case "vue":
			return strings.Contains(contentLower, "\"vue\"")
		case "angular":
			return strings.Contains(contentLower, "\"@angular/core\"")
		case "next":
			return strings.Contains(contentLower, "\"next\"")
		case "express":
			return strings.Contains(contentLower, "\"express\"")
		case "fastify":
			return strings.Contains(contentLower, "\"fastify\"")
		case "nest":
			return strings.Contains(contentLower, "\"@nestjs/core\"")
		}
	}

	// Check go.mod for Go frameworks
	modPath := filepath.Join(d.rootPath, "go.mod")
	if content, err := os.ReadFile(modPath); err == nil {
		contentStr := string(content)
		switch name {
		case "gin":
			return strings.Contains(contentStr, "github.com/gin-gonic/gin")
		case "echo":
			return strings.Contains(contentStr, "github.com/labstack/echo")
		case "fiber":
			return strings.Contains(contentStr, "github.com/gofiber/fiber")
		case "chi":
			return strings.Contains(contentStr, "github.com/go-chi/chi")
		}
	}

	// Check requirements.txt / pyproject.toml for Python frameworks
	reqPath := filepath.Join(d.rootPath, "requirements.txt")
	pyprojectPath := filepath.Join(d.rootPath, "pyproject.toml")

	var pythonContent string
	if content, err := os.ReadFile(reqPath); err == nil {
		pythonContent += strings.ToLower(string(content))
	}
	if content, err := os.ReadFile(pyprojectPath); err == nil {
		pythonContent += strings.ToLower(string(content))
	}

	if pythonContent != "" {
		switch name {
		case "django":
			return strings.Contains(pythonContent, "django")
		case "fastapi":
			return strings.Contains(pythonContent, "fastapi")
		case "flask":
			return strings.Contains(pythonContent, "flask")
		}
	}

	// Check pom.xml / build.gradle for Java frameworks
	pomPath := filepath.Join(d.rootPath, "pom.xml")
	gradlePath := filepath.Join(d.rootPath, "build.gradle")

	var javaContent string
	if content, err := os.ReadFile(pomPath); err == nil {
		javaContent += strings.ToLower(string(content))
	}
	if content, err := os.ReadFile(gradlePath); err == nil {
		javaContent += strings.ToLower(string(content))
	}

	if javaContent != "" && name == "spring" {
		return strings.Contains(javaContent, "spring-boot") || strings.Contains(javaContent, "springframework")
	}

	return false
}

// isReactFile checks if a file is likely a React component file
func isReactFile(ext, name string) bool {
	if ext == ".jsx" || ext == ".tsx" {
		return true
	}
	if (ext == ".js" || ext == ".ts") && !strings.HasSuffix(name, ".test.js") &&
		!strings.HasSuffix(name, ".test.ts") && !strings.HasSuffix(name, ".spec.js") &&
		!strings.HasSuffix(name, ".spec.ts") {
		return true
	}
	return false
}
