// Package filetree provides functionality for creating ASCII tree representations
// of directory structures. It converts a list of file paths into a visual tree
// format that clearly shows the hierarchy and relationships between directories
// and files.
package filetree

import (
	"sort"
	"strings"
)

// Node represents a single node in the file tree structure. It's a map where
// the key is the name of the directory or file, and the value is either another
// Node (for directories) or nil (for files).
type Node map[string]any

// FileTree creates ASCII tree representations of directory structures.
// It maintains an internal map-based representation of the directory
// hierarchy that can be rendered into a string format.
type FileTree struct {
	root Node
}

// New creates a new FileTree instance and processes the provided paths
// into an internal tree structure. Each path is split into its components
// and added to the tree while maintaining the hierarchical relationships.
func New(paths []string) *FileTree {
	tree := &FileTree{
		root: make(Node),
	}

	for _, path := range paths {
		tree.addPath(strings.Split(path, "/"))
	}

	return tree
}

// String renders the file tree into a string representation using ASCII characters
// for the tree structure. It includes an optional custom root name and uses standard
// tree drawing characters (├──, └──, │) to show the hierarchy.
func (t *FileTree) String(customRoot string) string {
	// Start with root marker
	result := []string{
		"/" + customRoot,
	}

	t.buildTree(t.root, "", &result)
	return strings.Join(result, "\n")
}

// addPath adds a path to the internal tree structure by iterating through
// its components and creating the necessary nested maps.
func (t *FileTree) addPath(parts []string) {
	current := t.root

	for _, part := range parts {
		if part == "" {
			continue
		}

		if current[part] == nil {
			current[part] = make(Node)
		}

		current = current[part].(Node)
	}
}

// buildTree recursively builds the ASCII tree representation by traversing
// the internal tree structure and applying the appropriate prefixes and
// connectors based on the item's position in the hierarchy.
func (t *FileTree) buildTree(node Node, prefix string, result *[]string) {
	// Separate and sort directories and files
	var dirs, files []string
	for name, children := range node {
		if len(children.(Node)) > 0 {
			dirs = append(dirs, name)
		} else {
			files = append(files, name)
		}
	}
	sort.Strings(dirs)
	sort.Strings(files)

	// Combine sorted directories and files
	entries := dirs
	entries = append(entries, files...)

	for i, name := range entries {
		isLast := i == len(entries)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		// Add directory indicator for non-files
		displayName := name
		if i < len(dirs) {
			displayName += "/"
		}

		*result = append(*result, prefix+connector+displayName)

		if i < len(dirs) {
			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			t.buildTree(node[name].(Node), newPrefix, result)
		}
	}
}

// Build provides a convenient way to create and render a file tree in one step.
// It creates a new FileTree instance, processes the paths, and returns the
// string representation.
func Build(paths []string, customRoot string) string {
	return New(paths).String(customRoot)
}
