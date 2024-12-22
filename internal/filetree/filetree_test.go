package filetree

import (
	"strings"
	"testing"
)

func TestFileTree(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		result := Build(nil, "")
		expected := "/"

		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("single level tree", func(t *testing.T) {
		paths := []string{"file1.txt", "file2.txt"}
		result := Build(paths, "")
		expected := strings.Join([]string{
			"/",
			"├── file1.txt",
			"└── file2.txt",
		}, "\n")

		if result != expected {
			t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
		}
	})

	t.Run("multi level tree", func(t *testing.T) {
		paths := []string{
			"dir1/file1.txt",
			"dir2/subdir/file2.txt",
			"file3.txt",
		}
		result := Build(paths, "")
		expected := strings.Join([]string{
			"/",
			"├── dir1/",
			"│   └── file1.txt",
			"├── dir2/",
			"│   └── subdir/",
			"│       └── file2.txt",
			"└── file3.txt",
		}, "\n")

		if result != expected {
			t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
		}
	})

	t.Run("custom root folder", func(t *testing.T) {
		paths := []string{
			"file1.txt",
			"dir/file2.txt",
		}
		result := Build(paths, "custom")
		expected := strings.Join([]string{
			"/custom",
			"├── dir/",
			"│   └── file2.txt",
			"└── file1.txt",
		}, "\n")

		if result != expected {
			t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
		}
	})
}
