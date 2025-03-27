package gitutils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// GitIgnore represents the entire gitignore system
type GitIgnore struct {
	rootPath  string
	ruleFiles map[string]*GitIgnoreRules
}

// GitIgnoreRules represents a single .gitignore file
type GitIgnoreRules struct {
	baseDir   string
	rules     []gitIgnoreRule
	isGitRoot bool
}

// gitIgnoreRule represents a single .gitignore rule
type gitIgnoreRule struct {
	Pattern     string // Original pattern
	Negated     bool   // Whether it's a negation rule (!)
	AnchoredDir bool   // Whether it's a directory rule (/)
	RootOnly    bool   // Whether it only matches root directory (starts with /)
}

// NewGitIgnoreRules creates a GitIgnoreRuleFile from a file
func NewGitIgnoreRules(filePath string) (*GitIgnoreRules, error) {
	baseDir := filepath.Dir(filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	rf, err := parseGitIgnoreFile(scanner, baseDir)
	if err != nil {
		return nil, err
	}

	return rf, nil
}

func NewGitIgnoreRulesFromString(rules string, baseDir string) (*GitIgnoreRules, error) {
	scanner := bufio.NewScanner(strings.NewReader(rules))
	return parseGitIgnoreFile(scanner, baseDir)
}

// NewGitIgnore creates a new GitIgnore system
func NewGitIgnore(rootPath string) *GitIgnore {
	ignorer := &GitIgnore{
		rootPath:  rootPath,
		ruleFiles: make(map[string]*GitIgnoreRules),
	}

	// Load root .gitignore if exists
	ignorer.loadGitIgnoreForDir(rootPath)

	return ignorer
}

// IsIgnored checks if a path should be ignored by this .gitignore file
func (f *GitIgnoreRules) IsIgnored(path string, isDir bool) bool {
	// Get the path relative to the base directory
	relPath, err := filepath.Rel(f.baseDir, path)
	if err != nil {
		return false
	}

	// Normalize path separators to forward slashes
	relPath = filepath.ToSlash(relPath)

	// Track whether currently ignored
	ignored := false

	// Apply all rules
	for _, rule := range f.rules {
		if rule.isIgnored(relPath, isDir) {
			ignored = !rule.Negated
		}
	}

	return ignored
}

// IsIgnored checks if a path should be ignored by considering all applicable .gitignore rules
func (g *GitIgnore) IsIgnored(relPath string, isDir bool) bool {
	// Convert the relative path to absolute
	absPath := filepath.Join(g.rootPath, relPath)

	// Start with the directory containing the file/dir
	dirPath := absPath
	if !isDir {
		dirPath = filepath.Dir(absPath)
	}

	/*
		// Load .gitignore files for all parent directories
		currPath := dirPath
		for currPath != g.rootPath && strings.HasPrefix(currPath, g.rootPath) {
			g.loadGitIgnoreForDir(currPath)
			currPath = filepath.Dir(currPath)
		}
		g.loadGitIgnoreForDir(g.rootPath)
	*/

	// Prepare list of directories to check, starting from most specific
	var dirsToCheck []string
	currPath := dirPath
	for currPath != g.rootPath && strings.HasPrefix(currPath, g.rootPath) {
		dirsToCheck = append(dirsToCheck, currPath)
		currPath = filepath.Dir(currPath)
	}
	// Add the root directory last (least specific)
	dirsToCheck = append(dirsToCheck, g.rootPath)

	/*
		// Check if parent directory is ignored
		if !isDir {
			parentDir := filepath.Dir(absPath)
			if parentDir != g.rootPath {
				parentRelPath, err := filepath.Rel(g.rootPath, parentDir)
				if err == nil {
					// If parent directory is ignored, files within it are also ignored
					if g.IsIgnored(parentRelPath, true) {
						return true
					}
				}
			}
		}
	*/

	// First check for negation rules (these have highest precedence)
	for _, dir := range dirsToCheck {
		if ruleFile := g.loadGitIgnoreForDir(dir); ruleFile != nil {
			// Check if there's an explicit negation rule for this file
			if ruleFile.isNegated(absPath, isDir) {
				return false
			}

			if ruleFile.isGitRoot {
				break
			}
		}
	}

	// Then check for ignore rules
	for _, dir := range dirsToCheck {
		if ruleFile, exists := g.ruleFiles[dir]; exists && ruleFile != nil {
			if ruleFile.IsIgnored(absPath, isDir) {
				return true
			}
		}
	}

	return false
}

// isIgnored checks if a path matches this rule
func (r *gitIgnoreRule) isIgnored(relPath string, isDir bool) bool {
	// If it's a directory rule but the target is not a directory, no match
	if r.AnchoredDir && !isDir {
		return false
	}

	// Clean the path for matching
	relPath = strings.TrimPrefix(relPath, "./")

	// Handle root directory rules
	if r.RootOnly {
		// For root directory rules, the path must be directly under the root
		return r.matchPattern(relPath)
	}

	// Non-root directory rules can match any subpath
	pathParts := strings.Split(relPath, "/")
	for i := range pathParts {
		subPath := strings.Join(pathParts[i:], "/")
		if r.matchPattern(subPath) {
			return true
		}
	}

	return false
}

// matchPattern checks if a path matches this rule's pattern
func (r *gitIgnoreRule) matchPattern(path string) bool {
	// Normalize pattern
	pattern := strings.TrimPrefix(r.Pattern, "./")

	// Special case for directory match
	if r.AnchoredDir && pattern == path {
		return true
	}

	// Handle ** wildcard case
	if strings.Contains(pattern, "**") {
		return r.matchWithDoubleAsterisk(pattern, path)
	}

	// Convert gitignore pattern to filepath.Match supported format
	match, err := filepath.Match(pattern, path)
	if err != nil {
		return false
	}
	return match
}

// matchWithDoubleAsterisk handles patterns containing **
func (r *gitIgnoreRule) matchWithDoubleAsterisk(pattern, path string) bool {
	// Handle special case for "**" pattern (matches everything)
	if pattern == "**" {
		return true
	}

	// Split pattern and path into segments
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	return r.matchSegments(patternParts, pathParts, 0, 0)
}

// matchSegments recursively matches path segments
func (r *gitIgnoreRule) matchSegments(pattern, path []string, patternIdx, pathIdx int) bool {
	// Base case: pattern fully matched
	if patternIdx >= len(pattern) && pathIdx >= len(path) {
		return true
	}

	// If pattern is exhausted but path isn't, no match
	if patternIdx >= len(pattern) {
		return false
	}

	// Handle **
	if pattern[patternIdx] == "**" {
		// ** can match 0 or more directories
		// Try skipping **
		if r.matchSegments(pattern, path, patternIdx+1, pathIdx) {
			return true
		}

		// If path is exhausted, can't continue matching
		if pathIdx >= len(path) {
			return false
		}

		// Try consuming one path segment and continue matching **
		return r.matchSegments(pattern, path, patternIdx, pathIdx+1)
	}

	// Path exhausted but pattern isn't, no match
	if pathIdx >= len(path) {
		return false
	}

	// Normal segment matching
	match, err := filepath.Match(pattern[patternIdx], path[pathIdx])
	if err != nil {
		return false
	}
	if match {
		return r.matchSegments(pattern, path, patternIdx+1, pathIdx+1)
	}

	return false
}

// parseGitIgnoreFile parses .gitignore rules
func parseGitIgnoreFile(scanner *bufio.Scanner, baseDir string) (*GitIgnoreRules, error) {
	var rules []gitIgnoreRule

	for scanner.Scan() {
		pattern := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}

		rules = append(rules, parseRule(pattern))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &GitIgnoreRules{
		baseDir: baseDir,
		rules:   rules,
	}, nil
}

// parseRule parses a single gitignore rule
func parseRule(pattern string) gitIgnoreRule {
	// Handle negation rule
	negated := false
	if strings.HasPrefix(pattern, "!") {
		negated = true
		pattern = pattern[1:]
	}

	// Clean the pattern
	pattern = strings.TrimSpace(pattern)

	// Check if it only matches root directory
	rootOnly := false
	if strings.HasPrefix(pattern, "/") {
		rootOnly = true
		pattern = pattern[1:]
	}

	// Check if it's a directory rule
	anchoredDir := false
	if strings.HasSuffix(pattern, "/") {
		anchoredDir = true
		pattern = strings.TrimSuffix(pattern, "/")
	}

	return gitIgnoreRule{
		Pattern:     pattern,
		Negated:     negated,
		AnchoredDir: anchoredDir,
		RootOnly:    rootOnly,
	}
}

// isNegated checks if a path has an explicit negation rule
func (f *GitIgnoreRules) isNegated(path string, isDir bool) bool {
	// Get the path relative to the base directory
	relPath, err := filepath.Rel(f.baseDir, path)
	if err != nil {
		return false
	}

	// Normalize path separators to forward slashes
	relPath = filepath.ToSlash(relPath)

	// Look specifically for negation rules
	for _, rule := range f.rules {
		if rule.Negated && rule.isIgnored(relPath, isDir) {
			return true
		}
	}

	return false
}

// loadGitIgnoreForDir loads .gitignore file for a directory if it exists
func (g *GitIgnore) loadGitIgnoreForDir(dir string) *GitIgnoreRules {
	if rf, ok := g.ruleFiles[dir]; ok {
		return rf
	}

	var rf *GitIgnoreRules

	gitIgnorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitIgnorePath); err == nil {
		rf, _ = NewGitIgnoreRules(gitIgnorePath)
	}

	if rf == nil {
		rf = &GitIgnoreRules{
			baseDir: dir,
			rules:   []gitIgnoreRule{},
		}
	}

	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		rf.isGitRoot = true
	}

	g.ruleFiles[dir] = rf
	return rf
}
