package cli

import (
	"os"
	"testing"
)

func TestGenerateCmd_Flags(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandworm-cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	opts := &Options{}
	rootCmd := NewRootCmd(opts)
	rootCmd.SetArgs([]string{"generate", tmpDir, "--line-numbers", "--follow-symlinks"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	if opts.ShowLineNumbers == nil || !*opts.ShowLineNumbers {
		t.Errorf("Expected ShowLineNumbers to be true, got %v", opts.ShowLineNumbers)
	}
	if opts.FollowSymlinks == nil || !*opts.FollowSymlinks {
		t.Errorf("Expected FollowSymlinks to be true, got %v", opts.FollowSymlinks)
	}
	if opts.Directory != tmpDir {
		t.Errorf("Expected Directory to be '%v', got %v", tmpDir, opts.Directory)
	}

	// Clean up generated output file
	os.Remove("sandworm.txt")
}

func TestGenerateCmd_FlagsOverrideConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandworm-cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	opts := &Options{}
	rootCmd := NewRootCmd(opts)
	rootCmd.SetArgs([]string{"generate", tmpDir, "--line-numbers=false", "--follow-symlinks=false"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	if opts.ShowLineNumbers == nil || *opts.ShowLineNumbers {
		t.Errorf("Expected ShowLineNumbers to be false, got %v", opts.ShowLineNumbers)
	}
	if opts.FollowSymlinks == nil || *opts.FollowSymlinks {
		t.Errorf("Expected FollowSymlinks to be false, got %v", opts.FollowSymlinks)
	}

	// Clean up generated output file
	os.Remove("sandworm.txt")
}

func TestGenerateCmd_OutputIgnoreKeepFlags(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandworm-cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputFile := "myoutput.txt"
	ignoreFile := "myignore.txt"
	ignorePath := tmpDir + string(os.PathSeparator) + ignoreFile
	if err := os.WriteFile(ignorePath, []byte("*.tmp\n"), 0644); err != nil {
		t.Fatalf("Failed to create dummy ignore file: %v", err)
	}

	opts := &Options{}
	rootCmd := NewRootCmd(opts)
	rootCmd.SetArgs([]string{"generate", tmpDir, "--output", outputFile, "--ignore", ignorePath, "--keep"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	// Clean up generated output file
	os.Remove(outputFile)

	if opts.OutputFile != outputFile {
		t.Errorf("Expected OutputFile to be '%v', got '%v'", outputFile, opts.OutputFile)
	}
	if opts.IgnoreFile != ignorePath {
		t.Errorf("Expected IgnoreFile to be '%v', got '%v'", ignorePath, opts.IgnoreFile)
	}
	if !opts.KeepFile {
		t.Errorf("Expected KeepFile to be true, got %v", opts.KeepFile)
	}
}
