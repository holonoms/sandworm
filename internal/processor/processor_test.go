package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessor(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "sandworm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Helper function to create test files
	createFile := func(path string, content string) {
		fullPath := filepath.Join(tmpDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0o755)
		if err != nil {
			t.Fatalf("Failed to create directories for %s: %v", path, err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0o644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	t.Run("basic file processing", func(t *testing.T) {
		// Create test files
		createFile("file1.txt", "Content 1")
		createFile("dir1/file2.txt", "Content 2")

		outputFile := filepath.Join(tmpDir, "output.txt")
		p, err := NewWithOptions(tmpDir, outputFile, "", SandwormOptions{PrintLineNumbers: false, FollowSymlinks: false})
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		size, err := p.Process()
		if err != nil {
			t.Fatalf("Process failed: %v", err)
		}

		if size == 0 {
			t.Error("Expected non-zero file size")
		}

		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		// Check for expected content
		output := string(content)
		if !strings.Contains(output, "PROJECT STRUCTURE:") {
			t.Error("Missing project structure section")
		}
		if !strings.Contains(output, "FILE CONTENTS:") {
			t.Error("Missing file contents section")
		}
		if !strings.Contains(output, "Content 1") {
			t.Error("Missing content from file1.txt")
		}
		if !strings.Contains(output, "Content 2") {
			t.Error("Missing content from file2.txt")
		}
	})

	t.Run("gitignore support", func(t *testing.T) {
		// Reset temp directory
		os.RemoveAll(tmpDir)
		os.MkdirAll(tmpDir, 0o755)

		// Create .gitignore
		createFile(".gitignore", "*.log\n/tmp/")
		createFile("test.log", "Should be ignored")
		createFile("tmp/ignore.txt", "Should be ignored")
		createFile("keep.txt", "Should be kept")

		outputFile := filepath.Join(tmpDir, "output.txt")
		p, err := NewWithOptions(tmpDir, outputFile, filepath.Join(tmpDir, ".gitignore"), SandwormOptions{PrintLineNumbers: false, FollowSymlinks: false})
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		_, err = p.Process()
		if err != nil {
			t.Fatalf("Process failed: %v", err)
		}

		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		output := string(content)
		if strings.Contains(output, "Should be ignored") {
			t.Error("Found content that should have been ignored")
		}
		if !strings.Contains(output, "Should be kept") {
			t.Error("Missing content that should have been kept")
		}
	})

	t.Run("binary file handling", func(t *testing.T) {
		// Reset temp directory
		os.RemoveAll(tmpDir)
		os.MkdirAll(tmpDir, 0o755)

		// Create a binary file
		binaryContent := []byte{0xFF, 0x00, 0xFF, 0x00}
		createFile("binary.bin", string(binaryContent))
		createFile("text.txt", "Regular text file")

		outputFile := filepath.Join(tmpDir, "output.txt")
		p, err := NewWithOptions(tmpDir, outputFile, "", SandwormOptions{PrintLineNumbers: false, FollowSymlinks: false})
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		_, err = p.Process()
		if err != nil {
			t.Fatalf("Process failed: %v", err)
		}

		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		output := string(content)
		if strings.Contains(output, "binary.bin") {
			t.Error("Binary file was not excluded")
		}
		if !strings.Contains(output, "Regular text file") {
			t.Error("Text file was incorrectly excluded")
		}
	})

	t.Run("custom ignore file", func(t *testing.T) {
		// Reset temp directory
		os.RemoveAll(tmpDir)
		os.MkdirAll(tmpDir, 0o755)

		// Create custom ignore file
		createFile("custom.ignore", "*.skip")
		createFile("test.skip", "Should be ignored")
		createFile("keep.txt", "Should be kept")

		outputFile := filepath.Join(tmpDir, "output.txt")
		p, err := NewWithOptions(tmpDir, outputFile, filepath.Join(tmpDir, "custom.ignore"), SandwormOptions{PrintLineNumbers: false, FollowSymlinks: false})
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		_, err = p.Process()
		if err != nil {
			t.Fatalf("Process failed: %v", err)
		}

		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		output := string(content)
		if strings.Contains(output, "Should be ignored") {
			t.Error("Found content that should have been ignored")
		}
		if !strings.Contains(output, "Should be kept") {
			t.Error("Missing content that should have been kept")
		}
	})

	t.Run("extra ignore patterns", func(t *testing.T) {
		// Reset temp directory
		os.RemoveAll(tmpDir)
		os.MkdirAll(tmpDir, 0o755)

		// Create test files that should be ignored
		ignoredFiles := []string{
			".sandworm",
			".sandwormignore",
			".sandworm-123456.txt",
			".gitignore",
			"CHANGELOG.md",
			"LICENSE",
			"package-lock.json",
			"error.log",
		}
		for _, file := range ignoredFiles {
			createFile(file, "This should be ignored")
		}

		// Create test files that should be included
		includedFiles := []string{
			"main.go",
			"README.md",
			"config.json",
			"src/app.js",
		}
		for _, file := range includedFiles {
			createFile(file, "This should be included")
		}

		// Process the files
		outputFile := filepath.Join(tmpDir, "output.txt")
		p, err := NewWithOptions(tmpDir, outputFile, "", SandwormOptions{PrintLineNumbers: false, FollowSymlinks: false})
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		_, err = p.Process()
		if err != nil {
			t.Fatalf("Process failed: %v", err)
		}

		// Read the output file
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}
		output := string(content)

		// Check that ignored files are not included
		for _, file := range ignoredFiles {
			if strings.Contains(output, file) {
				t.Errorf("Found ignored file in output: %s", file)
			}
		}

		// Check that other files are included
		for _, file := range includedFiles {
			if !strings.Contains(output, file) {
				t.Errorf("Missing expected file in output: %s", file)
			}
		}
	})

	t.Run("symbolic link following", func(t *testing.T) {
		// Create test files
		createFile("file1.txt", "Content 1")
		createFile("dir1/file2.txt", "Content 2")
		createFile("target/file3.txt", "Content 3")

		// Create symbolic link to directory (if supported by OS)
		symlinkDir := filepath.Join(tmpDir, "symlink_dir")
		targetDir := filepath.Join(tmpDir, "target")
		err := os.Symlink(targetDir, symlinkDir)
		if err != nil {
			t.Skipf("Symbolic links not supported on this system: %v", err)
		}

		outputFile := filepath.Join(tmpDir, "output_symlinks.txt")
		p, err := NewWithOptions(tmpDir, outputFile, "", SandwormOptions{PrintLineNumbers: false, FollowSymlinks: false})
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}
		// Test without following symlinks
		files, err := p.collectFiles()
		if err != nil {
			t.Fatalf("collectFiles failed: %v", err)
		}

		// Should not include symlinked content
		foundSymlinkedContent := false
		for _, file := range files {
			if strings.Contains(file.RelativePath, "symlink_dir/file3.txt") {
				foundSymlinkedContent = true
				break
			}
		}
		if foundSymlinkedContent {
			t.Error("Expected symlinked directory content to be excluded when not following symlinks")
		}

		// Test with following symlinks
		p.followSymlinks = true
		files, err = p.collectFiles()
		if err != nil {
			t.Fatalf("collectFiles with symlinks failed: %v", err)
		}

		// Should include symlinked content
		foundSymlinkedContent = false
		for _, file := range files {
			if strings.Contains(file.RelativePath, "symlink_dir/file3.txt") {
				foundSymlinkedContent = true
				break
			}
		}
		if !foundSymlinkedContent {
			t.Error("Expected symlinked directory content to be included when following symlinks")
		}

		// Process the files
		size, err := p.Process()
		if err != nil {
			t.Fatalf("Process with symlinks failed: %v", err)
		}

		if size == 0 {
			t.Error("Expected non-zero file size with symlinks")
		}
	})

	t.Run("symbolic link cycle prevention", func(t *testing.T) {
		// Create directories that will have circular symlinks
		dir1 := filepath.Join(tmpDir, "dir1")
		dir2 := filepath.Join(tmpDir, "dir2")
		err := os.MkdirAll(dir1, 0o755)
		if err != nil {
			t.Fatalf("Failed to create dir1: %v", err)
		}
		err = os.MkdirAll(dir2, 0o755)
		if err != nil {
			t.Fatalf("Failed to create dir2: %v", err)
		}

		// Create a file in each directory
		createFile("dir1/file1.txt", "Dir 1 content")
		createFile("dir2/file2.txt", "Dir 2 content")

		// Create circular symlinks
		symlink1 := filepath.Join(dir1, "link_to_dir2")
		symlink2 := filepath.Join(dir2, "link_to_dir1")

		err = os.Symlink(dir2, symlink1)
		if err != nil {
			t.Skipf("Symbolic links not supported on this system: %v", err)
		}
		err = os.Symlink(dir1, symlink2)
		if err != nil {
			t.Skipf("Symbolic links not supported on this system: %v", err)
		}

		outputFile := filepath.Join(tmpDir, "output_cycles.txt")
		p, err := NewWithOptions(tmpDir, outputFile, "", SandwormOptions{PrintLineNumbers: false, FollowSymlinks: true})
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		p.SetFollowSymlinks(true)

		// This should not hang or crash due to infinite recursion
		files, err := p.collectFiles()
		if err != nil {
			t.Fatalf("collectFiles with cycles failed: %v", err)
		}
		// Should include files from both directories but handle cycles gracefully
		foundFile1 := false
		foundFile2 := false
		for _, file := range files {
			if strings.Contains(file.RelativePath, "dir1/file1.txt") {
				foundFile1 = true
			}
			if strings.Contains(file.RelativePath, "dir2/file2.txt") {
				foundFile2 = true
			}
		}

		if !foundFile1 {
			t.Error("Expected to find file1.txt from dir1 directory")
		}
		if !foundFile2 {
			t.Error("Expected to find file2.txt from dir2 directory")
		}
	})

	t.Run("line numbers", func(t *testing.T) {
		// Reset temp directory
		os.RemoveAll(tmpDir)
		os.MkdirAll(tmpDir, 0o755)

		// Create test files with multiple lines
		createFile("file1.txt", "Line 1\nLine 2\nLine 3")
		createFile("dir1/file2.txt", "First line\nSecond line")

		outputFile := filepath.Join(tmpDir, "output.txt")
		p, err := NewWithOptions(tmpDir, outputFile, "", SandwormOptions{PrintLineNumbers: true, FollowSymlinks: false})
		if err != nil {
			t.Fatalf("Failed to create processor: %v", err)
		}

		size, err := p.Process()
		if err != nil {
			t.Fatalf("Process failed: %v", err)
		}

		if size == 0 {
			t.Error("Expected non-zero file size")
		}

		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		output := string(content)

		// Check for expected content with line numbers
		if !strings.Contains(output, "PROJECT STRUCTURE:") {
			t.Error("Missing project structure section")
		}
		if !strings.Contains(output, "FILE CONTENTS:") {
			t.Error("Missing file contents section")
		}

		// Check that line numbers are present
		if !strings.Contains(output, "1: Line 1") {
			t.Error("Missing line number 1 for file1.txt")
		}
		if !strings.Contains(output, "2: Line 2") {
			t.Error("Missing line number 2 for file1.txt")
		}
		if !strings.Contains(output, "3: Line 3") {
			t.Error("Missing line number 3 for file1.txt")
		}
		if !strings.Contains(output, "1: First line") {
			t.Error("Missing line number 1 for file2.txt")
		}
		if !strings.Contains(output, "2: Second line") {
			t.Error("Missing line number 2 for file2.txt")
		}

		// Check that the line number format is correct (3 spaces + number + colon + space)
		lines := strings.Split(output, "\n")
		lineNumberPattern := "1: "
		foundLineNumber := false
		for _, line := range lines {
			if strings.Contains(line, lineNumberPattern) {
				foundLineNumber = true
				break
			}
		}
		if !foundLineNumber {
			t.Error("Line number format is incorrect")
		}
	})
}
