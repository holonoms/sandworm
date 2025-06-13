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
		p, err := New(tmpDir, outputFile, "")
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
		p, err := New(tmpDir, outputFile, filepath.Join(tmpDir, ".gitignore"))
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
		p, err := New(tmpDir, outputFile, "")
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
		p, err := New(tmpDir, outputFile, filepath.Join(tmpDir, "custom.ignore"))
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
		p, err := New(tmpDir, outputFile, "")
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
}
