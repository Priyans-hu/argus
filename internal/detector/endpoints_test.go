package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestShouldSkipForEndpoints(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"main_test.go", true},
		{"main.test.ts", true},
		{"handler.spec.js", true},
		{"internal/detector/endpoints.go", true},
		{"internal/analyzer/analyzer.go", true},
		{"node_modules/express/index.js", true},
		{"src/routes/api.ts", false},
		{"internal/handler/user.go", false},
		{"pkg/api/routes.go", false},
	}

	for _, tt := range tests {
		result := shouldSkipForEndpoints(tt.path)
		if result != tt.expected {
			t.Errorf("shouldSkipForEndpoints(%q) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}

func TestDetectExpressEndpoints(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "express-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create package.json with express dependency
	pkgJSON := `{"dependencies": {"express": "^4.18.0"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	// Create an Express routes file at root level
	routesContent := `const express = require('express');
const router = express.Router();

router.get('/users', getUsers);
router.post('/users', createUser);
router.get('/users/:id', getUser);
router.put('/users/:id', updateUser);
router.delete('/users/:id', deleteUser);

module.exports = router;
`
	routesPath := filepath.Join(tmpDir, "users.js")
	if err := os.WriteFile(routesPath, []byte(routesContent), 0644); err != nil {
		t.Fatalf("failed to create routes file: %v", err)
	}

	files := []types.FileInfo{
		{Path: "users.js", Name: "users.js", Extension: ".js", IsDir: false},
	}

	detector := NewEndpointDetector(tmpDir, files)
	endpoints := detector.detectExpressEndpoints()

	if len(endpoints) != 5 {
		t.Errorf("expected 5 endpoints, got %d", len(endpoints))
		for _, ep := range endpoints {
			t.Logf("  found: %s %s", ep.Method, ep.Path)
		}
	}

	// Check for specific endpoints
	hasEndpoint := func(method, path string) bool {
		for _, ep := range endpoints {
			if ep.Method == method && ep.Path == path {
				return true
			}
		}
		return false
	}

	if !hasEndpoint("GET", "/users") {
		t.Error("expected GET /users endpoint")
	}
	if !hasEndpoint("POST", "/users") {
		t.Error("expected POST /users endpoint")
	}
	if !hasEndpoint("DELETE", "/users/:id") {
		t.Error("expected DELETE /users/:id endpoint")
	}
}

func TestDetectGinEndpoints(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gin-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create go.mod with gin dependency
	goMod := `module test

go 1.21

require github.com/gin-gonic/gin v1.9.0
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create a Gin routes file
	routesContent := `package routes

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine) {
	r.GET("/api/users", GetUsers)
	r.POST("/api/users", CreateUser)
	r.GET("/api/users/:id", GetUser)
	r.PUT("/api/users/:id", UpdateUser)
	r.DELETE("/api/users/:id", DeleteUser)
}
`
	routesPath := filepath.Join(tmpDir, "internal", "routes", "routes.go")
	if err := os.MkdirAll(filepath.Dir(routesPath), 0755); err != nil {
		t.Fatalf("failed to create routes dir: %v", err)
	}
	if err := os.WriteFile(routesPath, []byte(routesContent), 0644); err != nil {
		t.Fatalf("failed to create routes file: %v", err)
	}

	files := []types.FileInfo{
		{Path: "internal/routes/routes.go", Name: "routes.go", Extension: ".go", IsDir: false},
	}

	detector := NewEndpointDetector(tmpDir, files)
	endpoints := detector.detectGinEndpoints()

	if len(endpoints) != 5 {
		t.Errorf("expected 5 endpoints, got %d", len(endpoints))
	}
}

func TestDetectFastAPIEndpoints(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fastapi-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create requirements.txt with fastapi
	reqContent := `fastapi==0.104.0
uvicorn==0.24.0
`
	if err := os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte(reqContent), 0644); err != nil {
		t.Fatalf("failed to create requirements.txt: %v", err)
	}

	// Create a FastAPI routes file
	routesContent := `
from fastapi import FastAPI, APIRouter

app = FastAPI()
router = APIRouter()

@app.get("/")
async def root():
    return {"message": "Hello World"}

@router.get("/users")
async def get_users():
    return []

@router.post("/users")
async def create_user():
    pass

@router.get("/users/{user_id}")
async def get_user(user_id: int):
    pass
`
	routesPath := filepath.Join(tmpDir, "app", "main.py")
	if err := os.MkdirAll(filepath.Dir(routesPath), 0755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}
	if err := os.WriteFile(routesPath, []byte(routesContent), 0644); err != nil {
		t.Fatalf("failed to create routes file: %v", err)
	}

	files := []types.FileInfo{
		{Path: "app/main.py", Name: "main.py", Extension: ".py", IsDir: false},
	}

	detector := NewEndpointDetector(tmpDir, files)
	endpoints := detector.detectFastAPIEndpoints()

	if len(endpoints) < 3 {
		t.Errorf("expected at least 3 endpoints, got %d", len(endpoints))
	}
}

func TestDetectChiEndpoints(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chi-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create go.mod with chi dependency
	goMod := `module test

go 1.21

require github.com/go-chi/chi/v5 v5.0.10
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create a Chi routes file
	routesContent := `package routes

import (
	"net/http"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/api/products", ListProducts)
	r.Post("/api/products", CreateProduct)
	r.Get("/api/products/{id}", GetProduct)
	r.Put("/api/products/{id}", UpdateProduct)
	r.Delete("/api/products/{id}", DeleteProduct)

	return r
}
`
	routesPath := filepath.Join(tmpDir, "internal", "routes", "routes.go")
	if err := os.MkdirAll(filepath.Dir(routesPath), 0755); err != nil {
		t.Fatalf("failed to create routes dir: %v", err)
	}
	if err := os.WriteFile(routesPath, []byte(routesContent), 0644); err != nil {
		t.Fatalf("failed to create routes file: %v", err)
	}

	files := []types.FileInfo{
		{Path: "internal/routes/routes.go", Name: "routes.go", Extension: ".go", IsDir: false},
	}

	detector := NewEndpointDetector(tmpDir, files)
	endpoints := detector.detectChiEndpoints()

	if len(endpoints) != 5 {
		t.Errorf("expected 5 endpoints, got %d", len(endpoints))
	}
}

func TestDetectEndpoints_SkipsTestFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skip-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a test file with endpoint-like patterns
	testContent := `package routes_test

import "testing"

func TestRoutes(t *testing.T) {
	r.GET("/test/users", handler)
	router.get('/test/endpoint', callback)
}
`
	testPath := filepath.Join(tmpDir, "routes_test.go")
	if err := os.WriteFile(testPath, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	files := []types.FileInfo{
		{Path: "routes_test.go", Name: "routes_test.go", Extension: ".go", IsDir: false},
	}

	detector := NewEndpointDetector(tmpDir, files)
	endpoints, _ := detector.Detect()

	// Should not detect any endpoints from test files
	if len(endpoints) > 0 {
		t.Errorf("expected 0 endpoints from test file, got %d", len(endpoints))
	}
}

func TestDetectEndpoints_SkipsDetectorFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "detector-skip-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a file that looks like it's in a detector directory
	detectorContent := `package detector

// This file contains regex patterns that look like routes
// r.Get("/api/test", handler)
// router.get('/pattern', callback)
`
	detectorPath := filepath.Join(tmpDir, "internal", "detector", "patterns.go")
	if err := os.MkdirAll(filepath.Dir(detectorPath), 0755); err != nil {
		t.Fatalf("failed to create detector dir: %v", err)
	}
	if err := os.WriteFile(detectorPath, []byte(detectorContent), 0644); err != nil {
		t.Fatalf("failed to create detector file: %v", err)
	}

	files := []types.FileInfo{
		{Path: "internal/detector/patterns.go", Name: "patterns.go", Extension: ".go", IsDir: false},
	}

	detector := NewEndpointDetector(tmpDir, files)
	endpoints, _ := detector.Detect()

	// Should not detect endpoints from detector files
	if len(endpoints) > 0 {
		t.Errorf("expected 0 endpoints from detector file, got %d", len(endpoints))
	}
}
