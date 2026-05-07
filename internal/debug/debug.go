//go:build debug

package debug

import "log"

// Print logs only in debug build
// Excluded from release binary using build tags
func DebugLog(format string, args ...any) {
	log.Printf("[DEBUG] "+format, args...)
}
