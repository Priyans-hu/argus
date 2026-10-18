package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Priyans-hu/argus/internal/analyzer"
	"github.com/Priyans-hu/argus/internal/config"
	"github.com/Priyans-hu/argus/internal/merger"
	"github.com/Priyans-hu/argus/pkg/types"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// determinismFixture holds inputs that made output change between runs:
// equal Go and TypeScript file counts, overlapping import prefixes
// (next/router, next/navigation), logger patterns that match the same calls
// (winston, pino), a dependency that matches two notable patterns
// (gorm.io/driver/postgres), several path aliases, and many dependencies
// and scripts.
var determinismFixture = map[string]string{
	"package.json": `{
  "name": "shop",
  "scripts": {"dev": "next dev", "build": "next build", "start": "next start", "lint": "eslint .", "test": "vitest", "format": "prettier --write ."},
  "dependencies": {"next": "14.1.0", "react": "18.2.0", "react-dom": "18.2.0", "@prisma/client": "5.8.0", "zustand": "4.5.0", "winston": "3.11.0", "pino": "8.17.0", "axios": "1.6.0", "zod": "3.22.0"},
  "devDependencies": {"typescript": "5.3.0", "vitest": "1.2.0", "eslint": "8.56.0", "prettier": "3.2.0", "prisma": "5.8.0"}
}`,
	"tsconfig.json": `{"compilerOptions": {"strict": true, "paths": {"@/*": ["./src/*"], "@components/*": ["./src/components/*"], "~/*": ["./src/*"], "#lib/*": ["./lib/*"]}}}`,
	"Makefile":      "build:\n\tgo build ./...\n\ntest:\n\tgo test ./...\n\nlint:\n\tgolangci-lint run\n",
	"go.mod":        "module example.com/shop\n\ngo 1.22\n\nrequire (\n\tgorm.io/driver/postgres v1.5.0\n\tgorm.io/gorm v1.25.0\n)\n",
	"README.md":     "# Shop\n\nAn online shop with a Go API and a Next.js front end.\n\n## Features\n\n- Cart\n- Checkout\n",

	"cmd/shop/main.go": "package main\n\nimport \"example.com/shop/internal/api\"\n\nfunc main() {\n\tif err := api.Serve(); err != nil {\n\t\tpanic(err)\n\t}\n}\n",
	"internal/db/db.go": "package db\n\nimport (\n\t\"gorm.io/driver/postgres\"\n\t\"gorm.io/gorm\"\n)\n\n" +
		"func Open(dsn string) (*gorm.DB, error) {\n\tdb, err := gorm.Open(postgres.Open(dsn))\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn db, nil\n}\n",
	"internal/api/users.go": "package api\n\nimport (\n\t\"fmt\"\n\t\"net/http\"\n)\n\n" +
		"func Serve() error {\n\thttp.HandleFunc(\"/users\", users)\n\tif err := http.ListenAndServe(\":8080\", nil); err != nil {\n\t\treturn fmt.Errorf(\"serve: %w\", err)\n\t}\n\treturn nil\n}\n\n" +
		"func users(w http.ResponseWriter, r *http.Request) {}\n",
	"internal/api/users_test.go": "package api\n\nimport \"testing\"\n\nfunc TestServe(t *testing.T) {\n\tt.Run(\"users\", func(t *testing.T) {})\n}\n",

	"src/app/page.tsx": "import { useRouter } from \"next/navigation\";\nimport { logger } from \"../lib/log\";\nvar navigation = require(\"next/navigation\");\n\n" +
		"export default function Page() {\n  const router = useRouter();\n  logger.info(\"page\");\n  return <main onClick={() => router.push(\"/cart\")}>Shop</main>;\n}\n",
	"src/pages/legacy.tsx": "import { useRouter } from \"next/router\";\nimport { logger } from \"../lib/log\";\nvar router = require(\"next/router\");\n\n" +
		"export default function Legacy() {\n  const r = useRouter();\n  logger.info(\"legacy\");\n  return <div>{r.pathname}</div>;\n}\n",
	"src/lib/log.ts": "import winston from \"winston\";\nimport pino from \"pino\";\nvar w = require(\"winston\");\nvar p = require(\"pino\");\n\n" +
		"export const logger = winston.createLogger({});\nexport const fast = pino();\nlogger.info(\"started\");\nfast.info(\"started\");\n",
	"src/lib/cart.test.ts": "import { describe, it, expect } from \"vitest\";\n\ndescribe(\"cart\", () => {\n  it(\"adds\", () => {\n    expect(1 + 1).toBe(2);\n  });\n});\n",

	"scripts/api.py": "from fastapi import FastAPI\n\napp = FastAPI()\n\n\n@app.get(\"/health\")\ndef health():\n    return {\"ok\": True}\n",
}

// writeDeterminismFixture writes the fixture into a fresh directory named
// shop (so every copy has the same project name) and makes it a git repo.
func writeDeterminismFixture(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "shop")
	writeFiles(t, root, determinismFixture)
	initFixtureRepo(t, root)
	return root
}

// initFixtureRepo commits the fixture with scoped and unscoped messages and
// adds feat/ and fix/ branches and an origin remote that uses an SSH alias.
func initFixtureRepo(t *testing.T, root string) {
	t.Helper()
	repo, err := git.PlainInit(root, false)
	if err != nil {
		t.Fatal(err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if err := wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		t.Fatal(err)
	}
	sig := &object.Signature{Name: "Dev", Email: "dev@example.com", When: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
	var head plumbing.Hash
	for _, msg := range []string{
		"feat(api): add users endpoint", "fix(db): close idle connections", "feat: add cart",
		"docs: describe setup", "chore(deps): bump gorm", "feat(web)!: move to the app router",
	} {
		head, err = wt.Commit(msg, &git.CommitOptions{Author: sig, Committer: sig, AllowEmptyCommits: true})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, branch := range []string{"feat/cart", "feat/search", "fix/login", "fix/typo"} {
		ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branch), head)
		if err := repo.Storer.SetReference(ref); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{"git@github-personal:o/shop.git"}}); err != nil {
		t.Fatal(err)
	}
}

var allFormats = []string{"claude", "claude-code", "cursor", "copilot", "continue"}

// analyze runs the parallel or the sequential analyzer on root.
func analyze(t *testing.T, root string, useParallel bool) *types.Analysis {
	t.Helper()
	var (
		analysis *types.Analysis
		err      error
	)
	if useParallel {
		analysis, err = analyzer.NewParallelAnalyzer(root, nil).Analyze(context.Background())
	} else {
		analysis, err = analyzer.NewAnalyzer(root, nil).Analyze(context.Background())
	}
	if err != nil {
		t.Fatal(err)
	}
	return analysis
}

// writeFormats renders formats from analysis into dir with the default config.
func writeFormats(t *testing.T, dir string, analysis *types.Analysis, compact bool, formats ...string) {
	t.Helper()
	for _, format := range formats {
		if err := generateOutput(dir, "", format, analysis, config.DefaultConfig(), false, compact); err != nil {
			t.Fatalf("%s: %v", format, err)
		}
	}
}

// autoSpan returns the ARGUS:AUTO block of a merge file, markers included.
func autoSpan(t *testing.T, content string) string {
	t.Helper()
	start := strings.Index(content, merger.AutoStartMarker)
	end := strings.Index(content, merger.AutoEndMarker)
	if start < 0 || end < start {
		t.Fatalf("no ARGUS:AUTO block in:\n%s", content)
	}
	return content[start : end+len(merger.AutoEndMarker)]
}

func TestOutputIsDeterministic(t *testing.T) {
	runs := 10
	if testing.Short() {
		runs = 3
	}
	resetFlags(t)
	captureOutput(t)
	root := writeDeterminismFixture(t)

	for _, mode := range []struct {
		name        string
		useParallel bool
	}{{"parallel", true}, {"sequential", false}} {
		t.Run(mode.name, func(t *testing.T) {
			var first map[string]string
			for i := range runs {
				out := t.TempDir()
				analysis := analyze(t, root, mode.useParallel)
				writeFormats(t, filepath.Join(out, "full"), analysis, false, allFormats...)
				writeFormats(t, filepath.Join(out, "compact"), analysis, true, "claude")

				tree := snapshotTree(t, out)
				if i == 0 {
					first = tree
					continue
				}
				if diff := treeDiff(first, tree); len(diff) > 0 {
					t.Fatalf("run %d differs from run 0 in %v", i, diff)
				}
			}
		})
	}
}

func TestRescanIsIdempotent(t *testing.T) {
	resetFlags(t)
	captureOutput(t)
	root := writeDeterminismFixture(t)

	writeFormats(t, root, analyze(t, root, true), false, allFormats...)
	first := snapshotTree(t, root)
	for i := 1; i <= 2; i++ {
		// the outputs are now part of the tree; the walker must ignore them
		writeFormats(t, root, analyze(t, root, true), false, allFormats...)
		if diff := treeDiff(first, snapshotTree(t, root)); len(diff) > 0 {
			t.Fatalf("rescan %d changed %v", i, diff)
		}
	}
	if len(skipped) > 0 {
		t.Errorf("rescans skipped %v", skipped)
	}

	clean := writeDeterminismFixture(t)
	writeFormats(t, clean, analyze(t, clean, true), false, "claude")
	if got, want := autoSpan(t, first["CLAUDE.md"]), autoSpan(t, readFile(t, clean, "CLAUDE.md")); got != want {
		t.Errorf("rescanned CLAUDE.md differs from a clean scan:\n%s\n--- clean ---\n%s", got, want)
	}
}

func TestMonorepoRescanIsIdempotent(t *testing.T) {
	resetFlags(t)
	captureOutput(t)
	root := monorepoFixture(t)
	custom := merger.CustomStartMarker + "\nKEEP-ROOT-CUSTOM\n" + merger.CustomEndMarker + "\n"
	writeFiles(t, root, map[string]string{"CLAUDE.md": "# Team notes\n\n" + custom})

	var trees []map[string]string
	for range 3 {
		scanMonorepo(t, root, allFormats...)
		trees = append(trees, snapshotTree(t, root))
	}
	for i := 1; i < len(trees); i++ {
		if diff := treeDiff(trees[0], trees[i]); len(diff) > 0 {
			t.Errorf("run %d changed %v", i+1, diff)
		}
	}

	rootFile := trees[len(trees)-1]["CLAUDE.md"]
	checkMarked(t, rootFile)
	if !strings.Contains(rootFile, custom) {
		t.Errorf("root CLAUDE.md lost its CUSTOM block:\n%s", rootFile)
	}
	if len(skipped) > 0 {
		t.Errorf("rescans skipped %v", skipped)
	}
}
