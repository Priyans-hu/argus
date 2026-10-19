package detector

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

// goImports runs the Go AST detector on one file importing path and returns
// "name category" per pattern.
func goImports(t *testing.T, file, path string) []string {
	t.Helper()
	root, files := writeProject(t, map[string]string{file: fmt.Sprintf("package a\n\nimport _ %q\n", path)})
	var got []string
	for _, p := range NewGoASTDetector(root, files).Detect() {
		got = append(got, p.Name+" "+p.Category)
	}
	return got
}

func TestGoLibrariesFromImports(t *testing.T) {
	tests := []struct {
		file, path string
		want       []string
	}{
		{"a.go", "github.com/gin-gonic/gin", []string{"gin go-web"}},
		{"a.go", "github.com/gin-gonic/gin/binding", []string{"gin go-web"}},
		{"a.go", "github.com/gin-gonic/gin-contrib/cors", nil},
		{"a.go", "github.com/labstack/echo/v4", []string{"echo go-web"}},
		{"a.go", "github.com/gofiber/fiber/v2", []string{"fiber go-web"}},
		{"a.go", "github.com/go-chi/chi/v5", []string{"chi go-web"}},
		{"a.go", "github.com/gorilla/mux", []string{"mux go-web"}},
		{"a.go", "gorm.io/gorm", []string{"gorm go-database"}},
		{"a.go", "github.com/jmoiron/sqlx", []string{"sqlx go-database"}},
		{"a.go", "github.com/jackc/pgx/v5", []string{"pgx go-database"}},
		{"a.go", "go.mongodb.org/mongo-driver/mongo", []string{"mongo go-database"}},
		{"a.go", "github.com/uptrace/bun", []string{"bun go-database"}},
		{"a.go", "entgo.io/ent", []string{"ent go-database"}},
		{"a.go", "github.com/spf13/cobra", []string{"cobra go-cli"}},
		{"a.go", "github.com/urfave/cli/v2", []string{"urfave/cli go-cli"}},
		{"a.go", "gopkg.in/alecthomas/kingpin.v2", []string{"kingpin go-cli"}},
		{"a.go", "github.com/spf13/pflag", []string{"pflag go-cli"}},
		{"a.go", "github.com/spf13/viper", []string{"viper go-config"}},
		{"a.go", "github.com/joho/godotenv", []string{"godotenv go-config"}},
		{"a.go", "github.com/kelseyhightower/envconfig", []string{"envconfig go-config"}},
		{"a.go", "github.com/sirupsen/logrus", []string{"logrus go-logging"}},
		{"a.go", "go.uber.org/zap", []string{"zap go-logging"}},
		{"a.go", "github.com/rs/zerolog/log", []string{"zerolog go-logging"}},
		{"a.go", "log/slog", []string{"slog go-logging"}},
		{"a.go", "github.com/go-resty/resty/v2", []string{"resty go-rest"}},
		{"a.go", "resty.dev/v3", []string{"resty go-rest"}},
		{"a.go", "github.com/golang-jwt/jwt/v5", []string{"jwt go-auth"}},
		{"a.go", "github.com/google/wire", []string{"wire go-di"}},
		{"a.go", "go.uber.org/dig", []string{"dig go-di"}},
		{"a.go", "go.uber.org/fx", []string{"fx go-di"}},
		{"a.go", "github.com/stretchr/testify/assert", []string{"testify go-testing"}},
		{"a.go", "encoding/json", []string{"json go-stdlib"}},
		// test files count only for testing libraries
		{"a_test.go", "github.com/gin-gonic/gin", nil},
		{"a_test.go", "encoding/json", nil},
		{"a_test.go", "github.com/stretchr/testify/require", []string{"testify go-testing"}},
		{"a_test.go", "github.com/golang/mock/gomock", []string{"gomock go-testing"}},
	}
	for _, tt := range tests {
		if got := goImports(t, tt.file, tt.path); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s importing %s: %v, want %v", tt.file, tt.path, got, tt.want)
		}
	}
}

func TestGoLibrariesIgnoreStringsAndComments(t *testing.T) {
	root, files := writeProject(t, map[string]string{
		"a.go": "package a\n\n// gin.Context gorm.Open( fiber.New\nvar s = \"gin.Context gorm.Open( fiber.New\"\n",
	})
	if got := NewGoASTDetector(root, files).Detect(); len(got) > 0 {
		t.Errorf("patterns from text: %+v", got)
	}
	if got := allPatterns(NewCodePatternDetector(root, files, nil).Detect()); len(got) > 0 {
		t.Errorf("keyword patterns from text: %+v", got)
	}
}

func TestGoASTResultsAreSorted(t *testing.T) {
	root, files := writeProject(t, map[string]string{
		"a.go": "package a\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\n\t\"github.com/spf13/cobra\"\n\t\"go.uber.org/zap\"\n)\n\n" +
			"func f(ctx context.Context) error {\n\t_ = cobra.Command{}\n\t_ = zap.L()\n\treturn fmt.Errorf(\"x\")\n}\n",
	})
	got := stableResult(t, func() []types.PatternInfo { return NewGoASTDetector(root, files).Detect() })
	var names []string
	for _, p := range got {
		names = append(names, p.Name)
	}
	if want := []string{"cobra", "context", "fmt.Errorf", "zap"}; !reflect.DeepEqual(names, want) {
		t.Errorf("names = %v, want %v", names, want)
	}
}
