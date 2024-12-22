// Package util provides utility functions for common operations like
// formatting sizes and handling numerical conversions.
package util

import (
	"fmt"
	"math"
)

// FormatSize converts a byte count into a human-readable string using the most appropriate
// unit (B, KB, MB, GB, TB). The output is formatted with one decimal place for any size
// larger than bytes. For example:
//
//   - 0 bytes -> "0.0 B"
//   - 1024 bytes -> "1.0 KB"
//   - 1234567 bytes -> "1.2 MB"
//
// Particularly useful when displaying file sizes to users where exact byte counts are less
// important than quick comprehension.
func FormatSize(bytes int64) string {
	// Special case for zero to avoid math.Log calculations
	if bytes == 0 {
		return "0.0 B"
	}

	// Define our units and the size threshold (1024 bytes = 1 KB)
	units := []string{"B", "KB", "MB", "GB", "TB"}
	base := float64(1024)

	// Calculate the appropriate unit to use
	exp := int(math.Log(float64(bytes)) / math.Log(base))

	// Ensure we don't exceed our unit slice bounds
	if exp > len(units)-1 {
		exp = len(units) - 1
	}

	// Calculate the final value in the appropriate unit
	value := float64(bytes) / math.Pow(base, float64(exp))

	// Format with one decimal place, followed by the unit
	return fmt.Sprintf("%.1f %s", value, units[exp])
}
