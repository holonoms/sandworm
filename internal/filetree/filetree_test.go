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

	t.Run("mixed path separators", func(t *testing.T) {
		// Test that FileTree can handle mixed Windows and Unix paths
		paths := []string{
			"file1.txt",
			"dir\\subdir\\file2.txt",  // Windows-style
			"dir/file3.txt",          // Unix-style
		}
		result := Build(paths, "")
		expected := strings.Join([]string{
			"/",
			"├── dir/",
			"│   ├── subdir/",
			"│   │   └── file2.txt",
			"│   └── file3.txt",
			"└── file1.txt",
		}, "\n")

		if result != expected {
			t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
		}
	})

	t.Run("edge cases with slashes", func(t *testing.T) {
		// Test edge cases: double slashes, leading/trailing slashes
		paths := []string{
			"normal/path.txt",
			"double//slash.txt",       // double slash
			"/leading/slash.txt",      // leading slash  
			"trailing/slash/.txt",     // trailing slash
			"multiple///slashes.txt",  // multiple slashes
		}
		result := Build(paths, "")
		expected := strings.Join([]string{
			"/",
			"├── double/",
			"│   └── slash.txt",
			"├── leading/",
			"│   └── slash.txt", 
			"├── multiple/",
			"│   └── slashes.txt",
			"├── normal/",
			"│   └── path.txt",
			"└── trailing/",
			"    └── slash/",
			"        └── .txt",
		}, "\n")

		if result != expected {
			t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
		}
	})
}
