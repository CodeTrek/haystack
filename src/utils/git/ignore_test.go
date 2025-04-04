package gitutils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitIgnoreRule(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		path     string
		isDir    bool
		expected bool
	}{
		// 基本通配符测试
		{"basic wildcard", "*.log", "test.log", false, true},
		{"basic wildcard no match", "*.log", "test.txt", false, false},
		{"question mark", "test?.txt", "test1.txt", false, true},
		{"question mark no match", "test?.txt", "test.txt", false, false},

		// 目录规则测试
		{"dir rule match", "dir/", "dir", true, true},
		{"dir rule no match file", "dir/", "dir/file.txt", false, false},
		{"dir rule no match dir", "dir/", "other", true, false},

		// 根目录规则测试
		{"root rule match", "/test.txt", "test.txt", false, true},
		{"root rule no match", "/test.txt", "subdir/test.txt", false, false},

		// 双星号测试
		{"double star match", "**/test.txt", "test.txt", false, true},
		{"double star match subdir", "**/test.txt", "subdir/test.txt", false, true},
		{"double star match deep", "**/test.txt", "a/b/c/test.txt", false, true},
		{"double star no match", "**/test.txt", "other.txt", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := parseRule(tt.pattern)
			result := rule.isIgnored(tt.path, tt.isDir)
			if result != tt.expected {
				t.Errorf("rule.isIgnored(%q, %v) = %v; want %v", tt.path, tt.isDir, result, tt.expected)
			}
		})
	}
}

func TestGitIgnoreRuleFile(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "gitignore-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件和目录
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建dir目录
	dirPath := filepath.Join(tempDir, "dir")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("Failed to create dir directory: %v", err)
	}

	// 创建 .gitignore 文件
	gitignoreContent := `
# 注释行
*.log
!important.log
dir/
/test.txt
**/temp.txt
`
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// 创建规则文件
	ruleFile, err := NewGitIgnoreRules(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to create rule file: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		isDir    bool
		expected bool
	}{
		{"ignore log file", "test.log", false, true},
		{"don't ignore important log", "important.log", false, false},
		{"ignore directory", "dir", true, true},
		{"ignore root file", "test.txt", false, true},
		{"ignore temp file in any dir", "temp.txt", false, true},
		{"ignore temp file in subdir", "subdir/temp.txt", false, true},
		{"don't ignore other file", "other.txt", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath := filepath.Join(tempDir, tt.path)
			result := ruleFile.IsIgnored(fullPath, tt.isDir)
			if result != tt.expected {
				t.Errorf("ruleFile.isIgnored(%q) = %v; want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestGitIgnoreSystem(t *testing.T) {
	// 创建临时目录结构
	tempDir, err := os.MkdirTemp("", "gitignore-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建目录结构
	dirs := []string{
		"subdir1",
		"subdir1/subsubdir",
		"subdir2",
		"subdir2/ignored_dir",
		"subdir2/ignored_dir/subdir",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(tempDir, dir), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// 创建测试文件
	files := []string{
		"root_file.txt",
		"should_ignore.log",
		"subdir1/file1.txt",
		"subdir1/file2.log",
		"subdir1/subsubdir/deep_file.txt",
		"subdir1/subsubdir/should_ignore.tmp",
		"subdir2/file3.txt",
		"subdir2/should_ignore.log",
		"subdir2/ignored_dir/ignored_file.txt",
		"subdir2/ignored_dir/subdir/deep_file.txt",
	}

	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		f, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
		f.Close()
	}

	// 创建根目录 .gitignore
	rootGitIgnore := `
# Ignore log files in all directories
*.log

# Ignore the entire ignored_dir directory
ignored_dir/
`
	err = os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(rootGitIgnore), 0644)
	if err != nil {
		t.Fatalf("Failed to create root .gitignore: %v", err)
	}

	// 创建子目录 .gitignore
	subDirGitIgnore := `
# Don't ignore this specific log file
!file2.log

# Ignore tmp files
*.tmp
`
	err = os.WriteFile(filepath.Join(tempDir, "subdir1", ".gitignore"), []byte(subDirGitIgnore), 0644)
	if err != nil {
		t.Fatalf("Failed to create subdir .gitignore: %v", err)
	}

	// 创建 GitIgnore 系统
	ignorer := NewGitIgnore(tempDir)

	tests := []struct {
		name     string
		path     string
		isDir    bool
		expected bool
	}{
		// 根目录规则测试
		{"root file not ignored", "root_file.txt", false, false},
		{"root log file ignored", "should_ignore.log", false, true},

		// 子目录规则测试
		{"subdir file not ignored", "subdir1/file1.txt", false, false},
		{"subdir log file not ignored", "subdir1/file2.log", false, false},
		{"subdir tmp file ignored", "subdir1/subsubdir/should_ignore.tmp", false, true},

		// 目录规则测试
		{"ignored dir ignored", "subdir2/ignored_dir", true, true},
		{"file in ignored dir ignored", "subdir2/ignored_dir/ignored_file.txt", false, true},
		{"dir in ignored dir ignored", "subdir2/ignored_dir/subdir", true, true},
		{"deep file in ignored dir ignored", "subdir2/ignored_dir/subdir/deep_file.txt", false, true},

		// 边界情况测试
		{"non-existent file", "non_existent.txt", false, false},
		{"root directory", "", true, false},
		{"parent directory", "../test.txt", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ignorer.IsIgnored(tt.path, tt.isDir)
			if result != tt.expected {
				t.Errorf("ignorer.IsIgnored(%q, %v) = %v; want %v", tt.path, tt.isDir, result, tt.expected)
			}
		})
	}
}

func TestGitIgnoreEdgeCases(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "gitignore-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试空 .gitignore 文件
	emptyGitIgnore := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(emptyGitIgnore, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty .gitignore: %v", err)
	}

	// 测试无效的 .gitignore 文件
	invalidGitIgnore := filepath.Join(tempDir, "subdir", ".gitignore")
	if err := os.MkdirAll(filepath.Dir(invalidGitIgnore), 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	if err := os.WriteFile(invalidGitIgnore, []byte("invalid pattern [*"), 0644); err != nil {
		t.Fatalf("Failed to create invalid .gitignore: %v", err)
	}

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建 GitIgnore 系统
	ignorer := NewGitIgnore(tempDir)

	tests := []struct {
		name     string
		path     string
		isDir    bool
		expected bool
	}{
		{"empty gitignore", "test.txt", false, false},
		{"invalid gitignore", "subdir/test.txt", false, false},
		{"non-existent path", "non_existent.txt", false, false},
		{"empty path", "", false, false},
		{"root path", "/", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ignorer.IsIgnored(tt.path, tt.isDir)
			if result != tt.expected {
				t.Errorf("ignorer.IsIgnored(%q, %v) = %v; want %v", tt.path, tt.isDir, result, tt.expected)
			}
		})
	}
}
