// Package processor handles the concatenation of project files into a single document
// while respecting ignore patterns and preserving file structure. It's responsible for
// traversing directories, filtering files, and assembling the final output document
// with clear separation between files and maintaining context about the project structure.
package processor

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/holonoms/sandworm/internal/filetree"
)

const separator = "================================================================================"

// extraIgnores defines patterns for files that should typically be ignored
const extraIgnores = `
# === Non-binary files that are typically committed but irrelevant
# === for LLMs assistance (e.g. logs, package lock files, etc.)
.sandworm
.sandwormignore
.sandworm*.txt
.git*
CHANGELOG*
*LICENSE*
*.lock
*-lock.json
*-lock.yaml
go.sum
*.log

# === Binary files
# Image files
*.png
*.jpg
*.jpeg
*.gif
*.bmp
*.ico
*.webp

# Document files
*.pdf
*.doc
*.docx
*.xls
*.xlsx
*.ppt
*.pptx

# Archive files
*.zip
*.tar
*.gz
*.7z
*.rar

# Executable and library files
*.exe
*.dll
*.so
*.dylib

# Media files
*.mp3
*.mp4
*.avi
*.mov
*.wav

# Font files
*.ttf
*.otf
*.woff
*.woff2

# Generic binary files
*.bin
`

// Processor handles the concatenation of project files into a single document
type Processor struct {
	rootDir          string
	outputFile       string
	ignoreFile       string
	matcher          gitignore.Matcher
	printLineNumbers bool
}

// New creates a new Processor instance
func New(rootDir, outputFile, ignoreFile string, printLineNumbers bool) (*Processor, error) {
	rootDir = filepath.Clean(rootDir)

	p := &Processor{
		rootDir:          rootDir,
		outputFile:       outputFile,
		ignoreFile:       ignoreFile,
		printLineNumbers: printLineNumbers,
	}

	// Initialize patterns with EXTRA_IGNORES
	patterns := []gitignore.Pattern{}

	// Add patterns from extraIgnores when no specific ignore file is provided
	// or when using standard ignore files
	addExtraIgnores := ignoreFile == "" ||
		filepath.Base(ignoreFile) == ".gitignore" ||
		filepath.Base(ignoreFile) == ".sandwormignore"

	if addExtraIgnores {
		scanner := bufio.NewScanner(strings.NewReader(extraIgnores))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
				continue
			}
			pattern := gitignore.ParsePattern(line, []string{})
			patterns = append(patterns, pattern)
		}
	}

	// If no specific ignore file is provided, look for .sandwormignore first,
	// then fall back to .gitignore
	if ignoreFile == "" {
		sandwormIgnore := filepath.Join(rootDir, ".sandwormignore")
		if _, err := os.Stat(sandwormIgnore); err == nil {
			p.ignoreFile = sandwormIgnore
		} else {
			gitIgnore := filepath.Join(rootDir, ".gitignore")
			if _, err := os.Stat(gitIgnore); err == nil {
				p.ignoreFile = gitIgnore
			}
		}
	}

	// Add patterns from the ignore file if it exists
	if p.ignoreFile != "" {
		data, err := os.ReadFile(p.ignoreFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read ignore file: %w", err)
		}

		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
				continue
			}
			pattern := gitignore.ParsePattern(line, []string{})
			patterns = append(patterns, pattern)
		}
	}

	// Always ignore the output file
	pattern := gitignore.ParsePattern(p.outputFile, []string{})
	patterns = append(patterns, pattern)

	p.matcher = gitignore.NewMatcher(patterns)
	return p, nil
}

// Process concatenates all project files into a single document
func (p *Processor) Process() (int64, error) {
	files, err := p.collectFiles()
	if err != nil {
		return 0, fmt.Errorf("failed to collect files: %w", err)
	}

	out, err := os.Create(p.outputFile)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = out.Close() }()

	w := bufio.NewWriter(out)

	// Write project structure
	if err := p.writeStructure(w, files); err != nil {
		return 0, fmt.Errorf("failed to write structure: %w", err)
	}

	// Write file contents
	if err := p.writeContents(w, files); err != nil {
		return 0, fmt.Errorf("failed to write contents: %w", err)
	}

	if err := w.Flush(); err != nil {
		return 0, fmt.Errorf("failed to flush writer: %w", err)
	}

	// Get final file size
	info, err := out.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to get file stats: %w", err)
	}

	return info.Size(), nil
}

// collectFiles walks the directory tree and returns a list of files to include
func (p *Processor) collectFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(p.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}
		// Get relative path and normalize separators for cross-platform consistency
		relPath, err := filepath.Rel(p.rootDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Normalize to forward slashes for consistent processing
		// This ensures gitignore patterns work and output is uniform across platforms
		normalizedPath := filepath.ToSlash(relPath)
		// Check gitignore patterns using normalized path
		if p.matcher != nil && p.matcher.Match(strings.Split(normalizedPath, "/"), false) {
			return nil
		}

		// Store normalized path for consistent cross-platform output
		files = append(files, normalizedPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// writeStructure writes the directory tree structure to the output.
func (p *Processor) writeStructure(w *bufio.Writer, files []string) error {
	_, err := w.WriteString("PROJECT STRUCTURE:\n==================\n\n")
	if err != nil {
		return err
	}

	tree := filetree.Build(files, "")
	_, err = w.WriteString(tree)
	if err != nil {
		return err
	}

	_, err = w.WriteString("\n\nFILE CONTENTS:\n==============\n\n")
	return err
}

// writeContents writes the contents of each file to the output.
func (p *Processor) writeContents(w *bufio.Writer, files []string) error {
	for _, file := range files {
		// Write file header
		if _, err := fmt.Fprintf(w, "%s\nFILE: %s\n%s\n", separator, file, separator); err != nil {
			return err
		}

		// Read file contents
		content, err := os.ReadFile(filepath.Join(p.rootDir, file))
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file, err)
		}

		// Write file contents with optional line numbers
		if p.printLineNumbers {
			if err := p.writeContentWithLineNumbers(w, content); err != nil {
				return err
			}
		} else {
			if _, err := w.Write(content); err != nil {
				return err
			}
		}

		if _, err := w.WriteString("\n"); err != nil {
			return err
		}
	}
	return nil
}

// writeContentWithLineNumbers writes file content with line numbers
func (p *Processor) writeContentWithLineNumbers(w *bufio.Writer, content []byte) error {
	lines := strings.Split(string(content), "\n")

	// Calculate the number of digits needed for the largest line number
	numLines := len(lines)
	if numLines == 0 {
		return nil
	}

	// Calculate padding based on the number of lines to dynamically adjust
	// the width of the line numbers.
	padding := int(math.Log10(float64(numLines)))
	formatStr := fmt.Sprintf("%%%dd: %%s\n", (padding + 1))

	for i, line := range lines {
		lineNum := i + 1
		formattedLine := fmt.Sprintf(formatStr, lineNum, line)
		if _, err := w.WriteString(formattedLine); err != nil {
			return err
		}
	}

	return nil
}
