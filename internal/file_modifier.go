package internal

import (
	"path/filepath"
	"regexp"
	"strings"
)

// CodeChange represents a single change to a file
type CodeChange struct {
	Original    string
	Modified    string
	Description string
	LineNumbers [2]int // start, end lines of change
}

// FileModifier handles code modifications
type FileModifier struct {
	commentStyles map[string]string
	funcPatterns  map[string]*regexp.Regexp
}

// NewFileModifier creates a new FileModifier instance
func NewFileModifier() *FileModifier {
	return &FileModifier{
		commentStyles: map[string]string{
			".py":   "#",
			".js":   "//",
			".ts":   "//",
			".go":   "//",
			".java": "//",
			".cpp":  "//",
			".rs":   "//",
		},
		funcPatterns: map[string]*regexp.Regexp{
			".py": regexp.MustCompile(`def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\([^)]*\):`),
			".js": regexp.MustCompile(`(function\s+[a-zA-Z_][a-zA-Z0-9_]*|\w+\s*=\s*function)\s*\([^)]*\)`),
			".ts": regexp.MustCompile(`(function\s+[a-zA-Z_][a-zA-Z0-9_]*|\w+\s*=\s*function|\w+\s*:\s*\([^)]*\)\s*=>)`),
			".go": regexp.MustCompile(`func\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\([^)]*\)`),
		},
	}
}

// FileMetadata contains analyzed file information
type FileMetadata struct {
	Lines     []string
	Functions []FunctionInfo
	Comments  [][2]int // start, end line numbers
	Variables []VariableInfo
}

type FunctionInfo struct {
	Name      string
	StartLine int
	EndLine   int
	Content   string
}

type VariableInfo struct {
	Name    string
	LineNum int
	VarType string // declaration type (var, const, let, etc.)
}

// PrepareFileContent analyzes file content and prepares metadata
func (f *FileModifier) PrepareFileContent(filePath string, content string) (*FileMetadata, error) {
	ext := filepath.Ext(filePath)
	lines := strings.Split(content, "\n")

	metadata := &FileMetadata{
		Lines: lines,
	}

	// Find functions
	if pattern, ok := f.funcPatterns[ext]; ok {
		metadata.Functions = f.findFunctions(lines, pattern)
	}

	// Find comments
	if commentStyle, ok := f.commentStyles[ext]; ok {
		metadata.Comments = f.findComments(lines, commentStyle)
	}

	// Find variables
	metadata.Variables = f.findVariables(lines, ext)

	return metadata, nil
}

// SuggestChanges generates suggested changes for the file
func (f *FileModifier) SuggestChanges(filePath string, metadata *FileMetadata) []CodeChange {
	var changes []CodeChange
	ext := filepath.Ext(filePath)

	// Randomly choose 1-2 types of changes
	changeTypes := []func(string, *FileMetadata) []CodeChange{
		f.improveComments,
		f.enhanceErrorHandling,
		f.renameVariables,
		f.addLogging,
		f.optimizeCode,
	}

	numChanges := 1
	if len(metadata.Functions) > 1 {
		numChanges = 2
	}

	for i := 0; i < numChanges; i++ {
		changeType := changeTypes[i%len(changeTypes)]
		if newChanges := changeType(ext, metadata); len(newChanges) > 0 {
			changes = append(changes, newChanges...)
		}
	}

	return changes
}

func (f *FileModifier) findFunctions(lines []string, pattern *regexp.Regexp) []FunctionInfo {
	var functions []FunctionInfo

	for i, line := range lines {
		if match := pattern.FindStringSubmatch(line); match != nil {
			// Find function end (basic implementation)
			end := i + 1
			indent := len(line) - len(strings.TrimLeft(line, " \t"))
			for end < len(lines) && (len(lines[end]) == 0 ||
				len(strings.TrimLeft(lines[end], " \t")) > indent) {
				end++
			}

			functions = append(functions, FunctionInfo{
				Name:      match[1],
				StartLine: i,
				EndLine:   end,
				Content:   strings.Join(lines[i:end], "\n"),
			})
		}
	}

	return functions
}

func (f *FileModifier) findComments(lines []string, commentStyle string) [][2]int {
	var comments [][2]int
	var currentBlock *[2]int

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, commentStyle) {
			if currentBlock == nil {
				currentBlock = &[2]int{i, i}
			}
			currentBlock[1] = i
		} else if currentBlock != nil {
			comments = append(comments, *currentBlock)
			currentBlock = nil
		}
	}

	if currentBlock != nil {
		comments = append(comments, *currentBlock)
	}

	return comments
}

func (f *FileModifier) findVariables(lines []string, ext string) []VariableInfo {
	var variables []VariableInfo
	var pattern *regexp.Regexp

	switch ext {
	case ".go":
		pattern = regexp.MustCompile(`(var|const)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*.*`)
	case ".js", ".ts":
		pattern = regexp.MustCompile(`(var|let|const)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*[=:]`)
	case ".py":
		pattern = regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*[^=]`)
	default:
		return variables
	}

	for i, line := range lines {
		if matches := pattern.FindStringSubmatch(line); len(matches) > 0 {
			varType := "var"
			name := matches[len(matches)-1]
			if len(matches) > 2 {
				varType = matches[1]
			}

			variables = append(variables, VariableInfo{
				Name:    name,
				LineNum: i,
				VarType: varType,
			})
		}
	}

	return variables
}

// Change generation methods will be added in the next message...

func (f *FileModifier) improveComments(ext string, metadata *FileMetadata) []CodeChange {
	// Implementation needed
	return nil
}

func (f *FileModifier) enhanceErrorHandling(ext string, metadata *FileMetadata) []CodeChange {
	// Implementation needed
	return nil
}

func (f *FileModifier) renameVariables(ext string, metadata *FileMetadata) []CodeChange {
	// Implementation needed
	return nil
}

func (f *FileModifier) addLogging(ext string, metadata *FileMetadata) []CodeChange {
	// Implementation needed
	return nil
}

func (f *FileModifier) optimizeCode(ext string, metadata *FileMetadata) []CodeChange {
	// Implementation needed
	return nil
}
