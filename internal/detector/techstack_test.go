package detector

import (
	"reflect"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func detectStack(t *testing.T, files map[string]string) *types.TechStack {
	t.Helper()
	root, infos := writeProject(t, files)
	return stableResult(t, func() *types.TechStack {
		stack, err := NewTechStackDetector(root, infos).Detect()
		if err != nil {
			t.Fatal(err)
		}
		return stack
	})
}

func TestTechStackLanguagesTieBreakByName(t *testing.T) {
	stack := detectStack(t, map[string]string{
		"a.ts": "", "b.ts": "", "main.go": "", "util.go": "", "tool.py": "",
	})

	want := []types.Language{{Name: "Go", Percentage: 40}, {Name: "TypeScript", Percentage: 40}, {Name: "Python", Percentage: 20}}
	if !reflect.DeepEqual(stack.Languages, want) {
		t.Errorf("Languages = %+v, want %+v", stack.Languages, want)
	}
}

func TestTechStackPackageJSONIsStable(t *testing.T) {
	stack := detectStack(t, map[string]string{
		"package.json": `{
			"dependencies": {"next": "14.0.0", "react": "^18.2.0", "zustand": "4.0.0", "prisma": "^5.0.0", "@prisma/client": "^5.1.0", "pg": "8.0.0", "redis": "4.0.0"},
			"devDependencies": {"typescript": "5.3.0", "vitest": "1.0.0", "jest": "29.0.0", "eslint": "8.0.0", "prettier": "3.0.0"}
		}`,
	})

	var prisma []types.Framework
	for _, fw := range stack.Frameworks {
		if fw.Name == "Prisma" {
			prisma = append(prisma, fw)
		}
	}
	if len(prisma) != 1 || prisma[0].Version != "5.1.0" {
		t.Errorf("Prisma = %+v, want one entry with the @prisma/client version 5.1.0", prisma)
	}
	if want := []string{"PostgreSQL", "Redis"}; !reflect.DeepEqual(stack.Databases, want) {
		t.Errorf("Databases = %v, want %v", stack.Databases, want)
	}
}

func TestTechStackGoAndRustFrameworksAreStable(t *testing.T) {
	stack := detectStack(t, map[string]string{
		"go.mod":     "module x\n\ngo 1.24\n\nrequire (\n\tgithub.com/spf13/cobra v1.0.0\n\tgithub.com/gin-gonic/gin v1.0.0\n\tgorm.io/gorm v1.0.0\n)\n",
		"Cargo.toml": "[dependencies]\ntokio = \"1\"\nserde = \"1\"\naxum = \"0.7\"\nclap = \"4\"\n",
	})
	if len(stack.Frameworks) != 7 {
		t.Errorf("Frameworks = %+v, want 7", stack.Frameworks)
	}
}

func TestTechStackPythonFrameworksAreDeduplicated(t *testing.T) {
	stack := detectStack(t, map[string]string{
		"requirements.txt": "fastapi==0.110\npydantic\n",
		"pyproject.toml":   "[project]\ndependencies = [\"fastapi\", \"pydantic\", \"pytest\"]\n",
	})

	var names []string
	for _, fw := range stack.Frameworks {
		names = append(names, fw.Name)
	}
	if want := []string{"FastAPI", "Pydantic", "pytest"}; !reflect.DeepEqual(names, want) {
		t.Errorf("Frameworks = %v, want %v", names, want)
	}
}
