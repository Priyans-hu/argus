package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestFilterOutTestFiles(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "filter go test files",
			input:    []string{"main.go", "main_test.go", "handler.go", "handler_test.go"},
			expected: []string{"main.go", "handler.go"},
		},
		{
			name:     "filter js test files",
			input:    []string{"app.js", "app.test.js", "utils.js", "utils.spec.js"},
			expected: []string{"app.js", "utils.js"},
		},
		{
			name:     "filter __tests__ directory",
			input:    []string{"src/component.tsx", "src/__tests__/component.test.tsx"},
			expected: []string{"src/component.tsx"},
		},
		{
			name:     "no test files",
			input:    []string{"main.go", "handler.go", "service.go"},
			expected: []string{"main.go", "handler.go", "service.go"},
		},
		{
			name:     "all test files",
			input:    []string{"main_test.go", "handler.test.js"},
			expected: nil,
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterOutTestFiles(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("filterOutTestFiles() got %d files, expected %d", len(result), len(tt.expected))
				return
			}
			for i, f := range result {
				if f != tt.expected[i] {
					t.Errorf("filterOutTestFiles()[%d] = %s, expected %s", i, f, tt.expected[i])
				}
			}
		})
	}
}

func TestCodePatternDetector_DetectStateManagement(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-mgmt-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a React component with Redux
	reduxContent := `import React from 'react';
import { useSelector, useDispatch } from 'react-redux';

export function Counter() {
  const count = useSelector((state) => state.counter.value);
  const dispatch = useDispatch();

  return (
    <div>
      <span>{count}</span>
    </div>
  );
}
`
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "Counter.tsx"), []byte(reduxContent), 0644); err != nil {
		t.Fatalf("failed to create Counter.tsx: %v", err)
	}

	files := []types.FileInfo{
		{Path: "src/Counter.tsx", Name: "Counter.tsx", Extension: ".tsx", IsDir: false},
	}

	detector := NewCodePatternDetector(tmpDir, files)
	patterns := detector.detectStateManagement()

	// Should detect Redux patterns
	hasRedux := false
	for _, p := range patterns {
		if p.Name == "useSelector" || p.Name == "useDispatch" {
			hasRedux = true
			break
		}
	}

	if !hasRedux {
		t.Error("expected Redux state management patterns to be detected")
	}
}

func TestCodePatternDetector_DetectDataFetching(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "data-fetch-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with fetch patterns
	fetchContent := `package main

import (
	"net/http"
)

func getData() {
	resp, err := http.Get("https://api.example.com/data")
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "fetch.go"), []byte(fetchContent), 0644); err != nil {
		t.Fatalf("failed to create fetch.go: %v", err)
	}

	files := []types.FileInfo{
		{Path: "fetch.go", Name: "fetch.go", Extension: ".go", IsDir: false},
	}

	detector := NewCodePatternDetector(tmpDir, files)
	patterns := detector.detectDataFetching()

	// Should detect http.Get pattern
	hasHTTPGet := false
	for _, p := range patterns {
		if p.Name == "http.Get" {
			hasHTTPGet = true
			break
		}
	}

	if !hasHTTPGet {
		t.Error("expected http.Get data fetching pattern to be detected")
	}
}

func TestCodePatternDetector_DetectTesting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testing-pattern-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a Go test file
	testContent := `package main

import "testing"

func TestSomething(t *testing.T) {
	t.Run("subtest", func(t *testing.T) {
		if 1 != 1 {
			t.Error("math is broken")
		}
	})
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create main_test.go: %v", err)
	}

	files := []types.FileInfo{
		{Path: "main_test.go", Name: "main_test.go", Extension: ".go", IsDir: false},
	}

	detector := NewCodePatternDetector(tmpDir, files)
	patterns := detector.detectTesting()

	// Should detect Go testing patterns
	hasGoTest := false
	for _, p := range patterns {
		if p.Name == "func Test" || p.Name == "t.Run" || p.Name == "t.Error" {
			hasGoTest = true
			break
		}
	}

	if !hasGoTest {
		t.Error("expected Go testing patterns to be detected")
	}
}

func TestCodePatternDetector_DetectAuthentication(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "auth-pattern-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with JWT handling
	authContent := `package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

func ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
}
`
	authDir := filepath.Join(tmpDir, "internal", "auth")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		t.Fatalf("failed to create auth dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(authDir, "jwt.go"), []byte(authContent), 0644); err != nil {
		t.Fatalf("failed to create jwt.go: %v", err)
	}

	files := []types.FileInfo{
		{Path: "internal/auth/jwt.go", Name: "jwt.go", Extension: ".go", IsDir: false},
	}

	detector := NewCodePatternDetector(tmpDir, files)
	patterns := detector.detectAuthentication()

	// Should detect JWT pattern
	hasJWT := false
	for _, p := range patterns {
		if p.Name == "jwt." {
			hasJWT = true
			break
		}
	}

	if !hasJWT {
		t.Error("expected JWT authentication pattern to be detected")
	}
}

func TestCodePatternDetector_DetectDatabasePatterns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "db-pattern-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with GORM usage
	dbContent := `package repository

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string
}

func NewDB() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
}
`
	repoDir := filepath.Join(tmpDir, "internal", "repository")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repository dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "db.go"), []byte(dbContent), 0644); err != nil {
		t.Fatalf("failed to create db.go: %v", err)
	}

	files := []types.FileInfo{
		{Path: "internal/repository/db.go", Name: "db.go", Extension: ".go", IsDir: false},
	}

	detector := NewCodePatternDetector(tmpDir, files)
	patterns := detector.detectDatabasePatterns()

	// Should detect GORM patterns
	hasGORM := false
	for _, p := range patterns {
		if p.Name == "gorm.Open" || p.Name == "gorm.Model" {
			hasGORM = true
			break
		}
	}

	if !hasGORM {
		t.Error("expected GORM database pattern to be detected")
	}
}

func TestCodePatternDetector_DetectAPIPatterns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-pattern-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with gRPC usage
	grpcContent := `package api

import (
	"google.golang.org/grpc"
)

func NewServer() *grpc.Server {
	return grpc.NewServer()
}
`
	apiDir := filepath.Join(tmpDir, "internal", "api")
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		t.Fatalf("failed to create api dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(apiDir, "server.go"), []byte(grpcContent), 0644); err != nil {
		t.Fatalf("failed to create server.go: %v", err)
	}

	files := []types.FileInfo{
		{Path: "internal/api/server.go", Name: "server.go", Extension: ".go", IsDir: false},
	}

	detector := NewCodePatternDetector(tmpDir, files)
	patterns := detector.detectAPIPatterns()

	// Should detect gRPC pattern
	hasGRPC := false
	for _, p := range patterns {
		if p.Name == "grpc" {
			hasGRPC = true
			break
		}
	}

	if !hasGRPC {
		t.Error("expected gRPC API pattern to be detected")
	}
}

func TestKeys(t *testing.T) {
	input := map[string]string{
		"a": "value1",
		"b": "value2",
		"c": "value3",
	}

	result := keys(input)

	if len(result) != 3 {
		t.Errorf("keys() returned %d items, expected 3", len(result))
	}

	// Check all keys are present
	hasKey := func(k string) bool {
		for _, r := range result {
			if r == k {
				return true
			}
		}
		return false
	}

	for k := range input {
		if !hasKey(k) {
			t.Errorf("keys() missing key: %s", k)
		}
	}
}

func TestLimitSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		limit    int
		expected int
	}{
		{
			name:     "limit smaller than input",
			input:    []string{"a", "b", "c", "d", "e"},
			limit:    3,
			expected: 3,
		},
		{
			name:     "limit larger than input",
			input:    []string{"a", "b"},
			limit:    5,
			expected: 2,
		},
		{
			name:     "limit equals input",
			input:    []string{"a", "b", "c"},
			limit:    3,
			expected: 3,
		},
		{
			name:     "empty input",
			input:    []string{},
			limit:    3,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := limitSlice(tt.input, tt.limit)
			if len(result) != tt.expected {
				t.Errorf("limitSlice() returned %d items, expected %d", len(result), tt.expected)
			}
		})
	}
}
