package generator

import (
	"reflect"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func renderClaudeCode(t *testing.T, analysis *types.Analysis) map[string]string {
	t.Helper()
	files, err := NewClaudeCodeGenerator(nil).Generate(analysis)
	if err != nil {
		t.Fatal(err)
	}
	out := make(map[string]string, len(files))
	for _, f := range files {
		out[f.Path] = string(f.Content)
	}
	return out
}

func TestClaudeCodeRenderIdenticalBytes(t *testing.T) {
	analysis := renderFixture()
	first, err := NewClaudeCodeGenerator(nil).Generate(analysis)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) == 0 {
		t.Fatal("no files generated")
	}
	for i := 1; i < 20; i++ {
		got, err := NewClaudeCodeGenerator(nil).Generate(analysis)
		if err != nil {
			t.Fatal(err)
		}
		if reflect.DeepEqual(got, first) {
			continue
		}
		if len(got) != len(first) {
			t.Fatalf("render %d: %d files, first render had %d", i, len(got), len(first))
		}
		for j := range got {
			if got[j].Path != first[j].Path || string(got[j].Content) != string(first[j].Content) {
				t.Fatalf("render %d: %s differs from the first render's %s:\n%s\n--- first ---\n%s",
					i, got[j].Path, first[j].Path, got[j].Content, first[j].Content)
			}
		}
	}
}

func TestClaudeCodePatternListsAreSortedByName(t *testing.T) {
	files := renderClaudeCode(t, renderFixture())

	testing := []string{"- **describe(** - see `web/cart.test.ts`", "- **func Test** - see `api/orders_test.go`",
		"- **httptest** - see `api/health_test.go`", "- **t.Run** - see `api/users_test.go`"}
	auth := []string{"- **Bearer** - see `api/middleware.go`", "- **jwt.** - see `api/auth.go`",
		"- **middleware** - see `api/middleware.go`"}
	api := []string{"- **OpenAPI** - see `api/openapi.yaml`", "- **REST** - see `api/users.go`",
		"- **grpc** - see `api/grpc.go`"}

	tests := []struct {
		path string
		want []string
	}{
		{".claude/agents/go-reviewer.md", testing},
		{".claude/agents/ts-reviewer.md", testing},
		{".claude/agents/python-reviewer.md", testing},
		{".claude/agents/security-reviewer.md", append(auth, api...)},
		{".claude/rules/security.md", auth},
	}
	for _, tt := range tests {
		content, ok := files[tt.path]
		if !ok {
			t.Errorf("%s was not generated", tt.path)
			continue
		}
		assertInOrder(t, content, tt.want...)
	}
}
