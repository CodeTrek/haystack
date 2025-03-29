package conf

import (
	"os"
	"path/filepath"
	"testing"

	"search-indexer/runtime"
)

func init() {
	// Set server mode for testing
	runtime.SetServerModeForTest()
}

func TestLoadAndGet(t *testing.T) {
	// Create a temporary directory for test
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "config.yaml")

	// Test data
	testYAML := `
workspaces:
  - name: workspace1
    path: /test/path1
    exclude:
      use_git_ignore: true
      customized: ["*.log", "*.tmp"]
    files: ["*.txt", "*.md"]
  - name: workspace2
    path: /test/path2
    exclude:
      use_git_ignore: false
      customized: ["*.bak"]
    files: ["*.pdf"]
port: 8080
`

	// Write test config
	if err := os.WriteFile(confPath, []byte(testYAML), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set up test environment
	cleanup := runtime.SetServerConfForTest(confPath)
	defer cleanup()

	// Test Load
	if err := Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Test Get
	c := Get()
	if c == nil {
		t.Fatal("Get returned nil")
	}

	// Verify configuration
	if c.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", c.Port)
	}

	if len(c.Workspaces) != 2 {
		t.Fatalf("Expected 2 workspaces, got %d", len(c.Workspaces))
	}

	// Check first index
	idx1 := c.Workspaces[0]
	if idx1.Path != "/test/path1" {
		t.Errorf("Expected path /test/path1, got %s", idx1.Path)
	}
	if !idx1.Exclude.UseGitIgnore {
		t.Error("Expected UseGitIgnore to be true")
	}
	if len(idx1.Exclude.Customized) != 2 {
		t.Errorf("Expected 2 customized excludes, got %d", len(idx1.Exclude.Customized))
	}
	if len(idx1.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(idx1.Files))
	}

	// Check second index
	idx2 := c.Workspaces[1]
	if idx2.Path != "/test/path2" {
		t.Errorf("Expected path /test/path2, got %s", idx2.Path)
	}
	if idx2.Exclude.UseGitIgnore {
		t.Error("Expected UseGitIgnore to be false")
	}
	if len(idx2.Exclude.Customized) != 1 {
		t.Errorf("Expected 1 customized exclude, got %d", len(idx2.Exclude.Customized))
	}
	if len(idx2.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(idx2.Files))
	}
}

func TestLoadEmptyConfig(t *testing.T) {
	// Create a temporary directory for test
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "config.yaml")

	// Set up test environment
	cleanup := runtime.SetServerConfForTest(confPath)
	defer cleanup()

	// Test Load with empty config
	if err := Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Test Get
	c := Get()
	if c == nil {
		t.Fatal("Get returned nil")
	}

	// Verify default values
	if c.Port != runtime.DefaultListenPort() {
		t.Errorf("Expected default port %d, got %d", runtime.DefaultListenPort(), c.Port)
	}
	if len(c.Workspaces) != 0 {
		t.Errorf("Expected 0 workspaces, got %d", len(c.Workspaces))
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	// Create a temporary directory for test
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "config.yaml")

	// Write invalid YAML
	invalidYAML := `
workspaces:
  - name: workspace1
    path: /test/path1
    exclude:
      use_git_ignore: invalid
    files: ["*.txt"]
port: "not a number"
`

	if err := os.WriteFile(confPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set up test environment
	cleanup := runtime.SetServerConfForTest(confPath)
	defer cleanup()

	// Test Load with invalid YAML
	if err := Load(); err == nil {
		t.Error("Expected Load to fail with invalid YAML")
	}
}
